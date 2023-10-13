// Copyright 2020~2023 xgfone
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
	"expvar"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-atexit/signal"
	"github.com/xgfone/go-defaults"
	"github.com/xgfone/goapp/config"
	"github.com/xgfone/goapp/log"
)

var (
	// Version is the version of app.
	//
	// Default: github.com/xgfone/gover.Text()
	Version string

	// AppName is the name of app.
	//
	// Default: ""
	AppName string
)

var (
	loglevel = gconf.StrOpt("log.level", "The level of the log, such as trace, debug, info, warn, error, etc.").
			As("loglevel").D("info").U(updateLogLevel)
	logfile0 = gconf.StrOpt("log.file", "The file path of the log. The default is stderr.").
			As("logfile")
	logfilenum = gconf.IntOpt("log.filenum", "The number of the log files.").D(100)
)

func init() {
	gconf.Conf.Errorf = func(format string, args ...interface{}) {
		if len(args) > 0 {
			format = fmt.Sprintf(format, args...)
		}
		slog.Error(format)
	}
}

func init() {
	now := time.Now().Format(time.RFC3339Nano)
	expvar.NewString("starttime").Set(now)
	defaults.ExitFunc.Set(atexit.Exit)
}

func init() {
	if tp, ok := http.DefaultTransport.(*http.Transport); ok {
		tp.DialContext = dialContext
		tp.TLSHandshakeTimeout = time.Second * 2
		tp.IdleConnTimeout = time.Second * 30
		tp.MaxIdleConnsPerHost = 100
		tp.MaxIdleConns = 0
	}
}

func dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	d := net.Dialer{Timeout: 3 * time.Second, KeepAlive: 30 * time.Second}
	return d.DialContext(ctx, network, addr)
}

func updateLogLevel(old, new interface{}) {
	if err := log.SetLevel(new.(string)); err != nil {
		slog.Error("update the log level", "old", old, "new", new, "err", err)
	} else {
		slog.Info("update the log level", "old", old, "new", new)
	}
}

// Init is used to initialize the application.
//
//  1. Register the log options.
//  2. Initialize configuration.
//  3. Initialize the logging.
//  4. Call the registered initialization functions.
//  5. Start a goroutine to monitor the exit signals.
func Init(opts ...gconf.Opt) {
	gconf.RegisterOpts(logfile0, loglevel, logfilenum)
	config.Init(AppName, Version, opts...)

	logfile := gconf.GetString(logfile0.Name)
	loglevel := gconf.GetString(loglevel.Name)
	logfilenum := gconf.GetInt(logfilenum.Name)
	log.Init(AppName, loglevel, logfile, logfilenum)

	atexit.Init()
	go signal.WaitExit(atexit.Execute)
}

// ServeHTTPWithListener starts the http server with listener until it is stopped.
func ServeHTTPWithListener(server *http.Server, ln net.Listener) {
	atexit.OnExit(func() { _ = server.Shutdown(context.Background()) })
	serveHTTP(server, ln)
	atexit.Wait()
}

func serveHTTP(server *http.Server, ln net.Listener) {
	slog.Info("start the http server", "addr", server.Addr)
	defer slog.Info("stop the http server", "addr", server.Addr)
	_ = server.Serve(ln)
}
