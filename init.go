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
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-log"
	"github.com/xgfone/go-log/logf"
	"github.com/xgfone/goapp/config"
	_ "github.com/xgfone/goapp/exec" // import to initialize the log hook
	glog "github.com/xgfone/goapp/log"
	_ "github.com/xgfone/goapp/validate" // import to initialize the validator
)

var (
	logfile  = gconf.StrOpt("log.file", "The file path of the log. The default is stdout.")
	loglevel = gconf.StrOpt("log.level", "The level of the log, such as debug, info, etc.").D("info")
)

func init() {
	gconf.Conf.Errorf = logf.Errorf
	rand.Seed(time.Now().UnixNano())

	tp := http.DefaultTransport.(*http.Transport)
	tp.IdleConnTimeout = time.Second * 30
	tp.MaxIdleConnsPerHost = 100
	tp.MaxIdleConns = 0
}

var inits []func() error

// RegisterInit registers the initialization functions.
func RegisterInit(initfuncs ...func() error) {
	inits = append(inits, initfuncs...)
}

// CallInit calls the registered initialization functions.
func CallInit() (err error) {
	for _, f := range inits {
		if err = f(); err != nil {
			return
		}
	}
	return
}

// Init is used to initialize the application.
//
//  1. Register the log options.
//  2. Initialize configuration.
//  3. Initialize the logging.
//  4. Call the registered initialization functions.
//
func Init(appName string, opts ...gconf.Opt) {
	gconf.RegisterOpts(logfile, loglevel)
	config.InitConfig(appName, "", opts...)
	glog.InitLoging(appName, gconf.GetString(loglevel.Name), gconf.GetString(logfile.Name))

	if err := CallInit(); err != nil {
		log.Fatal().Err(err).Printf("fail to init")
	}

	go atexit.Signals()
}
