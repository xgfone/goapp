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

// Package log is used to initialize the logging.
package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/xgfone/go-apiserver/io2"
	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-atexit"
)

var level = new(log.LevelVar)

func init() { log.SetDefault(log.NewJSONHandler(log.Writer, level)) }

func parseLevel(lvl string) (level log.Level, err error) {
	switch strings.ToLower(lvl) {
	case "":
		level = log.LevelInfo
	case "trace":
		level = log.LevelTrace
	case "debug":
		level = log.LevelDebug
	case "info":
		level = log.LevelInfo
	case "warn":
		level = log.LevelWarn
	case "error":
		level = log.LevelError
	case "fatal":
		level = log.LevelFatal
	default:
		err = fmt.Errorf("unknown level '%s'", level)
	}
	return
}

// SetLevel resets the log level.
func SetLevel(loglevel string) error {
	lvl, err := parseLevel(loglevel)
	if err == nil {
		level.Set(lvl)
	}
	return err
}

// InitLoging initializes the logging configuration.
//
// If logfile is empty, output the log to os.Stderr.
func InitLoging(appName, loglevel, logfile string) {
	if lvl, err := parseLevel(loglevel); err != nil {
		log.Fatal("fail to parse log level", "err", err)
	} else {
		level.Set(lvl)
	}

	if logfile != "" {
		file, err := io2.NewFileWriter(logfile, "100M", 100)
		if err != nil {
			log.Fatal("fail to new the file log writer", "logfile", logfile, "err", err)
		}

		atexit.OnExitWithPriority(0, func() { file.Close() })
		switch old := log.Writer.Swap(file); old {
		case os.Stderr, os.Stdout:
		default:
			io2.Close(old)
		}
	}
}
