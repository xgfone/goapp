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
	"net/http"
	"net/http/pprof"
	"runtime"
	rpprof "runtime/pprof"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-log"
	"github.com/xgfone/go-log/logf"
	"github.com/xgfone/goapp/validate"
	"github.com/xgfone/gover"
	"github.com/xgfone/ship/v5"
)

// DefaultRuntimeRouteConfig is the default RuntimeRouteConfig
// with DefaultShellConfig.
var DefaultRuntimeRouteConfig = RuntimeRouteConfig{ShellConfig: DefaultShellConfig}

// Config is used to configure the app router.
type Config struct {
	LogReqBody bool
}

// InitRouter returns a new ship router.
func InitRouter(c *Config) *ship.Ship {
	var config Config
	if c != nil {
		config = *c
	}

	app := ship.Default()
	app.Validator = ship.ValidatorFunc(validate.StructValidator(nil))
	app.Pre(Logger(config.LogReqBody), Recover)
	app.Logger = logf.NewLogger(log.DefaultLogger, 0)
	return app
}

// StartServer starts the HTTP server.
func StartServer(addr string, handler http.Handler) {
	runner := ship.NewRunner(handler)
	runner.RegisterOnShutdown(atexit.Execute)
	atexit.Register(runner.Stop)
	runner.Start(addr)
}

// RuntimeRouteConfig is used to configure the runtime routes.
type RuntimeRouteConfig struct {
	ShellConfig

	Prefix    string
	IsReady   func() bool
	IsHealthy func() bool

	Config *gconf.Config
}

// AddRuntimeRoutes adds the runtime routes.
func AddRuntimeRoutes(app *ship.Ship, config ...RuntimeRouteConfig) {
	var conf RuntimeRouteConfig
	if len(config) > 0 {
		conf = config[0]
	}

	boolHandler := func(f func() bool) ship.Handler {
		return func(ctx *ship.Context) error {
			if f == nil || f() {
				return nil
			}
			return ship.ErrServiceUnavailable
		}
	}

	group := app.Group(conf.Prefix).Group("/runtime")
	group.Route("/version").GET(getVersion)
	group.Route("/configs").GET(getAllConfigs(conf.Config))
	group.Route("/routes").GET(getAllRoutes(app))
	group.Route("/ready").GET(boolHandler(conf.IsReady))
	group.Route("/healthy").GET(boolHandler((conf.IsHealthy)))
	group.Route("/metrics").GET(ship.FromHTTPHandler(promhttp.Handler()))
	group.Route("/debug/vars").GET(ship.FromHTTPHandler(expvar.Handler()))
	group.Route("/debug/pprof/profile").GET(ship.FromHTTPHandlerFunc(pprof.Profile))
	group.Route("/debug/pprof/cmdline").GET(ship.FromHTTPHandlerFunc(pprof.Cmdline))
	for _, p := range rpprof.Profiles() {
		group.Route("/debug/pprof/" + p.Name()).GET(ship.FromHTTPHandler(pprof.Handler(p.Name())))
	}

	if conf.ShellConfig.Shell != "" {
		group.Route("/shell").POST(ExecuteShell(nil, conf.ShellConfig))
	}
}

func getAllRoutes(s *ship.Ship) ship.Handler {
	return func(c *ship.Context) error { return c.JSON(200, s.Routes()) }
}

func getAllConfigs(conf *gconf.Config) ship.Handler {
	if conf == nil {
		conf = gconf.Conf
	}

	return func(c *ship.Context) error {
		_, snap := conf.Snapshot()
		return c.JSON(200, snap)
	}
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
