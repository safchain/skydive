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
	"net/http"
	"sync"

	"github.com/skydive-project/skydive/common"
	shttp "github.com/skydive-project/skydive/http"
	"github.com/skydive-project/skydive/logging"
	"github.com/skydive-project/skydive/statics"
	"github.com/skydive-project/skydive/topology/graph"
	"github.com/xeipuuv/gojsonschema"
)

// TopologyServer describes a service to reply to topology queries
type TopologyServer struct {
	sync.RWMutex
	shttp.DefaultWSClientEventHandler
	WSServer   *shttp.WSMessageServer
	Graph      *graph.Graph
	cached     *graph.CachedBackend
	nodeSchema gojsonschema.JSONLoader
	edgeSchema gojsonschema.JSONLoader
	// map used to store agent which uses this analyzer as master
	// basically sending graph messages
	authors map[string]bool
}

func (t *TopologyServer) hostGraphDeleted(host string, mode int) {
	t.cached.SetMode(mode)
	defer t.cached.SetMode(graph.DefaultMode)

	t.Graph.DelHostGraph(host)
}

// OnDisconnected websocket event
func (t *TopologyServer) OnDisconnected(c shttp.WSClient) {
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
func (t *TopologyServer) OnWSMessage(c shttp.WSClient, msg shttp.WSMessage) {
	msgType, obj, err := graph.UnmarshalWSMessage(msg)
	if err != nil {
		logging.GetLogger().Errorf("Graph: Unable to parse the event %v: %s", msg, err.Error())
		return
	}

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

	for _, client := range t.WSServer.GetClients() {
		if client.GetClientType() != common.AgentService {
			client.Send(msg)
		}
	}
}

// NewTopologyServer creates a new topology server
func NewTopologyServer(server *shttp.WSMessageServer) (*TopologyServer, error) {
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

	t := &TopologyServer{
		Graph:      g,
		WSServer:   server,
		cached:     cached,
		authors:    make(map[string]bool),
		nodeSchema: gojsonschema.NewBytesLoader(nodeSchema),
		edgeSchema: gojsonschema.NewBytesLoader(edgeSchema),
	}
	server.AddEventHandler(t)
	server.AddMessageHandler(t, []string{graph.Namespace})

	return t, nil
}
