// Copyright 2020~2022 xgfone
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

	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-atexit/signal"
	"github.com/xgfone/goapp/config"
	_ "github.com/xgfone/goapp/exec" // import to initialize the log hook
	glog "github.com/xgfone/goapp/log"
)

var (
	loglevel = gconf.StrOpt("log.level", "The level of the log, such as trace, debug, info, warn, error, etc.").
			As("loglevel").D("info").U(updateLogLevel)
	logfile0 = gconf.StrOpt("log.file", "The file path of the log. The default is stderr.").
			As("logfile")
	logfilenum = gconf.IntOpt("log.filenum", "The number of the log files.").D(100)
)

func init() {
	gconf.Conf.Errorf = log.Errorf
	rand.Seed(time.Now().UnixNano())

	if tp, ok := http.DefaultTransport.(*http.Transport); ok {
		tp.IdleConnTimeout = time.Second * 30
		tp.MaxIdleConnsPerHost = 100
		tp.MaxIdleConns = 0
	}
}

func updateLogLevel(old, new interface{}) {
	err := glog.SetLevel(new.(string))
	log.Info("update the log level", "old", old, "new", new, "err", err)
}

// Init is used to initialize the application.
//
//  1. Register the log options.
//  2. Initialize configuration.
//  3. Initialize the logging.
//  4. Call the registered initialization functions.
//  5. Start a goroutine to monitor the exit signals.
func Init(appName string, opts ...gconf.Opt) {
	gconf.RegisterOpts(logfile0, loglevel, logfilenum)
	config.InitConfig(appName, "", opts...)

	logfile := gconf.GetString(logfile0.Name)
	loglevel := gconf.GetString(loglevel.Name)
	logfilenum := gconf.GetInt(logfilenum.Name)
	glog.InitLoging(appName, loglevel, logfile, logfilenum)

	atexit.Init()
	go signal.WaitExit(atexit.Execute)
}
