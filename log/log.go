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

	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-log"
)

// LogOpts collects the options about the log.
var LogOpts = []gconf.Opt{
	gconf.StrOpt("log.file", "The file path of the log. The default is stdout.").As("logfile"),
	gconf.StrOpt("log.level", "The level of the log, such as debug, info").D("info").As("loglevel"),
}

// InitLogging is equal to InitLogging2(level, filepath, "100M", 100).
func InitLogging(level, filepath string) {
	InitLogging2(level, filepath, "100M", 100)
}

// InitLogging2 initializes the logging.
//
// If filepath is empty, it will use Stdout as the writer.
func InitLogging2(level, filepath, filesize string, filenum int) {
	if level != "" {
		log.SetLevel(log.NameToLevel(level))
	}

	if filepath != "" {
		writer := log.FileWriter(filepath, filesize, filenum)
		log.DefalutLogger.Encoder.SetWriter(log.SafeWriter(writer))
		stdlog.SetOutput(log.NewIOWriter(writer, log.LvlTrace))
		atexit.Register(func() { writer.Close() })
	}
}
