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

package probes

import (
	"net"
	"strings"

	"github.com/skydive-project/skydive/logging"
	"github.com/skydive-project/skydive/topology/graph"
)

type Layer3Probe struct {
	graph.DefaultGraphListener
	graph *graph.Graph
}

func (l3 *Layer3Probe) networkIntersec(ipv4A []string, ipv4B []string) []*net.IPNet {
	intersec := []*net.IPNet{}

	for _, ipA := range ipv4A {
		_, cidrA, err := net.ParseCIDR(ipA)
		if err != nil {
			logging.GetLogger().Errorf("Error while parsing IPV4 node address %s: %s", ipA, err.Error())
			continue
		}

		for _, ipB := range ipv4B {
			_, cidrB, err := net.ParseCIDR(ipB)
			if err != nil {
				logging.GetLogger().Errorf("Error while parsing IPV4 node address %s: %s", ipB, err.Error())
				continue
			}

			if cidrA.Network() == cidrB.Network() {
				intersec = append(intersec, cidrA)
			}
		}
	}

	return intersec
}

func (l3 *Layer3Probe) onNodeEvent(n *graph.Node) {
	if ipv4A, ok := n.Metadata()["IPV4"]; ok {
		for _, o := range l3.graph.GetNodes() {
			if n == o {
				continue
			}

			if ipv4B, ok := o.Metadata()["IPV4"]; ok {
				// check vlan first
				if n.Metadata()["Vlan"] != n.Metadata()["Vlan"] {
					continue
				}

				// check wether there is a layer2 path between the 2 nodes
				if sp := l3.graph.LookupShortestPath(n, o.Metadata(), graph.Metadata{"RelationType": "layer2"}); len(sp) == 0 {
					continue
				}

				a := strings.Split(ipv4A.(string), ",")
				b := strings.Split(ipv4B.(string), ",")

				// TODO need to be improved maybe with a radix tree
				for _, ipn := range l3.networkIntersec(a, b) {
					m := graph.Metadata{"RelationType": "layer3", "CIDR": ipn.Network()}
					if !l3.graph.AreLinked(n, o, m) {
						l3.graph.Link(n, o, m)
					}
				}
			}
		}
	}
}

func (l3 *Layer3Probe) OnNodeUpdated(n *graph.Node) {
	l3.onNodeEvent(n)
}

func (l3 *Layer3Probe) OnNodeAdded(n *graph.Node) {
	l3.onNodeEvent(n)
}

func (l3 *Layer3Probe) OnEdgeAdded(e *graph.Edge) {
	parent, child := l3.graph.GetEdgeNodes(e)
	if parent == nil || child == nil || e.Metadata()["RelationType"] != "layer2" {
		return
	}

	l3.onNodeEvent(child)
}

func (l3 *Layer3Probe) Start() {
}

func (l3 *Layer3Probe) Stop() {
}

func NewLayer3Probe(g *graph.Graph) *Layer3Probe {
	l := &Layer3Probe{graph: g}
	g.AddEventListener(l)
	return l
}
