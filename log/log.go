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
	"io"
	"os"
	"strings"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-atexit"
)

func parseLevel(level string) log.Level {
	switch strings.ToLower(level) {
	case "":
	case "trace":
		return log.LevelTrace
	case "debug":
		return log.LevelDebug
	case "info":
		return log.LevelInfo
	case "warn":
		return log.LevelWarn
	case "error":
		return log.LevelError
	case "fatal":
		return log.LevelFatal
	default:
		fmt.Printf("unknown the level '%s'\n", level)
		atexit.Exit(1)
	}

	return log.LevelInfo
}

// InitLoging initializes the logging configuration.
//
// If logfile is empty, output the log to os.Stderr.
func InitLoging(appName, loglevel, logfile string) {
	level := parseLevel(loglevel)

	var writer io.WriteCloser = os.Stderr
	if logfile != "" {
		file, err := log.NewFileWriter(logfile, "100M", 100)
		if err != nil {
			log.Fatal("fail to new the file log writer", "logfile", logfile, "err", err)
		}

		writer = file
		atexit.OnExitWithPriority(0, func() { file.Close() })
	}

	handler := log.NewJSONHandler(writer, level)
	if appName != "" {
		handler = handler.WithAttrs([]log.Attr{log.String("logger", appName)})
	}

	log.SetDefault(nil, handler)
}
