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

package main

import (
	"fmt"
	stdlog "log"

	"github.com/go-stack/stack"
	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-log"
	"github.com/xgfone/go-log/logf"
	"github.com/xgfone/go-log/writer"
	"github.com/xgfone/gover"
	"github.com/xgfone/ship/v5"
	"github.com/xgfone/ship/v5/middleware"
)

var logopts = []gconf.Opt{
	gconf.StrOpt("file", "The path of the log file.").D("/var/log/appName.log"),
	gconf.StrOpt("level", "The level of the log, such as debug, info, etc.").D("info"),
}

// Recover is a ship middleware to recover the panic if exists.
func Recover(next ship.Handler) ship.Handler {
	return func(ctx *ship.Context) (err error) {
		defer func() {
			switch e := recover().(type) {
			case nil:
			default:
				s := stack.Trace().TrimBelow(stack.Caller(1)).TrimRuntime()
				if len(s) == 0 {
					err = fmt.Errorf("panic: %v", e)
				} else {
					err = fmt.Errorf("panic: %v, stack=%v", e, s)
				}
			}
		}()

		return next(ctx)
	}
}

func main() {
	// Register the options.
	addr := gconf.NewString("addr", ":80", "The address to listen to.")
	loggroups := gconf.Group("log")
	loggroups.RegisterOpts(logopts...)

	// Initialize the config
	gconf.Conf.Errorf = log.Errorf
	gconf.SetVersion(gover.Text())
	gconf.AddAndParseOptFlag(gconf.Conf)
	gconf.LoadSource(gconf.NewFlagSource())
	gconf.LoadSource(gconf.NewEnvSource("appName"))
	configFile := gconf.GetString(gconf.ConfigFileOpt.Name)
	gconf.LoadAndWatchSource(gconf.NewFileSource(configFile))

	// Initialize the logging.
	log.SetLevel(log.ParseLevel(loggroups.GetString("level")))
	file := log.FileWriter(loggroups.GetString("file"), "100M", 100)
	log.SetWriter(writer.SafeWriter(file))
	stdlog.SetOutput(log.DefaultLogger)
	atexit.Register(func() { file.Close() })

	// TODO ...

	// Initialize the app router.
	app := ship.Default()
	app.Logger = logf.NewLogger(nil, 0)
	app.Use(middleware.Logger(), Recover)
	app.Route("/path1").GET(ship.OkHandler())
	app.Route("/path2").GET(func(c *ship.Context) error { return c.Text(200, "OK") })

	// Start the HTTP server.
	runner := ship.NewRunner(app)
	runner.RegisterOnShutdown(atexit.Execute)
	runner.Start(addr.Get())
}
