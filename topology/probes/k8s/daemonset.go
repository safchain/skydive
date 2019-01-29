/*
 * Copyright (C) 2018 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy ofthe License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specificlanguage governing permissions and
 * limitations under the License.
 *
 */

package k8s

import (
	"fmt"

	"github.com/skydive-project/skydive/graffiti/graph"

	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"
)

type daemonSetHandler struct {
}

func (h *daemonSetHandler) Dump(obj interface{}) string {
	ds := obj.(*v1beta1.DaemonSet)
	return fmt.Sprintf("daemonset{Namespace: %s, Name: %s}", ds.Namespace, ds.Name)
}

func (h *daemonSetHandler) Map(obj interface{}) (graph.Identifier, graph.Metadata) {
	ds := obj.(*v1beta1.DaemonSet)

	m := NewMetadataFields(&ds.ObjectMeta)
	m.SetField("DesiredNumberScheduled", ds.Status.DesiredNumberScheduled)
	m.SetField("CurrentNumberScheduled", ds.Status.CurrentNumberScheduled)
	m.SetField("NumberMisscheduled", ds.Status.NumberMisscheduled)

	return graph.Identifier(ds.GetUID()), NewMetadata(Manager, "daemonset", m, ds, ds.Name)
}

func newDaemonSetProbe(client interface{}, g *graph.Graph) Subprobe {
	return NewResourceCache(client.(*kubernetes.Clientset).ExtensionsV1beta1().RESTClient(), &v1beta1.DaemonSet{}, "daemonsets", g, &daemonSetHandler{})
}
