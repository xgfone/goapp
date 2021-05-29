// Copyright 2020 xgfone
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

package opentracing

import (
	"fmt"
	"os"
	"plugin"
	"strings"

	"github.com/xgfone/gconf/v5"
	"github.com/xgfone/go-log"
)

// PluginOpts collects the options of the OpenTracing Plugin.
var PluginOpts = []gconf.Opt{
	gconf.StrOpt("path", "The path of the plugin implementing the OpenTracing tracer."),
	gconf.StrOpt("config", "The configuration information of the plugin."),
}

// PluginOptGroup is the group of the OpenTracing plugin config options.
var PluginOptGroup = gconf.NewGroup("opentracing.plugin")

// RegisterPluginOpts registers the options of opentracing plugin.
func RegisterPluginOpts() { PluginOptGroup.RegisterOpts(PluginOpts...) }

func getPluginPathAndConfigFromEnv() (p string, c interface{}) {
	p = PluginOptGroup.GetString("path")
	c = PluginOptGroup.GetString("config")

	for _, env := range os.Environ() {
		if index := strings.IndexByte(env, '='); index > 0 {
			switch key := strings.ToUpper(strings.TrimSpace(env[:index])); key {
			case "OPENTRACING_PLUGIN_PATH":
				p = strings.TrimSpace(env[index+1:])
			case "OPENTRACING_PLUGIN_CONFIG":
				c = strings.TrimSpace(env[index+1:])
			}
		}
	}

	return
}

func getPluginPathAndConfig(p string, c interface{}) (string, interface{}) {
	_p, _c := getPluginPathAndConfigFromEnv()
	if _p != "" {
		p = _p
	}
	if _c != nil {
		c = _c
	}

	return p, c
}

// MustInitOpenTracingFromPlugin is the same as InitOpenTracing,
// but logs the error and exits the program when an error occurs.
func MustInitOpenTracingFromPlugin(pluginPath string, config interface{}) {
	pluginPath, config = getPluginPathAndConfig(pluginPath, config)
	if err := InitOpenTracingFromPlugin(pluginPath, config); err != nil {
		log.Fatal("failed to initialize the opentracing implementation",
			log.F("plugin", pluginPath), log.F("config", config), log.E(err))
	}
}

// InitOpenTracingFromPlugin initializes the OpenTracing implementation, which will load
// the implementation plugin and call the function InitOpenTracing with config.
//
// The plugin must contain the function
//   func InitOpenTracing(config interface{}) error
//
// Notice:
//   1. If config is empty, retry the env variable "OPENTRACING_PLUGIN_CONFIG".
//   2. If pluginPath is empty, retry the env variable "OPENTRACING_PLUGIN_PATH".
//   3. If pluginPath is empty, it returns nil and does nothing.
func InitOpenTracingFromPlugin(pluginPath string, config interface{}) (err error) {
	pluginPath, config = getPluginPathAndConfig(pluginPath, config)
	if pluginPath == "" {
		return
	}

	p, err := plugin.Open(pluginPath)
	if err != nil {
		return
	}

	trf, err := p.Lookup("InitOpenTracing")
	if err != nil {
		return
	}

	if f, ok := trf.(func(interface{}) error); ok {
		return f(config)
	}

	panic(fmt.Errorf("invalid the function InitOpenTracing: %T", trf))
}
