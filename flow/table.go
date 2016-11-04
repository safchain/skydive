/*
 * Copyright (C) 2016 Red Hat, Inc.
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

package flow

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/skydive-project/skydive/common"
	"github.com/skydive-project/skydive/logging"
)

// TableQuery contains a type and a query obj as an array of bytes.
// The query can be encoded in different ways according the type.
type TableQuery struct {
	Type string
	Obj  []byte
}

// TableReply is the response to a TableQuery containing a Status and an array
// of replies that can be encoded in many ways, ex: json, protobuf.
type TableReply struct {
	Status int
	Obj    [][]byte
}

type ExpireUpdateFunc func(f []*Flow)

type FlowHandler struct {
	callback ExpireUpdateFunc
	every    time.Duration
}

func NewFlowHandler(callback ExpireUpdateFunc, every time.Duration) *FlowHandler {
	return &FlowHandler{
		callback: callback,
		every:    every,
	}
}

type Table struct {
	sync.RWMutex
	table         map[string]*Flow
	stats         map[string]*FlowMetric
	defaultFunc   func()
	flush         chan bool
	flushDone     chan bool
	query         chan *TableQuery
	reply         chan *TableReply
	state         int64
	lockState     sync.RWMutex
	wg            sync.WaitGroup
	updateHandler *FlowHandler
	expireHandler *FlowHandler
	tableClock    int64
}

func NewTable(updateHandler *FlowHandler, expireHandler *FlowHandler) *Table {
	t := &Table{
		table:         make(map[string]*Flow),
		stats:         make(map[string]*FlowMetric),
		flush:         make(chan bool),
		flushDone:     make(chan bool),
		state:         common.StoppedState,
		updateHandler: updateHandler,
		expireHandler: expireHandler,
	}
	atomic.StoreInt64(&t.tableClock, time.Now().Unix())
	return t
}

func NewTableFromFlows(flows []*Flow, updateHandler *FlowHandler, expireHandler *FlowHandler) *Table {
	nft := NewTable(updateHandler, expireHandler)
	nft.Update(flows)
	return nft
}

func (ft *Table) String() string {
	ft.RLock()
	defer ft.RUnlock()
	return fmt.Sprintf("%d flows", len(ft.table))
}

func (ft *Table) Update(flows []*Flow) {
	ft.Lock()
	for _, f := range flows {
		if _, ok := ft.table[f.UUID]; !ok {
			ft.table[f.UUID] = f
		} else {
			ft.table[f.UUID].Metric = f.Metric
		}
	}
	ft.Unlock()
}

func (ft *Table) GetTime() int64 {
	return atomic.LoadInt64(&ft.tableClock)
}

func (ft *Table) GetFlows(query *FlowSearchQuery) *FlowSet {
	ft.RLock()
	defer ft.RUnlock()

	var it *common.Iterator
	if query != nil && query.Range != nil {
		it = common.NewIterator(0, query.Range.From, query.Range.To)
	} else {
		it = common.NewIterator()
	}

	flowset := NewFlowSet()
	for _, f := range ft.table {
		if it.Done() {
			break
		}
		if (query == nil || query.Filter == nil || query.Filter.Eval(f)) && it.Next() {
			if flowset.Start == 0 || flowset.Start > f.Metric.Start {
				flowset.Start = f.Metric.Start
			}
			if flowset.End == 0 || flowset.Start < f.Metric.Last {
				flowset.End = f.Metric.Last
			}
			flowset.Flows = append(flowset.Flows, f)
		}
	}
	return flowset
}

func (ft *Table) GetFlow(key string) *Flow {
	ft.RLock()
	defer ft.RUnlock()
	if flow, found := ft.table[key]; found {
		return flow
	}

	return nil
}

func (ft *Table) GetOrCreateFlow(key string) (*Flow, bool) {
	ft.Lock()
	defer ft.Unlock()
	if flow, found := ft.table[key]; found {
		return flow, false
	}

	new := &Flow{
		Metric:           &FlowMetric{},
		LastUpdateMetric: &FlowMetric{},
	}
	ft.table[key] = new

	return new, true
}

/* Return a new flow.Table that contain <last> active flows */
func (ft *Table) FilterLast(last time.Duration) []*Flow {
	var flows []*Flow
	selected := time.Now().Unix() - int64((last).Seconds())
	ft.RLock()
	defer ft.RUnlock()
	for _, f := range ft.table {
		if f.Metric.Last >= selected {
			flows = append(flows, f)
		}
	}
	return flows
}

func (ft *Table) SelectLayer(protocol FlowProtocol, list []string) *FlowSet {
	meth := make(map[string][]*Flow)
	ft.RLock()
	for _, f := range ft.table {
		layerFlow := f.Link
		if layerFlow == nil || layerFlow.A == "ff:ff:ff:ff:ff:ff" || layerFlow.B == "ff:ff:ff:ff:ff:ff" {
			continue
		}
		meth[layerFlow.A] = append(meth[layerFlow.A], f)
		meth[layerFlow.B] = append(meth[layerFlow.B], f)
	}
	ft.RUnlock()

	mflows := make(map[*Flow]struct{})
	flowset := NewFlowSet()
	for _, eth := range list {
		if flist, ok := meth[eth]; ok {
			for _, f := range flist {
				if _, found := mflows[f]; !found {
					mflows[f] = struct{}{}

					if flowset.Start == 0 || flowset.Start > f.Metric.Start {
						flowset.Start = f.Metric.Start
					}
					if flowset.End == 0 || flowset.Start < f.Metric.Last {
						flowset.End = f.Metric.Last
					}

					flowset.Flows = append(flowset.Flows, f)
				}
			}
		}
	}
	return flowset
}

/* Internal call only, Must be called under ft.Lock() */
func (ft *Table) expired(expireBefore int64) {
	var expiredFlows []*Flow
	flowTableSzBefore := len(ft.table)
	for k, f := range ft.table {
		if f.Metric.Last < expireBefore {
			duration := time.Duration(f.Metric.Last - f.Metric.Start)
			logging.GetLogger().Debugf("Expire flow %s Duration %v", f.UUID, duration)
			expiredFlows = append(expiredFlows, f)

			// need to use the key as the key could be not equal to the UUID
			delete(ft.table, k)

			// stats are always indexed by UUID
			delete(ft.stats, f.UUID)
		}
	}
	/* Advise Clients */
	if ft.expireHandler.callback != nil {
		ft.expireHandler.callback(expiredFlows)
	}

	flowTableSz := len(ft.table)
	logging.GetLogger().Debugf("Expire Flow : removed %v ; new size %v", flowTableSzBefore-flowTableSz, flowTableSz)
}

func (ft *Table) Updated(now time.Time) {
	timepoint := now.Unix() - int64((ft.updateHandler.every).Seconds())
	ft.RLock()
	ft.updated(timepoint)
	ft.RUnlock()
}

/* Internal call only, Must be called under ft.RLock() */
func (ft *Table) updated(updateFrom int64) {
	every := int64(ft.updateHandler.every.Seconds())

	var updatedFlows []*Flow
	for _, f := range ft.table {
		if f.Metric.Last > updateFrom {
			updatedFlows = append(updatedFlows, f)

			e := f.Metric
			f.LastUpdateMetric.ABPackets = e.ABPackets
			f.LastUpdateMetric.ABBytes = e.ABBytes
			f.LastUpdateMetric.BAPackets = e.BAPackets
			f.LastUpdateMetric.BABytes = e.BABytes

			f.LastUpdateMetric.Start = updateFrom
			f.LastUpdateMetric.Last = updateFrom + every

			// substract previous values to get the diff so that we store the
			// amount of data between two updates
			if s, ok := ft.stats[f.UUID]; ok {
				f.LastUpdateMetric.ABPackets -= s.ABPackets
				f.LastUpdateMetric.ABBytes -= s.ABBytes
				f.LastUpdateMetric.BAPackets -= s.BAPackets
				f.LastUpdateMetric.BABytes -= s.BABytes
			}
		} else {
			f.LastUpdateMetric = &FlowMetric{}
		}

		ft.stats[f.UUID] = f.Metric.Copy()
	}

	/* Advise Clients */
	if ft.updateHandler.callback != nil {
		ft.updateHandler.callback(updatedFlows)
	}
	logging.GetLogger().Debugf("Send updated Flow %d", len(updatedFlows))
}

func (ft *Table) expireNow() {
	const Now = int64(^uint64(0) >> 1)
	ft.Lock()
	ft.expired(Now)
	ft.Unlock()
}

func (ft *Table) Expire(now time.Time) {
	timepoint := now.Unix() - int64((ft.expireHandler.every).Seconds())
	ft.Lock()
	ft.expired(timepoint)
	ft.Unlock()
}

func (ft *Table) RegisterDefault(fn func()) {
	ft.Lock()
	ft.defaultFunc = fn
	ft.Unlock()
}

/* Window returns a FlowSet with flows fitting in the given time range
Need to Rlock the table before calling. Returned flows may not be unique */
func (ft *Table) Window(start, end int64) *FlowSet {
	flowset := NewFlowSet()
	flowset.Start = start
	flowset.End = end

	if end >= start {
		filter := &Filter{
			BoolFilter: &BoolFilter{
				Op: BoolFilterOp_AND,
				Filters: []*Filter{
					{
						LtInt64Filter: &LtInt64Filter{
							Key:   "Metric.Start",
							Value: end,
						},
					},
					{
						GteInt64Filter: &GteInt64Filter{
							Key:   "Metric.Last",
							Value: start,
						},
					},
				},
			},
		}

		set := ft.GetFlows(&FlowSearchQuery{Filter: filter})
		flowset.Flows = set.Flows
	}

	return flowset
}

func (ft *Table) Flush() {
	ft.flush <- true
	<-ft.flushDone
}

func (ft *Table) onFlowSearchQueryMessage(fsq *FlowSearchQuery) (*FlowSearchReply, int) {
	flowset := ft.GetFlows(fsq)
	if len(flowset.Flows) == 0 {
		return &FlowSearchReply{
			FlowSet: flowset,
		}, http.StatusNoContent
	}

	if fsq.Sort {
		flowset.Sort()
	}

	return &FlowSearchReply{
		FlowSet: flowset,
	}, http.StatusOK
}

func (ft *Table) onQuery(query *TableQuery) *TableReply {
	reply := &TableReply{
		Status: http.StatusBadRequest,
		Obj:    make([][]byte, 0),
	}

	switch query.Type {
	case "FlowSearchQuery":
		var fsq FlowSearchQuery
		if err := proto.Unmarshal(query.Obj, &fsq); err != nil {
			logging.GetLogger().Errorf("Unable to decode the flow search query: %s", err.Error())
			break
		}

		fsr, status := ft.onFlowSearchQueryMessage(&fsq)
		if status != http.StatusOK {
			reply.Status = status
			break
		}
		pb, _ := proto.Marshal(fsr)

		// TableReply returns an array of replies so in that case an array of
		// protobuf replies
		reply.Obj = append(reply.Obj, pb)

		reply.Status = http.StatusOK
	}

	return reply
}

func (ft *Table) Query(query *TableQuery) *TableReply {
	ft.lockState.Lock()
	defer ft.lockState.Unlock()

	if atomic.LoadInt64(&ft.state) == common.RunningState {
		ft.query <- query

		timer := time.NewTicker(1 * time.Second)
		defer timer.Stop()

		for {
			select {
			case r := <-ft.reply:
				return r
			case <-timer.C:
				if atomic.LoadInt64(&ft.state) != common.RunningState {
					return nil
				}
			}
		}
	}
	return nil
}

func (ft *Table) Start() {
	ft.wg.Add(1)
	defer ft.wg.Done()

	updateTicker := time.NewTicker(ft.updateHandler.every)
	defer updateTicker.Stop()

	expireTicker := time.NewTicker(ft.expireHandler.every)
	defer expireTicker.Stop()

	nowTicker := time.NewTicker(time.Second * 1)
	defer nowTicker.Stop()

	ft.query = make(chan *TableQuery, 100)
	ft.reply = make(chan *TableReply, 100)

	atomic.StoreInt64(&ft.state, common.RunningState)
	for atomic.LoadInt64(&ft.state) == common.RunningState {
		select {
		case now := <-expireTicker.C:
			ft.Expire(now)
		case now := <-updateTicker.C:
			ft.Updated(now)
		case <-ft.flush:
			ft.expireNow()
			ft.flushDone <- true
		case query, ok := <-ft.query:
			if ok {
				ft.reply <- ft.onQuery(query)
			}
		case now := <-nowTicker.C:
			atomic.StoreInt64(&ft.tableClock, now.Unix())
		default:
			if ft.defaultFunc != nil {
				ft.defaultFunc()
			} else {
				time.Sleep(20 * time.Millisecond)
			}
		}
	}
}

func (ft *Table) Stop() {
	if atomic.CompareAndSwapInt64(&ft.state, common.RunningState, common.StoppingState) {
		ft.wg.Wait()

		ft.lockState.Lock()
		close(ft.query)
		close(ft.reply)
		ft.lockState.Unlock()
	}

	// FIX trigger deadlock since Stop is called from a place where the graph
	// is locked and usually the enhance pipeline lock the graph as well.
	// expireNow calls the enhance.
	//ft.expireNow()
}
