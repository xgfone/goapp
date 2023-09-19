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

package log

import (
	"fmt"
	"log/slog"
	"strings"
)

// Define some extra levels.
const (
	LevelTrace = slog.LevelDebug - 4
	LevelFatal = slog.LevelError + 4
)

// Level is the global log level.
var Level = new(slog.LevelVar)

// SetLevel resets the log level.
func SetLevel(level string) error {
	lvl, err := parseLevel(level)
	if err == nil {
		Level.Set(lvl)
	}
	return err
}

func parseLevel(lvl string) (level slog.Level, err error) {
	switch strings.ToLower(lvl) {
	case "":
		level = slog.LevelInfo
	case "trace":
		level = LevelTrace
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	case "fatal":
		level = LevelFatal
	default:
		err = fmt.Errorf("unknown level '%s'", level)
	}
	return
}
