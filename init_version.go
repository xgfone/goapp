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

	"github.com/xgfone/gover"
)

// Version is the version of app.
//
// Default: github.com/xgfone/gover.Text()
var Version string

func init() {
	if Version == "" {
		Version = gover.Text()
	}
	expvar.NewString("version").Set(Version)
}

func initversion() {
	slog.Info("print version", "version", Version)
}
