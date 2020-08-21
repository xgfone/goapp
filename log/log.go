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

// Package log is used to initialize the logger, and supplies some assistant
// functions about log.
package log

import (
	stdlog "log"

	"github.com/xgfone/gconf/v5"
	"github.com/xgfone/go-tools/v7/lifecycle"
	"github.com/xgfone/klog/v3"
)

// LogOpts collects the options about the log.
var LogOpts = []gconf.Opt{
	gconf.StrOpt("logfile", "The file path of the log. The default is stdout.").As("log_file"),
	gconf.StrOpt("loglevel", "The level of the log, such as debug, info").D("info").As("log_level"),
}

// Predefine some log functions.
//
// Please refer to https://godoc.org/github.com/xgfone/klog
var (
	E      = klog.E
	F      = klog.F
	Ef     = klog.Ef
	Tracef = klog.Tracef
	Debugf = klog.Debugf
	Infof  = klog.Infof
	Warnf  = klog.Warnf
	Errorf = klog.Errorf
	Printf = klog.Printf
	Panicf = klog.Panicf
	Fatalf = klog.Fatalf

	Trace = klog.Trace
	Debug = klog.Debug
	Info  = klog.Info
	Warn  = klog.Warn
	Error = klog.Error
	Panic = klog.Panic
	Fatal = klog.Fatal

	RegisterCallOnExit = klog.RegisterCallOnExit
)

func init() {
	RegisterCallOnExit(lifecycle.Stop)
}

// InitLogging is equal to InitLogging2(level, filepath, "100M", 100).
func InitLogging(level, filepath string) {
	InitLogging2(level, filepath, "100M", 100)
}

// InitLogging2 initializes the logging.
//
// If filepath is empty, it will use Stdout as the writer.
func InitLogging2(level, filepath, filesize string, filenum int) {
	log := klog.WithLevel(klog.NameToLevel(level)).WithCtx(klog.Caller("caller"))
	klog.SetDefaultLogger(log)

	writer, err := klog.FileWriter(filepath, filesize, filenum)
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