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

package goapp

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/xgfone/gconf/v5"
	"github.com/xgfone/go-tools/v7/execution"
	"github.com/xgfone/goapp/config"
	"github.com/xgfone/goapp/log"
	"github.com/xgfone/goapp/router"
	"github.com/xgfone/gover"
	"github.com/xgfone/ship/v3/router/echo"
)

// AppRouter is the default app router.
var AppRouter = router.App

func init() {
	rand.Seed(time.Now().UnixNano())
	echo.RemoveTrailingSlash = true
	http.DefaultClient.Timeout = time.Second * 3
	execution.DefaultTimeout = time.Second * 3

	tp := http.DefaultTransport.(*http.Transport)
	tp.IdleConnTimeout = time.Second * 30
	tp.MaxIdleConnsPerHost = 100
	tp.MaxIdleConns = 0
}

// Init is equal to InitApp(appName, gover.Text(), configOptions...).
func Init(appName string, configOptions ...interface{}) {
	InitApp(appName, gover.Text(), configOptions...)
}

// InitApp initializes the application, which is equal to
//   InitApp2(appName, version, "100M", 100, options...)
func InitApp(appName, version string, options ...interface{}) {
	InitApp2(appName, version, "100M", 100, options...)
}

// InitApp2 initializes the application.
//
//  1. Register the log options.
//  2. Initialize configuration.
//  3. Initialize the logging.
//
func InitApp2(appName, version, logfilesize string, logfilenum int, options ...interface{}) {
	gconf.RegisterOpts(log.LogOpts...)
	config.InitConfig(appName, options, version)

	logfile := gconf.GetString(log.LogOpts[0].Name)
	loglevel := gconf.GetString(log.LogOpts[1].Name)
	log.InitLogging2(loglevel, logfile, logfilesize, logfilenum)
	if appName != "" {
		log.GetDefaultLogger().Name = appName
	}
}
