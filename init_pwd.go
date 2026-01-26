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
	"expvar"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/xgfone/gconf/v6"
)

// PWD is the current working directory.
var PWD string

func trysetpwd() {
	if PWD == "" {
		PWD = os.Getenv("WorkingDirectory")
	}

	if PWD == "" {
		configfile := gconf.GetString(gconf.ConfigFileOpt.Name)
		if configfile != "" {
			PWD = filepath.Dir(configfile)
		} else {
			PWD = filepath.Dir(os.Args[0])
		}
	}

	if PWD == "." {
		slog.Debug("log the current working directory", "pwd", PWD)
		return
	}

	PWD, err := filepath.Abs(PWD)
	if err != nil {
		slog.Error("fail to get the absolute path", "pwd", PWD, "err", err)
		return
	}

	if err := os.Chdir(PWD); err != nil {
		slog.Error("fail to change the current working directory", "pwd", PWD, "err", err)
	} else {
		slog.Debug("change the current working directory", "pwd", PWD)
		expvar.NewString("pwd").Set(PWD)
	}
}
