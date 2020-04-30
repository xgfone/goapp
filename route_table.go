// Copyright 2020 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goapp

import (
	"errors"
	"net"
	"strings"

	"github.com/xgfone/netaddr"
)

// ErrRouteTableNotImplemented is returned when the route table has not implemented.
var ErrRouteTableNotImplemented = errors.New("route table has not implemented")

// Route is the route information.
type Route struct {
	Dst *net.IPNet
	Src net.IP
	Gw  net.IP
}

// Routes is a set of Routes.
type Routes []Route

// DefaultGateway returns the default gateway. Return nil if not exist.
func (rs Routes) DefaultGateway() net.IP {
	return rs.defaultGateway().Gw
}

func (rs Routes) defaultGateway() Route {
	for _, r := range rs {
		if r.Gw != nil {
			return r
		}
	}
	return Route{}
}

func getDefaultGateway() (route Route, err error) {
	var routes Routes

	// Test IPv4
	if routes, err = GetIPv4Routes(); err == nil {
		if route = routes.defaultGateway(); route.Gw != nil {
			return
		}
	}

	// Test IPv6
	if routes, err = GetIPv6Routes(); err == nil {
		route = routes.defaultGateway()
	}

	return
}

// GetDefaultGateway returns the default gateway.
func GetDefaultGateway() (gateway net.IP, err error) {
	route, err := getDefaultGateway()
	gateway = route.Gw
	return
}

// GetDefaultIP returns the default ip.
func GetDefaultIP() (ip string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	switch len(addrs) {
	case 0:
		return
	case 1:
		return strings.Split(addrs[0].String(), "/")[0], nil
	}

	route, err := getDefaultGateway()
	if err != nil || route.Gw == nil {
		return
	}

	for _, addr := range addrs {
		net, err := netaddr.NewIPNetwork(addr.String())
		if err != nil {
			return "", err
		} else if net.HasStringIP(route.Gw.String()) {
			return strings.Split(addr.String(), "/")[0], nil
		}
	}

	return
}
