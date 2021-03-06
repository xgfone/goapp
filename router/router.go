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
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-log"
	"github.com/xgfone/goapp/config"
	"github.com/xgfone/goapp/validate"
	"github.com/xgfone/gover"
	"github.com/xgfone/ship/v4"
	"github.com/xgfone/ship/v4/middleware"
	"github.com/xgfone/ship/v4/router/echo"
)

// App is the default global router app.
var App = InitRouter()

// DefaultRuntimeRouteConfig is the default RuntimeRouteConfig
// with DefaultShellConfig.
var DefaultRuntimeRouteConfig = RuntimeRouteConfig{ShellConfig: DefaultShellConfig}

// Config is used to configure the app router.
type Config struct {
	middleware.LoggerConfig
}

// InitRouter returns a new ship router.
func InitRouter(config ...Config) *ship.Ship {
	var rconf Config
	if len(config) > 0 {
		rconf = config[0]
	}
	if rconf.LoggerConfig.Log == nil {
		rconf.LoggerConfig.Log = logRequest
	}

	app := ship.Default()
	app.Validator = validate.StructValidator(nil)
	app.Use(middleware.Logger(&rconf.LoggerConfig), Recover)
	app.RegisterOnShutdown(atexit.Execute)
	app.SetLogger(log.DefalutLogger)
	app.SetNewRouter(func() ship.Router {
		return echo.NewRouter(&echo.Config{RemoveTrailingSlash: true})
	})

	atexit.Register(app.Stop)
	return app
}

func logRequest(req *http.Request, hasReqBody bool, reqBody string, code int,
	startTime time.Time, cost time.Duration, err error) {
	fields := make([]log.Field, 6, 8)
	fields[0] = log.F("addr", req.RemoteAddr)
	fields[1] = log.F("method", req.Method)
	fields[2] = log.F("uri", req.RequestURI)
	fields[3] = log.F("code", code)
	fields[4] = log.F("start", startTime.Unix())
	fields[5] = log.F("cost", cost.String())
	if hasReqBody {
		fields = append(fields, log.F("reqbody", reqBody))
	}
	if err != nil {
		fields = append(fields, log.E(err))
	}

	if code < 400 {
		log.Info("request", fields...)
	} else if code < 500 {
		log.Warn("request", fields...)
	} else {
		log.Error("request", fields...)
	}
}

// RuntimeRouteConfig is used to configure the runtime routes.
type RuntimeRouteConfig struct {
	ShellConfig

	Prefix    string
	IsReady   func() bool
	IsHealthy func() bool

	Config *config.Config
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

func getAllConfigs(conf *config.Config) ship.Handler {
	return func(c *ship.Context) error {
		return c.JSON(200, config.GetAllConfigs(conf))
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
