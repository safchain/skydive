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

package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/abbot/go-http-auth"

	shttp "github.com/skydive-project/skydive/http"
	"github.com/skydive-project/skydive/logging"
	"github.com/skydive-project/skydive/packet_injector"
	"github.com/skydive-project/skydive/topology/graph"
	"github.com/skydive-project/skydive/topology/graph/traversal"
	"github.com/skydive-project/skydive/validator"
)

type PacketInjectorApi struct {
	Service  string
	PIClient *packet_injector.PacketInjectorClient
	Graph    *graph.Graph
}

type PacketParamsReq struct {
	Src     string `valid:"isGremlinExpr"`
	Dst     string `valid:"isGremlinExpr"`
	MsgType string
	Message string
}

func (pi *PacketInjectorApi) injectPacket(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	var pp packet_injector.PacketParams
	decoder := json.NewDecoder(r.Body)
	var ppr PacketParamsReq
	err := decoder.Decode(&ppr)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		logging.GetLogger().Errorf("Not able to decode the request body")
		return
	}
	defer r.Body.Close()

	if errs := validator.Validate(&ppr); errs != nil {
		writeError(w, http.StatusBadRequest, errs)
		logging.GetLogger().Errorf("Gremlin syntax error")
		return
	}

	srcNode := pi.getNode(ppr.Src)
	dstNode := pi.getNode(ppr.Dst)
	if srcNode == nil || dstNode == nil {
		writeError(w, http.StatusBadRequest, errors.New("Not able to find a Node"))
		logging.GetLogger().Errorf("Not able to find a Node")
		return
	}

	srcdata := srcNode.Metadata()
	dstdata := dstNode.Metadata()
	if srcdata["IPV4"] == "" || srcdata["MAC"] == "" ||
		dstdata["IPV4"] == "" || dstdata["MAC"] == "" {
		writeError(w, http.StatusBadRequest, errors.New("Selected nodes are not proper"))
		logging.GetLogger().Errorf("Selected nodes are not proper")
		return
	}

	pp.SourceNode = srcNode.ID
	pp.DestNode = dstNode.ID
	pp.Type = ppr.MsgType
	pp.Message = ppr.Message

	host := srcNode.Host()
	if !pi.PIClient.InjectPacket(host, &pp) {
		writeError(w, http.StatusBadRequest, errors.New("Somthing wrong to connect with agent"))
		logging.GetLogger().Errorf("Something wrong to connect with agent")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func (pi *PacketInjectorApi) getNode(gremlinQuery string) *graph.Node {
	pi.Graph.RLock()
	defer pi.Graph.RUnlock()

	tr := traversal.NewGremlinTraversalParser(strings.NewReader(gremlinQuery), pi.Graph)
	ts, err := tr.Parse()
	if err != nil {
		return nil
	}

	res, err := ts.Exec()
	if err != nil {
		return nil
	}

	for _, value := range res.Values() {
		switch value.(type) {
		case *graph.Node:
			return value.(*graph.Node)
		default:
			return nil
		}
	}
	return nil
}

func (pi *PacketInjectorApi) registerEndpoints(r *shttp.Server) {
	routes := []shttp.Route{
		{
			"InjectPacket",
			"POST",
			"/api/injectpacket",
			pi.injectPacket,
		},
	}

	r.RegisterRoutes(routes)
}

func RegisterPacketInjectorApi(s string, pic *packet_injector.PacketInjectorClient, g *graph.Graph, r *shttp.Server) {
	pia := &PacketInjectorApi{
		Service:  s,
		PIClient: pic,
		Graph:    g,
	}

	pia.registerEndpoints(r)
}
