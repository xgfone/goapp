// Copyright 2026 xgfone
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
	"log/slog"

	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/goapp/log"
)

var (
	loglevel = gconf.StrOpt("log.level", "The level of the log, such as trace, debug, info, warn, error, etc.").
			As("loglevel").D("info").U(updateLogLevel)
	logfile0 = gconf.StrOpt("log.file", "The file path of the log. The default is stderr.").
			As("logfile")
	logfilenum = gconf.IntOpt("log.filenum", "The number of the log files.").D(100)
)

func updateLogLevel(old, new any) {
	if err := log.SetLevel(new.(string)); err != nil {
		slog.Error("update the log level", "old", old, "new", new, "err", err)
	} else {
		slog.Info("update the log level", "old", old, "new", new)
	}
}

func initlog() {
	logfile := gconf.GetString(logfile0.Name)
	loglevel := gconf.GetString(loglevel.Name)
	logfilenum := gconf.GetInt(logfilenum.Name)
	log.Init(loglevel, logfile, logfilenum)
}
