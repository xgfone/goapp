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
	"github.com/xgfone/gconf/v5"
	"github.com/xgfone/gconf/v5/field"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-log"
	"github.com/xgfone/gover"
	"github.com/xgfone/ship/v4"
	"github.com/xgfone/ship/v4/middleware"
)

// Config is used to configure the app.
type Config struct {
	Addr     field.StringOptField `default:":80" help:"The address to listen to."`
	LogFile  field.StringOptField `default:"" help:"The path of the log file."`
	LogLevel field.StringOptField `default:"info" help:"The level of the log, such as debug, info, etc."`
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
	var conf Config

	// Initialize the config
	gconf.SetErrHandler(gconf.ErrorHandler(func(err error) { log.Errorf(err.Error()) }))
	gconf.RegisterStruct(&conf)
	gconf.SetStringVersion(gover.Text())
	gconf.AddAndParseOptFlag(gconf.Conf)
	gconf.LoadSource(gconf.NewFlagSource())
	gconf.LoadSource(gconf.NewEnvSource("PREFIX"))
	gconf.LoadSource(gconf.NewFileSource(gconf.GetString(gconf.ConfigFileOpt.Name)))

	// Initialize the log
	log.DefalutLogger.Level = log.NameToLevel(conf.LogLevel.Get())
	writer := log.FileWriter(conf.LogFile.Get(), "100M", 100)
	log.DefalutLogger.Encoder.SetWriter(log.SafeWriter(writer))
	stdlog.SetOutput(log.NewIOWriter(writer, log.LvlTrace))
	atexit.Register(func() { writer.Close() })

	// TODO ...

	// Initialize and start the app.
	app := ship.Default()
	app.SetLogger(log.DefalutLogger)
	app.RegisterOnShutdown(atexit.Execute)
	app.Use(middleware.Logger(nil), Recover)
	app.Route("/path1").GET(ship.OkHandler())
	app.Route("/path2").GET(func(c *ship.Context) error { return c.Text(200, "OK") })
	app.Start(conf.Addr.Get()).Wait()
}
