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
	stdlog "log"

	"github.com/xgfone/gconf/v4"
	"github.com/xgfone/go-tools/v6/lifecycle"
	"github.com/xgfone/klog/v3"
)

// LogOpts collects the options about the log.
var LogOpts = []gconf.Opt{
	gconf.StrOpt("logfile", "The file path of the log. The default is stdout.").As("log_file"),
	gconf.StrOpt("loglevel", "The level of the log, such as debug, info").D("info").As("log_level"),
}

// DatabaseOpts collects the options of the SQL database.
var DatabaseOpts = []gconf.Opt{
	gconf.StrOpt("connection", "The URL connection to the alarm database, user:password@tcp(127.0.0.1:3306)/db").C(false),
	gconf.IntOpt("maxconnnum", "The maximum number of the connections.").C(false).D(100),
}

// InitLogging initializes the logging.
//
// If filepath is empty, it will use Stdout as the writer.
func InitLogging(level, filepath string) {
	log := klog.WithLevel(klog.NameToLevel(level)).WithCtx(klog.Caller("caller"))
	klog.SetDefaultLogger(log)

	writer, err := klog.FileWriter(filepath, "100M", 100)
	if err != nil {
		klog.Error(err.Error())
		lifecycle.Exit(1)
	}

	log.Encoder().SetWriter(writer)
	stdlog.SetOutput(klog.ToIOWriter(writer))
	lifecycle.Register(func() { writer.Close() })
}

// LogPanic wrapps and logs the panic.
func LogPanic(name ...string) {
	if err := recover(); err != nil {
		if len(name) == 0 || name[0] == "" {
			klog.Error("panic", klog.CallerStack("stack"), klog.F("err", err))
		} else {
			klog.Error("panic", klog.F("name", name[0]), klog.CallerStack("stack"), klog.F("err", err))
		}
	}
}
