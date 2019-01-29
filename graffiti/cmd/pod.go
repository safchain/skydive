/*
 * Copyright (C) 2018 Red Hat, Inc.
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

package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	api "github.com/skydive-project/skydive/api/server"
	"github.com/skydive-project/skydive/common"
	"github.com/skydive-project/skydive/graffiti/graph"
	"github.com/skydive-project/skydive/graffiti/graph/traversal"
	"github.com/skydive-project/skydive/graffiti/pod"
	ge "github.com/skydive-project/skydive/gremlin/traversal"
	shttp "github.com/skydive-project/skydive/http"
	"github.com/skydive-project/skydive/logging"
	"github.com/skydive-project/skydive/websocket"
	"github.com/spf13/cobra"
)

var (
	hubServers  []string
	podListen   string
	serviceType = common.ServiceType("Pod")
)

func newHubClientPool(host string, addresses []common.ServiceAddress, opts websocket.ClientOpts) *websocket.StructClientPool {
	pool := websocket.NewStructClientPool("HubClientPool")

	for _, sa := range addresses {
		url, _ := url.Parse(fmt.Sprintf("ws://%s:%d/ws/pod", sa.Addr, sa.Port))
		client := websocket.NewClient(host, serviceType, url, opts)
		pool.AddClient(client)
	}

	return pool
}

// PodCmd describes the graffiti pod command
var PodCmd = &cobra.Command{
	Use:          "pod",
	Short:        "Graffiti pod",
	Long:         "Graffiti pod",
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		logging.GetLogger().Noticef("Graffiti pod starting...")

		sa, err := common.ServiceAddressFromString(podListen)
		if err != nil {
			logging.GetLogger().Errorf("Configuration error: %s", err)
			os.Exit(1)
		}

		hostname, err := os.Hostname()
		if err != nil {
			logging.GetLogger().Errorf("Failed to get hostname: %s", err)
			os.Exit(1)
		}

		service := common.Service{ID: hostname, Type: serviceType}
		backend, err := graph.NewMemoryBackend()
		if err != nil {
			logging.GetLogger().Errorf("Failed to get hostname: %s", err)
			os.Exit(1)
		}

		clusterAuthOptions := &shttp.AuthenticationOpts{}

		// declare all extension available throught API and filtering
		tr := traversal.NewGremlinTraversalParser()
		tr.AddTraversalExtension(ge.NewDescendantsTraversalExtension())

		g := graph.NewGraph(hostname, backend, serviceType)

		authBackend := shttp.NewNoAuthenticationBackend()

		httpServer := shttp.NewServer(hostname, serviceType, sa.Addr, sa.Port, nil)

		if err := httpServer.Listen(); err != nil {
			logging.GetLogger().Error(err)
			os.Exit(1)
		}

		apiServer, err := api.NewAPI(httpServer, nil, service, authBackend)
		if err != nil {
			logging.GetLogger().Error(err)
			os.Exit(1)
		}
		api.RegisterTopologyAPI(httpServer, g, tr, authBackend)

		var addresses []common.ServiceAddress
		for _, address := range hubServers {
			sa, err := common.ServiceAddressFromString(address)
			if err != nil {
				logging.GetLogger().Error(err)
				os.Exit(1)
			}
			addresses = append(addresses, sa)
		}

		if len(addresses) == 0 {
			logging.GetLogger().Info("Pod is running in standalone mode")
		}

		opts := websocket.ClientOpts{
			AuthOpts:         clusterAuthOptions,
			WriteCompression: writeCompression,
			QueueSize:        queueSize,
		}

		clientPool := newHubClientPool(hostname, addresses, opts)

		pod, err := pod.NewPod(apiServer, clientPool, g, authBackend, nil, tr, writeCompression, queueSize, time.Second*time.Duration(pingDelay), time.Second*time.Duration(pongTimeout))
		if err != nil {
			logging.GetLogger().Error(err)
			os.Exit(1)
		}

		go httpServer.Serve()

		// everything is ready, then initiate the websocket connection
		go clientPool.ConnectAll()

		pod.Start()

		logging.GetLogger().Notice("Graffiti pod started !")
		ch := make(chan os.Signal)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch

		logging.GetLogger().Notice("Graffiti pod stopped.")
	},
}

func init() {
	PodCmd.Flags().StringArrayVar(&hubServers, "hubs", nil, "address and port for the pod server")
	PodCmd.Flags().StringVarP(&podListen, "listen", "l", "127.0.0.1:8081", "address and port for the pod server")
	PodCmd.Flags().IntVar(&queueSize, "queueSize", 10000, "websocket queue size")
	PodCmd.Flags().IntVar(&pingDelay, "pingDelay", 2, "websocket ping delay")
	PodCmd.Flags().IntVar(&pongTimeout, "pongTimeout", 10, "websocket pong timeout")
}
