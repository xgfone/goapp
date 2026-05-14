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
	"os"
	"path/filepath"

	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/go-toolkit/app"
)

func getdefaultpwd() string {
	if pwd := os.Getenv("WorkingDirectory"); pwd != "" {
		return pwd
	}

	configfile := gconf.GetString(gconf.ConfigFileOpt.Name)
	if configfile != "" {
		return filepath.Dir(configfile)
	}

	return filepath.Dir(os.Args[0])
}

func init() {
	pwd := getdefaultpwd()
	if pwd == "." {
		return
	}

	_pwd, err := filepath.Abs(pwd)
	if err != nil {
		slog.Error("fail to get the absolute path", "pwd", pwd, "err", err)
		return
	}

	if curpwd, err := os.Getwd(); err != nil {
		slog.Error("fail to get the current working directory", "err", err)
	} else if curpwd != _pwd {
		if err := os.Chdir(_pwd); err != nil {
			slog.Error("fail to change the current working directory", "pwd", _pwd, "err", err)
			return
		}
	}

	expvar.Publish("pwd", expvar.Func(func() any {
		pwd, _ := os.Getwd()
		return pwd
	}))
}

func init() {
	app.StageReady.On(func(_ context.Context, app *app.App) error {
		if pwd, err := os.Getwd(); err != nil {
			slog.Error("fail to get current working directory", "err", err)
		} else {
			slog.Debug("current working directory", "pwd", pwd)
		}
		return nil
	})
}
