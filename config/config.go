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

import "github.com/xgfone/gconf/v6"

// InitConfig initializes the configuration, which will set the version,
// register the options, parse the CLI arguments with "flag",
// load the "flag", "env" and "file" configuration sources.
func InitConfig(app, version string, opts ...gconf.Opt) {
	gconf.SetVersion(version)
	gconf.RegisterOpts(opts...)
	gconf.AddAndParseOptFlag(gconf.Conf)
	gconf.LoadSource(gconf.NewFlagSource())
	gconf.LoadSource(gconf.NewEnvSource(app))
	configFile := gconf.GetString(gconf.ConfigFileOpt.Name)
	gconf.LoadAndWatchSource(gconf.NewFileSource(configFile))
}
