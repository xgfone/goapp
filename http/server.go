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

// Package http provides some convenient http functions.
package http

import (
	"crypto/tls"
	"net/http"

	"github.com/xgfone/go-apiserver/entrypoint"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-log"
)

// StartServer is a simple convenient function to start a http server.
func StartServer(name, addr string, handler http.Handler, tlsconfig *tls.Config, forceTLS bool) {
	ep, err := entrypoint.NewHTTPEntryPoint(name, addr, handler)
	if err != nil {
		log.Fatal().Str("name", name).Str("addr", addr).Err(err).
			Printf("fail to start the http server")
	}

	atexit.Register(ep.Stop)
	ep.ForceTLS = forceTLS
	ep.TLSConfig = tlsconfig
	ep.OnShutdown(atexit.Execute)
	ep.Start()
}
