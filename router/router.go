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

// Package router suppies some router assistant functions, such as the router
// Handlers and Middlewares.
package router

import (
	"expvar"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xgfone/go-tools/v7/lifecycle"
	"github.com/xgfone/goapp/validate"
	"github.com/xgfone/gover"
	"github.com/xgfone/klog/v3"
	"github.com/xgfone/ship/v3"
	"github.com/xgfone/ship/v3/middleware"
)

// App is the default global router app.
var App = InitRouter()

// DefaultRuntimeRouteConfig is the default RuntimeRouteConfig
// with DefaultShellConfig.
var DefaultRuntimeRouteConfig = RuntimeRouteConfig{ShellConfig: DefaultShellConfig}

// InitRouter returns a new ship router.
func InitRouter() *ship.Ship {
	app := ship.Default()
	app.Use(middleware.Logger(), Recover)
	app.RegisterOnShutdown(lifecycle.Stop)
	app.SetLogger(klog.ToFmtLogger(klog.GetDefaultLogger()))
	app.Validator = validate.StructValidator(nil)
	return app
}

// RuntimeRouteConfig is used to configure the runtime routes.
type RuntimeRouteConfig struct {
	ShellConfig

	Prefix    string
	IsReady   func() bool
	IsHealthy func() bool
}

// AddRuntimeRoutes adds the runtime routes.
func AddRuntimeRoutes(app *ship.Ship, config ...RuntimeRouteConfig) {
	var conf RuntimeRouteConfig
	if len(config) > 0 {
		conf = config[0]
	}

	boolHandler := func(f func() bool) ship.Handler {
		return func(ctx *ship.Context) error {
			if f == nil && f() {
				return nil
			}
			return ship.ErrServiceUnavailable
		}
	}

	group := app.Group(conf.Prefix).Group("/runtime")
	group.R("/version").GET(getVersion)
	group.R("/routes").GET(getAllRoutes(app))
	group.R("/ready").GET(boolHandler(conf.IsReady))
	group.R("/healthy").GET(boolHandler((conf.IsHealthy)))
	group.R("/metrics").GET(ship.FromHTTPHandler(promhttp.Handler()))
	group.R("/debug/vars").GET(ship.FromHTTPHandler(expvar.Handler()))
	group.AddRoutes(ship.HTTPPprofToRouteInfo()...)

	if conf.ShellConfig.Shell != "" {
		group.R("/shell").POST(ExecuteShell(nil, conf.ShellConfig))
	}
}

func getAllRoutes(s *ship.Ship) ship.Handler {
	return func(c *ship.Context) error { return c.JSON(200, s.Routes()) }
}

func getVersion(ctx *ship.Context) error {
	return ctx.JSON(200, map[string]string{
		"commit":    gover.Commit,
		"version":   gover.Version,
		"goversion": runtime.Version(),

		"build":   gover.GetBuildTime().Format(time.RFC3339Nano),
		"start":   gover.StartTime.Format(time.RFC3339Nano),
		"elapsed": gover.GetElapsedTime().String(),
	})
}
