/*
 * Copyright (C) 2018 IBM, Inc.
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

package k8s

import (
	"fmt"

	"github.com/skydive-project/skydive/probe"
	"github.com/skydive-project/skydive/topology"
	"github.com/skydive-project/skydive/topology/graph"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

var namespaceEventHandler = graph.NewEventHandler(100)

type namespaceHandler struct {
}

func (h *namespaceHandler) Dump(obj interface{}) string {
	ns := obj.(*v1.Namespace)
	return fmt.Sprintf("namespace{Name: %s}", ns.Name)
}

func (h *namespaceHandler) Map(obj interface{}) (graph.Identifier, graph.Metadata) {
	ns := obj.(*v1.Namespace)

	m := NewMetadataFields(&ns.ObjectMeta)
	m.SetField("Status", ns.Status.Phase)

	return graph.Identifier(ns.GetUID()), NewMetadata(Manager, "namespace", m, ns, ns.Name)
}

func newNamespaceProbe(client interface{}, g *graph.Graph) Subprobe {
	return NewResourceCache(client.(*kubernetes.Clientset).Core().RESTClient(), &v1.Namespace{}, "namespaces", g, &namespaceHandler{}, namespaceEventHandler)
}

func newNamespaceLinker(g *graph.Graph, manager string, types ...string) probe.Probe {
	namespaceFilter := newTypesFilter(Manager, "namespace")
	namespaceIndexer := newObjectIndexerFromFilter(g, namespaceEventHandler, namespaceFilter, MetadataFields("Name")...)
	namespaceIndexer.Start()

	objectFilter := newTypesFilter(manager, types...)
	objectIndexer := newObjectIndexerFromFilter(g, g, objectFilter, MetadataFields("Namespace")...)
	objectIndexer.Start()

	return graph.NewMetadataIndexerLinker(g, namespaceIndexer, objectIndexer, topology.OwnershipMetadata())
}
