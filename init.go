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
)

func init() {
	rand.Seed(time.Now().UnixNano())
	http.DefaultClient.Timeout = time.Second * 3
	execution.DefaultCmd.Timeout = time.Second * 3
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
	gconf.RegisterOpts(LogOpts...)
	InitConfig(appName, options, version)

	logfile := gconf.GetString(LogOpts[0].Name)
	loglevel := gconf.GetString(LogOpts[1].Name)
	InitLogging2(loglevel, logfile, logfilesize, logfilenum)
}
