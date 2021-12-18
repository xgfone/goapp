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

	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/go-log"
	"github.com/xgfone/go-log/logf"
	glog "github.com/xgfone/goapp/log"
	"github.com/xgfone/gover"
)

func init() {
	gconf.Conf.Errorf = logf.Errorf
	rand.Seed(time.Now().UnixNano())

	// http.DefaultClient.Timeout = time.Second * 3
	tp := http.DefaultTransport.(*http.Transport)
	tp.IdleConnTimeout = time.Second * 30
	tp.MaxIdleConnsPerHost = 100
	tp.MaxIdleConns = 0
}

var inits []func() error

// InitConfig initializes the configuration, which will set the version,
// register the options, parse the CLI arguments with "flag",
// load the "flag", "env" and "file" sources.
func InitConfig(app, version string, opts ...gconf.Opt) {
	gconf.SetVersion(version)
	gconf.RegisterOpts(opts...)
	gconf.AddAndParseOptFlag(gconf.Conf)
	gconf.LoadSource(gconf.NewFlagSource())
	gconf.LoadSource(gconf.NewEnvSource(app))
	configFile := gconf.GetString(gconf.ConfigFileOpt.Name)
	gconf.LoadAndWatchSource(gconf.NewFileSource(configFile))
}

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

// Init is equal to InitApp(appName, gover.Text(), opts...).
func Init(appName string, opts ...gconf.Opt) {
	InitApp(appName, gover.Text(), opts...)
}

// InitApp initializes the application, which is equal to
//   InitApp2(appName, version, "100M", 100, opts...)
func InitApp(appName, version string, opts ...gconf.Opt) {
	InitApp2(appName, version, "100M", 100, opts...)
}

// InitApp2 initializes the application.
//
//  1. Register the log options.
//  2. Initialize configuration.
//  3. Initialize the logging.
//  4. Call the registered initialization functions.
//
func InitApp2(appName, version, logfilesize string, logfilenum int, opts ...gconf.Opt) {
	gconf.RegisterOpts(glog.LogOpts...)
	InitConfig(appName, version, opts...)

	logfile := gconf.GetString(glog.LogOpts[0].Name)
	loglevel := gconf.GetString(glog.LogOpts[1].Name)
	glog.InitLogging2(loglevel, logfile, logfilesize, logfilenum)
	log.DefaultLogger = log.DefaultLogger.WithName(appName)

	if err := CallInit(); err != nil {
		log.Fatal().Kv("err", err).Printf("failed to init")
	}
}
