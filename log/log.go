// Copyright 2020~2026 xgfone
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

	"github.com/xgfone/go-toolkit/app"
)

func init() {
	app.Default().Finalize("logger", closeWriter)
	handler := NewOptionHandler(NewJSONHandler(Writer, Level))
	handler.ReplaceFunc = replaceAttrForAppName
	SetDefault(handler)
}

func replaceAttrForAppName(c context.Context, r slog.Record) slog.Record {
	r.AddAttrs(slog.String("app", app.Name()))
	return r
}

// SetDefault is used to set default global logger with the handler.
func SetDefault(handler slog.Handler, attrs ...slog.Attr) {
	if len(attrs) > 0 {
		handler = handler.WithAttrs(attrs)
	}

	log.SetFlags(log.Lshortfile | log.Llongfile)
	slog.SetDefault(slog.New(handler))
}

// Init initializes the logging configuration.
//
// If file is empty or equal to "stderr", output the log to os.Stderr.
// If file is equal to "stdout", output the log to os.Stdout.
// Or, output the log to the given file.
func Init(level, file string, logfilenum int) (err error) {
	if err = SetLevel(level); err != nil {
		return
	}

	switch file {
	case "":
		return

	case "stdout":
		err = replaceWriter(os.Stdout)

	case "stderr":
		err = replaceWriter(os.Stderr)

	default:
		err = setfilewriter(file, logfilenum)
	}

	return
}

func setfilewriter(file string, logfilenum int) (err error) {
	if logfilenum <= 0 {
		logfilenum = 100
	}

	_file, err := NewFileWriter(file, "100M", logfilenum)
	if err != nil {
		return
	}

	return replaceWriter(_file)
}

// closeWriter runs after resource cleanup and all exit notifications.
func closeWriter(context.Context) error {
	return replaceWriter(os.Stderr)
}

func replaceWriter(next io.Writer) (err error) {
	old := Writer.Swap(next)
	switch old {
	case os.Stderr, os.Stdout:
	default:
		if closer, ok := old.(io.Closer); ok {
			return closer.Close()
		}
	}

	return
}
