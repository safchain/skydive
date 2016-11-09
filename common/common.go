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

package common

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const (
	StoppedState = iota
	RunningState
	StoppingState
)

type CaptureType struct {
	Allowed []string
	Default string
}

var (
	CantCompareInterface error = errors.New("Can't compare interface")
	CaptureTypes               = map[string]CaptureType{}
)

func initCaptureTypes() {
	// add ovs type
	CaptureTypes["ovsbridge"] = CaptureType{Allowed: []string{"ovssflow"}, Default: "ovssflow"}

	// anything else will be handled by gopacket
	types := []string{
		"device", "internal", "veth", "tun", "bridge", "dummy", "gre",
		"bond", "can", "hsr", "ifb", "macvlan", "macvtap", "vlan", "vxlan",
		"gretap", "ip6gretap", "geneve", "ipoib", "vcan", "ipip", "ipvlan",
		"lowpan", "ip6tnl", "ip6gre", "sit",
	}

	for _, t := range types {
		CaptureTypes[t] = CaptureType{Allowed: []string{"afpacket", "pcap"}, Default: "afpacket"}
	}
}

func init() {
	initCaptureTypes()
}

func toInt64(i interface{}) (int64, error) {
	switch i.(type) {
	case int:
		return int64(i.(int)), nil
	case uint:
		return int64(i.(uint)), nil
	case int32:
		return int64(i.(int32)), nil
	case uint32:
		return int64(i.(uint32)), nil
	case int64:
		return i.(int64), nil
	case uint64:
		return int64(i.(uint64)), nil
	case float32:
		return int64(i.(float32)), nil
	case float64:
		return int64(i.(float64)), nil
	}
	return 0, fmt.Errorf("not an integer: %v", i)
}

func integerCompare(a interface{}, b interface{}) (int, error) {
	n1, err := toInt64(a)
	if err != nil {
		return 0, err
	}

	n2, err := toInt64(b)
	if err != nil {
		return 0, err
	}

	if n1 == n2 {
		return 0, nil
	} else if n1 < n2 {
		return -1, nil
	} else {
		return 1, nil
	}
}

func toFloat64(f interface{}) (float64, error) {
	switch f.(type) {
	case int, uint, int32, uint32, int64, uint64:
		i, err := toInt64(f)
		if err != nil {
			return 0, err
		}
		return float64(i), nil
	case float32:
		return float64(f.(float32)), nil
	case float64:
		return f.(float64), nil
	}
	return 0, fmt.Errorf("not a float: %v", f)
}

func floatCompare(a interface{}, b interface{}) (int, error) {
	f1, err := toFloat64(a)
	if err != nil {
		return 0, err
	}

	f2, err := toFloat64(b)
	if err != nil {
		return 0, err
	}

	if f1 == f2 {
		return 0, nil
	} else if f1 < f2 {
		return -1, nil
	} else {
		return 1, nil
	}
}

func CrossTypeCompare(a interface{}, b interface{}) (int, error) {
	switch a.(type) {
	case float32, float64:
		return floatCompare(a, b)
	}

	switch b.(type) {
	case float32, float64:
		return floatCompare(a, b)
	}

	switch a.(type) {
	case int, uint, int32, uint32, int64, uint64:
		return integerCompare(a, b)
	default:
		return 0, CantCompareInterface
	}
}

func CrossTypeEqual(a interface{}, b interface{}) bool {
	result, err := CrossTypeCompare(a, b)
	if err == CantCompareInterface {
		return a == b
	} else if err != nil {
		return false
	}
	return result == 0
}

// Retry tries to execute the given function until a success applying a delay
// between each try
func Retry(fnc func() error, try int, delay time.Duration) error {
	var err error
	if err = fnc(); err == nil {
		return nil
	}

	for i := 0; i != try; i++ {
		time.Sleep(delay)
		if err = fnc(); err == nil {
			return nil
		}
	}

	return err
}

func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func IsCaptureAllowed(nodeType string) bool {
	_, ok := CaptureTypes[nodeType]
	return ok
}

type Iterator struct {
	at, from, to int64
}

func (it *Iterator) Done() bool {
	return it.to != -1 && it.at >= it.to
}

func (it *Iterator) Next() bool {
	it.at++
	return it.at-1 >= it.from
}

func NewIterator(values ...int64) (it *Iterator) {
	it = &Iterator{to: -1}
	if len(values) > 0 {
		it.at = values[0]
	}
	if len(values) > 1 {
		it.from = values[1]
	}
	if len(values) > 2 {
		it.to = values[2]
	}
	return
}

func IPv6Supported() bool {
	if _, err := os.Stat("/proc/net/if_inet6"); os.IsNotExist(err) {
		return false
	}

	data, err := ioutil.ReadFile("/proc/sys/net/ipv6/conf/all/disable_ipv6")
	if err != nil {
		return false
	}

	if strings.TrimSpace(string(data)) == "1" {
		return false
	}

	return true
}
