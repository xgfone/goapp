// Copyright 2022 xgfone
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

// Package config is used to configure the application.
package config

import (
	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/gover"
)

// ParseAndLoadSource is used to parse and load the options from some sources.
//
// Default: use "flag" to parse and load the options from the CLI arguments.
var ParseAndLoadSource func(app string)

func init() {
	ParseAndLoadSource = func(app string) {
		_ = gconf.AddAndParseOptFlag(gconf.Conf)
		_ = gconf.LoadSource(gconf.NewFlagSource())
	}
}

// InitConfig initializes the configuration, which will set the version,
// register the options, parse the CLI arguments with "flag",
// load the "flag", "env" and "file" configuration sources.
func InitConfig(app, version string, opts ...gconf.Opt) {
	if version == "" {
		version = gover.Text()
	}

	gconf.SetVersion(version)
	gconf.RegisterOpts(opts...)
	ParseAndLoadSource(app)
	_ = gconf.LoadSource(gconf.NewEnvSource(app))
	if cfile, _ := gconf.Get(gconf.ConfigFileOpt.Name).(string); cfile != "" {
		_ = gconf.LoadAndWatchSource(gconf.NewFileSource(cfile))
	}
}
