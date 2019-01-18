/*
 * Copyright (C) 2016 Red Hat, Inc.
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

package etcd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	etcd "github.com/coreos/etcd/client"

	"github.com/skydive-project/skydive/common"
)

// Client describes a ETCD configuration client
type Client struct {
	service common.Service
	client  *etcd.Client
	KeysAPI etcd.KeysAPI
}

// GetInt64 returns an int64 value from the configuration key
func (client *Client) GetInt64(key string) (int64, error) {
	resp, err := client.KeysAPI.Get(context.Background(), key, nil)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(resp.Node.Value, 10, 64)
}

// SetInt64 set an int64 value to the configuration key
func (client *Client) SetInt64(key string, value int64) error {
	_, err := client.KeysAPI.Set(context.Background(), key, strconv.FormatInt(value, 10), nil)
	return err
}

// Stop the client
func (client *Client) Stop() {
	if tr, ok := etcd.DefaultTransport.(interface {
		CloseIdleConnections()
	}); ok {
		tr.CloseIdleConnections()
	}
}

// NewElection creates a new ETCD master elector
func (client *Client) NewElection(name string) common.MasterElection {
	return NewMasterElector(client, name)
}

// NewClient creates a new ETCD client connection to ETCD servers
func NewClient(service common.Service, etcdServers []string, clientTimeout time.Duration) (*Client, error) {
	cfg := etcd.Config{
		Endpoints:               etcdServers,
		Transport:               etcd.DefaultTransport,
		HeaderTimeoutPerRequest: clientTimeout,
	}

	client, err := etcd.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to etcd: %s", err)
	}

	kapi := etcd.NewKeysAPI(client)

	return &Client{
		service: service,
		client:  &client,
		KeysAPI: kapi,
	}, nil
}
