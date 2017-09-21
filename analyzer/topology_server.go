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

package analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/skydive-project/skydive/api"
	"github.com/skydive-project/skydive/common"
	"github.com/skydive-project/skydive/config"
	shttp "github.com/skydive-project/skydive/http"
	"github.com/skydive-project/skydive/logging"
	"github.com/skydive-project/skydive/statics"
	"github.com/skydive-project/skydive/topology/graph"
	"github.com/xeipuuv/gojsonschema"
)

// TopologyReplicatorPeer describes a topology forwarder peer
type TopologyReplicatorPeer struct {
	shttp.DefaultWSSpeakerEventHandler
	Addr        string
	Port        int
	Graph       *graph.Graph
	AuthOptions *shttp.AuthenticationOpts
	wsclient    *shttp.WSClient
	host        string
}

// TopologyServer describes a service to reply to topology queries
type TopologyServer struct {
	sync.RWMutex
	shttp.DefaultWSSpeakerEventHandler
	WSServer   *shttp.WSJSONMessageServer
	Graph      *graph.Graph
	cached     *graph.CachedBackend
	nodeSchema gojsonschema.JSONLoader
	edgeSchema gojsonschema.JSONLoader
	// map used to store agent which uses this analyzer as master
	// basically sending graph messages
	authors      map[string]bool
	peers        []*TopologyReplicatorPeer
	wg           sync.WaitGroup
	replicateMsg atomic.Value
}

func (p *TopologyReplicatorPeer) getHostID() string {
	client := shttp.NewRestClient(p.Addr, p.Port, p.AuthOptions)
	contentReader := bytes.NewReader([]byte{})

	var data []byte
	var info api.Info

	for {
		resp, err := client.Request("GET", "api", contentReader, nil)
		if err != nil {
			goto NotReady
		}

		if resp.StatusCode != http.StatusOK {
			goto NotReady
		}

		data, _ = ioutil.ReadAll(resp.Body)
		if len(data) == 0 {
			goto NotReady
		}

		if err := json.Unmarshal(data, &info); err != nil {
			goto NotReady
		}
		p.host = info.Host

		return p.host

	NotReady:
		time.Sleep(1 * time.Second)
	}
}

// OnConnected send the whole local graph the remote peer(analyzer) once connected
func (p *TopologyReplicatorPeer) OnConnected(c shttp.WSSpeaker) {
	p.Graph.RLock()
	defer p.Graph.RUnlock()

	logging.GetLogger().Infof("Send the whole graph to: %s", p.Graph.GetHost())

	// re-added all the nodes and edges
	p.wsclient.Send(shttp.NewWSJSONMessage(graph.Namespace, graph.SyncMsgType, p.Graph.GetHost()))
}

func (p *TopologyReplicatorPeer) connect(wg *sync.WaitGroup) {
	defer wg.Done()

	// check whether the peer is the analyzer itself or not thanks to the /api
	if p.getHostID() == config.GetConfig().GetString("host_id") {
		logging.GetLogger().Debugf("No connection to %s:%d as it's me", p.Addr, p.Port)
		return
	}

	authClient := shttp.NewAuthenticationClient(p.Addr, p.Port, p.AuthOptions)
	p.wsclient = shttp.NewWSClientFromConfig(common.AnalyzerService, p.Addr, p.Port, "/ws", authClient)
	p.wsclient.AddEventHandler(p)

	p.wsclient.Connect()
}

func (p *TopologyReplicatorPeer) disconnect() {
	if p.wsclient != nil {
		p.wsclient.Disconnect()
	}
}

func (t *TopologyServer) addPeer(addr string, port int, auth *shttp.AuthenticationOpts, g *graph.Graph) {
	peer := &TopologyReplicatorPeer{
		Addr:        addr,
		Port:        port,
		Graph:       g,
		AuthOptions: auth,
	}

	t.peers = append(t.peers, peer)
}

// ConnectAll peers
func (t *TopologyServer) ConnectPeers() {
	for _, peer := range t.peers {
		t.wg.Add(1)
		go peer.connect(&t.wg)
	}
}

// DisconnectAll peers
func (t *TopologyServer) DisconnectPeers() {
	for _, peer := range t.peers {
		peer.disconnect()
	}
	t.wg.Wait()
}

func (t *TopologyServer) hostGraphDeleted(host string, mode int) {
	t.cached.SetMode(mode)
	defer t.cached.SetMode(graph.DefaultMode)

	t.Graph.DelHostGraph(host)
}

// OnDisconnected websocket event
func (t *TopologyServer) OnDisconnected(c shttp.WSSpeaker) {
	host := c.GetHost()

	t.RLock()
	_, ok := t.authors[host]
	t.RUnlock()

	// not an author so do not delete resources
	if !ok {
		return
	}

	t.Graph.Lock()
	defer t.Graph.Unlock()

	// it's an authors so already received a message meaning that the client chose this analyzer as master
	logging.GetLogger().Debugf("Authoritative client unregistered, delete resources %s", host)
	t.hostGraphDeleted(host, graph.DefaultMode)

	t.Lock()
	delete(t.authors, host)
	t.Unlock()
}

// OnWSMessage websocket event
func (t *TopologyServer) OnWSJSONMessage(c shttp.WSSpeaker, msg shttp.WSJSONMessage) {
	msgType, obj, err := graph.UnmarshalWSMessage(msg)
	if err != nil {
		logging.GetLogger().Errorf("Graph: Unable to parse the event %v: %s", msg, err.Error())
		return
	}

	// this kind of message usually comes from external clients like the WebUI
	if msgType == graph.SyncRequestMsgType {
		t.Graph.RLock()
		context, status := obj.(graph.GraphContext), http.StatusOK
		g, err := t.Graph.WithContext(context)
		if err != nil {
			logging.GetLogger().Errorf("analyzer is unable to get a graph with context %+v: %s", context, err.Error())
			g, status = nil, http.StatusBadRequest
		}
		reply := msg.Reply(g, graph.SyncReplyMsgType, status)
		c.Send(reply)
		t.Graph.RUnlock()

		return
	}

	clientType := c.GetClientType()

	// author if message coming from another client than analyzer
	if clientType != "" && clientType != common.AnalyzerService {
		t.Lock()
		t.authors[c.GetHost()] = true
		t.Unlock()
	} else {
		t.replicateMsg.Store(false)
		defer t.replicateMsg.Store(true)
	}

	if clientType != common.AnalyzerService && clientType != common.AgentService {
		loader := gojsonschema.NewGoLoader(obj)

		var schema gojsonschema.JSONLoader
		switch msgType {
		case graph.NodeAddedMsgType, graph.NodeUpdatedMsgType, graph.NodeDeletedMsgType:
			schema = t.nodeSchema
		case graph.EdgeAddedMsgType, graph.EdgeUpdatedMsgType, graph.EdgeDeletedMsgType:
			schema = t.edgeSchema
		}

		if schema != nil {
			if _, err := gojsonschema.Validate(t.edgeSchema, loader); err != nil {
				logging.GetLogger().Errorf("Invalid message: %s", err.Error())
				return
			}
		}
	}

	t.Graph.Lock()
	defer t.Graph.Unlock()

	// got HostGraphDeleted, so if not an analyzer we need to do two things:
	// force the deletion from the cache and force the delete from the persistent
	// backend. We need to use the persistent only to be use to retrieve nodes/edges
	// from the persistent backend otherwise the memory backend would be used.
	if msgType == graph.HostGraphDeletedMsgType {
		logging.GetLogger().Debugf("Got %s message for host %s", graph.HostGraphDeletedMsgType, obj.(string))

		t.hostGraphDeleted(obj.(string), graph.CacheOnlyMode)
		if clientType != common.AnalyzerService {
			t.hostGraphDeleted(obj.(string), graph.PersistentOnlyMode)
		}

		return
	}

	// If the message comes from analyzer we need to apply it only on cache only
	// as it is a forwarded message.
	if clientType == common.AnalyzerService {
		t.cached.SetMode(graph.CacheOnlyMode)
	}
	defer t.cached.SetMode(graph.DefaultMode)

	switch msgType {
	case graph.SyncMsgType, graph.SyncReplyMsgType:
		r := obj.(*graph.SyncMsg)
		for _, n := range r.Nodes {
			if t.Graph.GetNode(n.ID) == nil {
				t.Graph.NodeAdded(n)
			}
		}
		for _, e := range r.Edges {
			if t.Graph.GetEdge(e.ID) == nil {
				t.Graph.EdgeAdded(e)
			}
		}
	case graph.NodeUpdatedMsgType:
		t.Graph.NodeUpdated(obj.(*graph.Node))
	case graph.NodeDeletedMsgType:
		t.Graph.NodeDeleted(obj.(*graph.Node))
	case graph.NodeAddedMsgType:
		t.Graph.NodeAdded(obj.(*graph.Node))
	case graph.EdgeUpdatedMsgType:
		t.Graph.EdgeUpdated(obj.(*graph.Edge))
	case graph.EdgeDeletedMsgType:
		t.Graph.EdgeDeleted(obj.(*graph.Edge))
	case graph.EdgeAddedMsgType:
		t.Graph.EdgeAdded(obj.(*graph.Edge))
	}
}

// notifyClients aims to forward local graph modification to external clients
// the goal here is not to handle analyzer replication.
func (t *TopologyServer) notifyClients(msg *shttp.WSJSONMessage) {
	for _, c := range t.WSServer.GetClients() {
		clientType := c.GetClientType()
		if clientType != common.AnalyzerService && clientType != common.AgentService {
			c.Send(msg)
		}
	}
}

// notifyClients aims to forward local graph modification to external clients
// the goal here is not to handle analyzer replication.
func (t *TopologyServer) notifyPeers(msg *shttp.WSJSONMessage) {
	for _, p := range t.peers {
		if p.wsclient != nil {
			p.wsclient.Send(msg)
		}
	}
}

// OnNodeUpdated websocket event handler
func (t *TopologyServer) OnNodeUpdated(n *graph.Node) {
	msg := shttp.NewWSJSONMessage(graph.Namespace, graph.NodeUpdatedMsgType, n)
	t.notifyClients(msg)
	if t.replicateMsg.Load() == true {
		t.notifyPeers(msg)
	}
}

// OnNodeAdded websocket event handler
func (t *TopologyServer) OnNodeAdded(n *graph.Node) {
	msg := shttp.NewWSJSONMessage(graph.Namespace, graph.NodeAddedMsgType, n)
	t.notifyClients(msg)
	if t.replicateMsg.Load() == true {
		t.notifyPeers(msg)
	}
}

// OnNodeDeleted websocket event handler
func (t *TopologyServer) OnNodeDeleted(n *graph.Node) {
	msg := shttp.NewWSJSONMessage(graph.Namespace, graph.NodeDeletedMsgType, n)
	t.notifyClients(msg)
	if t.replicateMsg.Load() == true {
		t.notifyPeers(msg)
	}
}

// OnEdgeUpdated websocket event handler
func (t *TopologyServer) OnEdgeUpdated(e *graph.Edge) {
	msg := shttp.NewWSJSONMessage(graph.Namespace, graph.EdgeUpdatedMsgType, e)
	t.notifyClients(msg)
	if t.replicateMsg.Load() == true {
		t.notifyPeers(msg)
	}
}

// OnEdgeAdded websocket event handler
func (t *TopologyServer) OnEdgeAdded(e *graph.Edge) {
	msg := shttp.NewWSJSONMessage(graph.Namespace, graph.EdgeAddedMsgType, e)
	t.notifyClients(msg)
	if t.replicateMsg.Load() == true {
		t.notifyPeers(msg)
	}
}

// OnEdgeDeleted websocket event handler
func (t *TopologyServer) OnEdgeDeleted(e *graph.Edge) {
	msg := shttp.NewWSJSONMessage(graph.Namespace, graph.EdgeDeletedMsgType, e)
	t.notifyClients(msg)
	if t.replicateMsg.Load() == true {
		t.notifyPeers(msg)
	}
}

// NewTopologyServer creates a new topology server
func NewTopologyServer(server *shttp.WSJSONMessageServer) (*TopologyServer, error) {
	persistent, err := graph.BackendFromConfig()
	if err != nil {
		return nil, err
	}

	cached, err := graph.NewCachedBackend(persistent)
	if err != nil {
		return nil, err
	}

	g := graph.NewGraphFromConfig(cached)

	nodeSchema, err := statics.Asset("statics/schemas/node.schema")
	if err != nil {
		return nil, err
	}

	edgeSchema, err := statics.Asset("statics/schemas/edge.schema")
	if err != nil {
		return nil, err
	}

	addresses, err := config.GetAnalyzerServiceAddresses()
	if err != nil {
		return nil, fmt.Errorf("Unable to get the analyzers list: %s", err)
	}

	t := &TopologyServer{
		Graph:      g,
		WSServer:   server,
		cached:     cached,
		authors:    make(map[string]bool),
		nodeSchema: gojsonschema.NewBytesLoader(nodeSchema),
		edgeSchema: gojsonschema.NewBytesLoader(edgeSchema),
	}
	t.replicateMsg.Store(true)

	server.AddEventHandler(t)
	server.AddJSONMessageHandler(t, []string{graph.Namespace})
	g.AddEventListener(t)

	authOptions := &shttp.AuthenticationOpts{
		Username: config.GetConfig().GetString("auth.analyzer_username"),
		Password: config.GetConfig().GetString("auth.analyzer_password"),
	}

	for _, sa := range addresses {
		t.addPeer(sa.Addr, sa.Port, authOptions, g)
	}

	return t, nil
}
