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
	"testing"

	"github.com/xgfone/gconf/v5"
)

func TestRegisterOpt(t *testing.T) {
	conf := gconf.New()
	registerOpt(conf, gconf.StrOpt("opt", ""))
	for _, group := range conf.AllGroups() {
		for _, opt := range group.AllOpts() {
			if opt.Name != "opt" {
				t.Errorf("unknow option '%s'", opt.Name)
			}
		}
	}

	conf = gconf.New()
	registerOpt(conf, []gconf.Opt{gconf.StrOpt("opt1", ""), gconf.IntOpt("opt2", "")})
	for _, group := range conf.AllGroups() {
		for _, opt := range group.AllOpts() {
			switch opt.Name {
			case "opt1", "opt2":
			default:
				t.Errorf("unknow option '%s'", opt.Name)
			}
		}
	}

	type st struct {
		Opt1 string
		Opt2 int
	}
	var v st
	conf = gconf.New()
	registerOpt(conf, &v)
	for _, group := range conf.AllGroups() {
		for _, opt := range group.AllOpts() {
			switch opt.Name {
			case "opt1", "opt2":
			default:
				t.Errorf("unknow option '%s'", opt.Name)
			}
		}
	}

}
