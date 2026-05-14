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
	"expvar"
	"log/slog"

	"github.com/xgfone/go-toolkit/app"
	"github.com/xgfone/gover"
)

func init() {
	if version := gover.Text(); version != "" && app.Version() == "0.0.0" {
		app.DefaultApp.SetVersion(version)
	}
}

func init() {
	expvar.Publish("version", expvar.Func(func() any { return app.Version() }))
	app.StageReady.On(func(_ context.Context, app *app.App) error {
		slog.Info("print version", "version", app.Version())
		return nil
	})
}
