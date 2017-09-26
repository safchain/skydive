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

package agent

import (
	"github.com/skydive-project/skydive/config"
	shttp "github.com/skydive-project/skydive/http"
	"github.com/skydive-project/skydive/logging"
	"github.com/skydive-project/skydive/topology/graph"
)

// TopologyForwarder forwards the topology to only one analyzer. Analyzers will forward
// messages between them in order to be synchronized. When switching from one analyzer to another one
// the agent will do a full re-sync because some messages could have been lost.
type TopologyForwarder struct {
	masterElection *shttp.WSMasterElection
	Graph          *graph.Graph
	Host           string
}

func (t *TopologyForwarder) triggerResync() {
	logging.GetLogger().Infof("Start a re-sync for %s", t.Host)

	t.Graph.RLock()
	defer t.Graph.RUnlock()

	// request for deletion of everything belonging to Root node
	t.masterElection.SendMessageToMaster(shttp.NewWSJSONMessage(graph.Namespace, graph.HostGraphDeletedMsgType, t.Host))

	// re-add all the nodes and edges
	t.masterElection.SendMessageToMaster(shttp.NewWSJSONMessage(graph.Namespace, graph.SyncMsgType, t.Graph))
}

// OnNewMaster websocket event handler
func (t *TopologyForwarder) OnNewMaster(c shttp.WSSpeaker) {
	if c == nil {
		logging.GetLogger().Warn("Lost connection to master")
	} else {
		addr, port := c.GetAddrPort()
		logging.GetLogger().Infof("Using %s:%d as master of topology forwarder", addr, port)
		t.triggerResync()
	}
}

// OnNodeUpdated websocket event handler
func (t *TopologyForwarder) OnNodeUpdated(n *graph.Node) {
	t.masterElection.SendMessageToMaster(shttp.NewWSJSONMessage(graph.Namespace, graph.NodeUpdatedMsgType, n))
}

// OnNodeAdded websocket event handler
func (t *TopologyForwarder) OnNodeAdded(n *graph.Node) {
	t.masterElection.SendMessageToMaster(shttp.NewWSJSONMessage(graph.Namespace, graph.NodeAddedMsgType, n))
}

// OnNodeDeleted websocket event handler
func (t *TopologyForwarder) OnNodeDeleted(n *graph.Node) {
	t.masterElection.SendMessageToMaster(shttp.NewWSJSONMessage(graph.Namespace, graph.NodeDeletedMsgType, n))
}

// OnEdgeUpdated websocket event handler
func (t *TopologyForwarder) OnEdgeUpdated(e *graph.Edge) {
	t.masterElection.SendMessageToMaster(shttp.NewWSJSONMessage(graph.Namespace, graph.EdgeUpdatedMsgType, e))
}

// OnEdgeAdded websocket event handler
func (t *TopologyForwarder) OnEdgeAdded(e *graph.Edge) {
	t.masterElection.SendMessageToMaster(shttp.NewWSJSONMessage(graph.Namespace, graph.EdgeAddedMsgType, e))
}

// OnEdgeDeleted websocket event handler
func (t *TopologyForwarder) OnEdgeDeleted(e *graph.Edge) {
	t.masterElection.SendMessageToMaster(shttp.NewWSJSONMessage(graph.Namespace, graph.EdgeDeletedMsgType, e))
}

// NewTopologyForwarder is a mechanism aiming to distribute all graph node notifications to a WebSocket client pool
func NewTopologyForwarder(host string, g *graph.Graph, pool shttp.WSSpeakerPool) *TopologyForwarder {
	masterElection := shttp.NewWSMasterElection(pool)

	t := &TopologyForwarder{
		masterElection: masterElection,
		Graph:          g,
		Host:           host,
	}

	masterElection.AddEventHandler(t)
	g.AddEventListener(t)

	return t
}

// NewTopologyForwarderFromConfig creates a TopologyForwarder from configuration
func NewTopologyForwarderFromConfig(g *graph.Graph, pool shttp.WSSpeakerPool) *TopologyForwarder {
	host := config.GetConfig().GetString("host_id")
	return NewTopologyForwarder(host, g, pool)
}
