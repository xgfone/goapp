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
	"context"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-defaults"
)

// Attrs is the global log attributes appended into the log context
// when setting the global default logger.
var Attrs []slog.Attr

func init() {
	SetDefault(NewJSONHandler(Writer, Level))
	defaults.HandlePanicFunc.Set(func(r any) { logpanic(r, 5) })
}

func logpanic(r any, skip int) {
	stacks := defaults.GetStacks(skip)
	slog.Error("wrap a panic", "panic", r, "stacks", stacks)
}

// SetDefault is used to set default global logger with the handler.
func SetDefault(handler slog.Handler, attrs ...slog.Attr) {
	attrs = append(attrs, Attrs...)
	if len(attrs) > 0 {
		handler = handler.WithAttrs(attrs)
	}

	log.SetFlags(log.Lshortfile | log.Llongfile)
	slog.SetDefault(slog.New(handler))
}

// Trace emits a TRACE log message.
func Trace(msg string, args ...any) {
	slog.Log(context.Background(), LevelTrace, msg, args...)
}

// Fatal emits a FATAL log message.
func Fatal(msg string, args ...any) {
	slog.Log(context.Background(), LevelFatal, msg, args...)
	atexit.Exit(1)
}

// InitLoging initializes the logging configuration.
//
// If file is empty, output the log to os.Stderr.
func InitLoging(level, file string, logfilenum int) {
	if err := SetLevel(level); err != nil {
		Fatal("fail to set the log level", "level", level, "err", err)
	}

	if file == "" {
		return
	}

	if logfilenum <= 0 {
		logfilenum = 100
	}

	_file, err := NewFileWriter(file, "100M", logfilenum)
	if err != nil {
		Fatal("fail to new the file log writer", "file", file, "err", err)
	}

	atexit.OnExitWithPriority(0, func() { _file.Close() })
	switch old := Writer.Swap(_file); old {
	case os.Stderr, os.Stdout:
	default:
		if c, ok := old.(io.Closer); ok {
			c.Close()
		}
	}
}
