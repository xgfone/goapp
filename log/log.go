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
	stdlog "log"
	"time"

	apilog "github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-log"
	"github.com/xgfone/go-log/writer"
)

// InitLoging initializes the logging configuration.
//
// If logfile is empty, output the log to os.Stderr.
func InitLoging(appName, loglevel, logfile string) {
	if loglevel != "" {
		log.SetLevel(log.ParseLevel(loglevel))
	}

	if logfile != "" {
		file := log.FileWriter(logfile, "100M", 100)
		fwriter := writer.SafeWriter(writer.BufferWriter(file, 0))

		log.SetWriter(fwriter)
		atexit.OnExitWithPriority(0, func() { fwriter.Close() })
		go loopFlushWriter(fwriter.(writer.Flusher), 0)
	}

	if appName != "" {
		log.DefaultLogger = log.DefaultLogger.WithName(appName)
	}

	apilog.DefaultLogger = log.DefaultLogger
	stdlog.SetOutput(log.DefaultLogger.WithDepth(2))
}

func loopFlushWriter(f writer.Flusher, interval time.Duration) {
	if interval <= 0 {
		interval = time.Second * 10
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-atexit.Done():
			return
		case <-ticker.C:
			f.Flush()
		}
	}
}
