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

package topology

import (
	"fmt"

	"github.com/skydive-project/skydive/topology/graph"
)

type NodePath []*graph.Node

func (p NodePath) Marshal() string {
	var path string
	for i := len(p) - 1; i >= 0; i-- {
		if len(path) > 0 {
			path += "/"
		}

		metadata := p[i].Metadata()
		n := metadata["Name"]
		t := metadata["Type"]

		if n == nil || t == nil {
			return ""
		}

		path += fmt.Sprintf("%s[Type=%s]", n.(string), t.(string))
	}

	return path
}

func GraphPath(g *graph.Graph, n *graph.Node) string {
	nodes := g.LookupShortestPath(n, graph.Metadata{"Type": "host"}, graph.Metadata{"RelationType": "ownership"})
	if len(nodes) > 0 {
		return NodePath(nodes).Marshal()
	}
	return ""
}
