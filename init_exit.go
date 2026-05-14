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

	"github.com/xgfone/go-toolkit/app"
)

// Deprecated.
func OnInit(f func()) {
	app.StageStart.On(callback(f))
}

// Deprecated.
func OnInitPre(f func()) {
	app.StageInit.On(callback(f))
}

// Deprecated.
func OnExit(f func()) {
	app.StageCleanup.On(func(context.Context, *app.App) error {
		f()
		return nil
	})
}

// Deprecated.
func OnExitPost(f func()) {
	app.StageExited.On(callback(f))
}

func callback(f func()) func(context.Context, *app.App) error {
	return func(context.Context, *app.App) error {
		f()
		return nil
	}
}
