/*
 * Copyright (C) 2018 Red Hat, Inc.
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

package stub

import (
	"sync/atomic"
	"time"

	"github.com/skydive-project/skydive/common"
	"github.com/skydive-project/skydive/logging"
	"github.com/skydive-project/skydive/topology/graph"
)

// Probe ...
type Probe struct {
	g     *graph.Graph
	state int64
}

func (p *Probe) run() {
	atomic.StoreInt64(&p.state, common.RunningState)

	for atomic.LoadInt64(&p.state) == common.RunningState {
		logging.GetLogger().Debugf("Stub probe process...")

		// REST CALL...
		p.g.Lock()

		// creates two nodes
		n1 := p.g.NewNode(graph.GenID("NODE1"), graph.Metadata{"Name": "node1", "Type": "vm"})
		n2 := p.g.NewNode(graph.GenID("NODE2"), graph.Metadata{"Name": "node2", "Type": "vm"})

		// link them
		p.g.NewEdge(graph.GenID("NODE1/NODE2"), n1, n2, graph.Metadata{"Type": "l2vpn"})

		p.g.Unlock()

		time.Sleep(1 * time.Second)
	}
}

// Start ...
func (p *Probe) Start() {
	go p.run()
}

// Stop ....
func (p *Probe) Stop() {
	atomic.CompareAndSwapInt64(&p.state, common.RunningState, common.StoppingState)
}

// NewProbe ...
func NewProbe(g *graph.Graph) (*Probe, error) {
	probe := &Probe{
		g: g,
	}
	atomic.StoreInt64(&probe.state, common.StoppedState)

	return probe, nil
}
