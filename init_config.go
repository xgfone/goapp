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
	"context"
	"runtime"
	"strings"
	"time"

	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/go-toolkit/app"
)

func init() {
	app.Default().SetConfigLoader(loadConfig)
}

func tryWriteString(buf *strings.Builder, key, value string) {
	if value != "" {
		_ = buf.WriteByte(' ')
		_, _ = buf.WriteString(key)
		_ = buf.WriteByte('=')
		_, _ = buf.WriteString(value)
	}
}

func loadConfig(ctx context.Context, app *app.App) (err error) {
	var builtat string
	if t := app.BuildTime(); !t.IsZero() {
		builtat = app.BuildTime().Format(time.RFC3339)
	}

	var buf strings.Builder
	buf.Grow(140)
	buf.WriteString(app.Name())
	tryWriteString(&buf, "version", app.Version())
	tryWriteString(&buf, "commit", app.Commit())
	tryWriteString(&buf, "builtat", builtat)
	tryWriteString(&buf, "goversion", runtime.Version())
	tryWriteString(&buf, "platform", runtime.GOOS+"/"+runtime.GOARCH)
	gconf.SetVersion(buf.String())

	// Register and Parse the options with flag
	err = gconf.AddAndParseOptFlag(gconf.Conf)
	if err != nil {
		return
	}

	// Load the configs from flag
	err = gconf.LoadSource(gconf.NewFlagSource())
	if err != nil {
		return
	}

	// Load the configs from env
	err = gconf.LoadSource(gconf.NewEnvSource(app.Name()))
	if err != nil {
		return
	}

	// Load the configs from file
	if cfile, _ := gconf.Get(gconf.ConfigFileOpt.Name).(string); cfile != "" {
		err = gconf.LoadAndWatchSource(gconf.NewFileSource(cfile))
		if err != nil {
			return
		}
	}

	return
}
