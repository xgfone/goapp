// Copyright 2022 xgfone
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

// Package httpserver provides some http server functions.
package httpserver

import (
	"crypto/tls"
	"net/http"

	"github.com/xgfone/go-apiserver/entrypoint"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-log"
)

// Start is a simple convenient function to start a http server.
func Start(name, addr string, handler http.Handler, tlsconfig *tls.Config, forceTLS bool) {
	ep := entrypoint.NewEntryPoint(name, addr, handler)
	if err := ep.Init(); err != nil {
		log.Fatal().Str("name", name).Str("addr", addr).Err(err).
			Printf("fail to start the http server")
	}

	ep.SetTLSConfig(tlsconfig)
	ep.SetTLSForce(forceTLS)
	atexit.Register(ep.Stop)
	ep.OnShutdown(atexit.Execute)
	ep.Start()
}
