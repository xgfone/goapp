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

// Package config is used to initialize the configuration.
package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/xgfone/gconf/v5"
	"github.com/xgfone/goapp/log"
)

func init() {
	gconf.SetErrHandler(gconf.ErrorHandler(func(err error) { log.Errorf(err.Error()) }))
}

func registerOpt(conf *gconf.Config, options interface{}) {
	if options == nil {
		return
	}

	v := reflect.ValueOf(options)
	switch v.Kind() {
	case reflect.Struct:
		if opt, ok := options.(gconf.Opt); ok {
			conf.RegisterOpts(opt)
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
// options may be gconf.Opt, []gconf.Opt, a pointer to the struct variable,
// or the list of the pointers to the struct variables. For example,
//
//    InitConfig("", gconf.StrOpt("optname", "HELP TEXT"))
//    InitConfig("appname", []gconf.Opt{gconf.StrOpt("optname", "HELP TEXT")})
//    InitConfig("", []interface{}{structPtr1, structPtr2}, "1.0.0")
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

// ConvertOptsToCliFlags the config group to []cli.Flag. For example,
//
//   ConvertOptsToCliFlags()                   // Convert the options in the DEFAULT group
//   ConvertOptsToCliFlags("group1")           // Convert the options in the group named "group1"
//   ConvertOptsToCliFlags("group1.group2")    // Convert the options in the group named "group1.group2"
//   ConvertOptsToCliFlags("group1", "group2") // The same as the last.
//
func ConvertOptsToCliFlags(groups ...string) []cli.Flag {
	group := gconf.Conf.OptGroup
	for _, gname := range groups {
		group = group.MustGroup(gname)
	}
	return gconf.ConvertOptsToCliFlags(group)
}

// LoadCliSource loads the config into the groups from the CLI source.
func LoadCliSource(ctx *cli.Context, groups ...string) {
	if len(groups) > 0 {
		groups = strings.Split(strings.Join(groups, "."), ".")
	}

	glen := len(groups)
	ctxs := ctx.Lineage()
	for i := len(ctxs) - 2; i >= 0; i-- {
		gconf.LoadSource(gconf.NewCliSource(ctxs[i], groups[:glen-i]...))
	}
}
