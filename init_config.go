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

	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/go-toolkit/app"
)

func init() {
	app.DefaultApp.SetConfigLoader(loadConfig)
}

func loadConfig(ctx context.Context, app *app.App) (err error) {
	if version := app.Version(); version != "" {
		gconf.SetVersion(version)
	}

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
