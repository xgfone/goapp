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

	"github.com/go-stack/stack"
	"github.com/xgfone/gconf/v4"
	"github.com/xgfone/go-tools/v6/lifecycle"
	"github.com/xgfone/gover"
	"github.com/xgfone/klog/v3"
	"github.com/xgfone/ship/v2"
	"github.com/xgfone/ship/v2/middleware"
)

// Config is used to configure the app.
type Config struct {
	Addr     gconf.StringOptField `default:":80" help:"The address to listen to."`
	LogFile  gconf.StringOptField `default:"" help:"The path of the log file."`
	LogLevel gconf.StringOptField `default:"info" help:"The level of the log, such as debug, info, etc."`
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
	gconf.SetErrHandler(gconf.ErrorHandler(func(err error) { klog.Errorf(err.Error()) }))
	gconf.RegisterStruct(&conf)
	gconf.SetStringVersion(gover.Text())
	gconf.AddAndParseOptFlag(gconf.Conf)
	gconf.LoadSource(gconf.NewFlagSource())
	gconf.LoadSource(gconf.NewEnvSource("PREFIX"))
	gconf.LoadSource(gconf.NewFileSource(gconf.GetString(gconf.ConfigFileOpt.Name)))

	// Initialize the log
	klog.SetLevel(klog.NameToLevel(conf.LogLevel.Get()))
	klog.SetDefaultLogger(klog.GetDefaultLogger().WithCtx(klog.Caller("caller")))
	if writer, err := klog.FileWriter(conf.LogFile.Get(), "100M", 100); err != nil {
		fmt.Println(err)
		lifecycle.Exit(1)
	} else {
		klog.GetEncoder().SetWriter(writer)
		lifecycle.Register(func() { writer.Close() })
	}

	// TODO ...

	// Initialize and start the app.
	app := ship.Default()
	app.RegisterOnShutdown(lifecycle.Stop)
	app.SetLogger(klog.ToFmtLogger(klog.GetDefaultLogger()))
	app.Use(middleware.Logger(), Recover)
	app.Route("/path1").GET(ship.OkHandler())
	app.Route("/path2").GET(func(c *ship.Context) error { return c.Text(200, "OK") })
	app.Start(conf.Addr.Get()).Wait()
}
