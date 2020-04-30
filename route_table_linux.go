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

// +build linux

package goapp

import (
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
)

const rtnUNICAST = 0x1

func getRoutes(tid int, family int, tableType int) (Routes, error) {
	routeFilter := &netlink.Route{
		Table: tid,
		Type:  tableType,
	}

	filterMask := netlink.RT_FILTER_TABLE | netlink.RT_FILTER_TYPE
	routes, err := netlink.RouteListFiltered(family, routeFilter, filterMask)
	if err != nil {
		return nil, err
	}

	rs := make(Routes, len(routes))
	for i, r := range routes {
		rs[i] = Route{
			Src: r.Src,
			Dst: r.Dst,
			Gw:  r.Gw,
		}
	}

	return rs, nil
}

// GetIPv4Routes returns the ipv4 routes.
func GetIPv4Routes() (Routes, error) {
	return getRoutes(0, nl.FAMILY_V4, rtnUNICAST)
}

// GetIPv6Routes returns the ipv4 routes.
func GetIPv6Routes() (Routes, error) {
	return getRoutes(0, nl.FAMILY_V6, rtnUNICAST)
}
