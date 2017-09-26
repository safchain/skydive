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
	"strings"
	"testing"

	"github.com/skydive-project/skydive/topology/graph"
	"github.com/skydive-project/skydive/topology/graph/traversal"
)

func TestGraphPathTraversal(t *testing.T) {
	g := newGraph(t)

	n1 := g.NewNode(graph.GenID(), graph.Metadata{"Type": "host", "Name": "localhost"})
	n2 := g.NewNode(graph.GenID(), graph.Metadata{"Name": "N2", "Type": "T2"})
	n3 := g.NewNode(graph.GenID(), graph.Metadata{"Name": "N3", "Type": "T3"})

	g.Link(n1, n2, OwnershipMetadata)
	g.Link(n2, n3, OwnershipMetadata)

	query := `G.V().Has("Name", "N3").GraphPath()`

	tp := traversal.NewGremlinTraversalParser(g)
	tp.AddTraversalExtension(NewTopologyTraversalExtension())

	ts, err := tp.Parse(strings.NewReader(query), false)
	if err != nil {
		t.Fatal(err.Error())
	}

	res, err := ts.Exec()
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(res.Values()) != 1 || res.Values()[0].(string) != "localhost[Type=host]/N2[Type=T2]/N3[Type=T3]" {
		t.Fatalf("Should return 1 path, returned: %v", res.Values())
	}
}

func TestRegexPredicate(t *testing.T) {
	g := newGraph(t)
	g.NewNode(graph.GenID(), graph.Metadata{"Type": "host", "Name": "localhost"})

	query := `G.V().Has("Name", Regex("local.*st")).Count()`

	tp := traversal.NewGremlinTraversalParser(g)
	tp.AddTraversalExtension(NewTopologyTraversalExtension())

	ts, err := tp.Parse(strings.NewReader(query), false)
	if err != nil {
		t.Fatal(err.Error())
	}

	res, err := ts.Exec()
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(res.Values()) != 1 || res.Values()[0].(int) != 1 {
		t.Fatalf("Regex should exactly match 1 node, returned: %v", res.Values())
	}
}
