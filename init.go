// Copyright 2020~2026 xgfone
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
	"expvar"
	"time"

	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/goapp/config"
	"github.com/xgfone/goapp/internal"
)

func init() {
	now := time.Now().Format(time.RFC3339Nano)
	expvar.NewString("starttime").Set(now)
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

	initlog()
	trysetpwd()
	initversion()

	internal.RunInit()
	go internal.SignalForExit()
}
