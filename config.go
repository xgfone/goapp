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

package goapp

import (
	"fmt"
	"reflect"

	"github.com/xgfone/gconf/v4"
	"github.com/xgfone/klog/v3"
)

func init() {
	gconf.SetErrHandler(gconf.ErrorHandler(func(err error) { klog.Errorf(err.Error()) }))
}

func registerOpt(conf *gconf.Config, options interface{}) {
	if options == nil {
		return
	}

	v := reflect.ValueOf(options)
	switch v.Kind() {
	case reflect.Struct:
		if opt, ok := options.(gconf.Opt); ok {
			conf.RegisterOpt(opt)
		} else {
			panic("the struct must be a pointer")
		}
	case reflect.Ptr:
		conf.RegisterStruct(options)
	case reflect.Slice, reflect.Array:
		for _len := v.Len() - 1; _len >= 0; _len-- {
			registerOpt(conf, v.Index(_len).Interface())
		}
	default:
		panic(fmt.Errorf("unknown option %T", options))
	}
}

// InitConfig initliazlies the configuration options.
//
// iptions may be gconf.Opt, []gconf.Opt, a pointer to the struct variable,
// or the list of the pointers to the struct variables. For example,
//
//    InitConfig("", gconf.StrOpt("optname", "HELP TEXT"))
//    InitConfig("appname", []gconf.Opt{gconf.StrOpt("optname", "HELP TEXT")})
//    InitConfig("", structPtr, "1.0.0")
//
func InitConfig(app string, options interface{}, version ...string) {
	// Set the version
	if len(version) > 0 && version[0] != "" {
		gconf.SetStringVersion(version[0])
	}

	// Add the options
	registerOpt(gconf.Conf, options)

	// Parse the CLI arguments
	gconf.AddAndParseOptFlag(gconf.Conf)

	// Load the configuration from the flag, env and file sources.
	gconf.LoadSource(gconf.NewFlagSource())
	if app != "" {
		gconf.LoadSource(gconf.NewEnvSource(app))
	}
	gconf.LoadSource(gconf.NewFileSource(gconf.GetString(gconf.ConfigFileOpt.Name)))

}
