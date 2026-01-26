// Copyright 2026 xgfone
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
	"context"
	"net"
	"net/http"
	"time"
)

func init() {
	if tp, ok := http.DefaultTransport.(*http.Transport); ok {
		tp.DialContext = dialContext
		tp.TLSHandshakeTimeout = time.Second * 3
		tp.IdleConnTimeout = time.Second * 30
		tp.MaxIdleConnsPerHost = 100
		tp.MaxIdleConns = 0
	}
}

func dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	d := net.Dialer{Timeout: 3 * time.Second, KeepAlive: 30 * time.Second}
	return d.DialContext(ctx, network, addr)
}
