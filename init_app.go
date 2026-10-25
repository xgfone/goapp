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
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/xgfone/go-toolkit/app"
	"github.com/xgfone/go-toolkit/runtimex"
)

var defaultRunner = appRunner{app: app.Default()}

// Run runs the default application through its complete shutdown lifecycle.
func Run() { _ = defaultRunner.run(runtimex.ExitContext()) }

type appRunner struct {
	app           *app.App
	exit          func(int)
	mu            sync.Mutex
	running       bool
	exiting       bool
	exitRequested bool
	exitCode      int
}

func (r *appRunner) run(ctx context.Context) error {
	r.mu.Lock()
	if r.running || r.exiting {
		r.mu.Unlock()
		return errors.New("goapp: Run is already running")
	}
	r.running = true
	r.mu.Unlock()

	err := r.app.Run(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to run app: %v\n", err)
	}

	r.mu.Lock()
	r.running = false
	requested := r.exitRequested
	code := r.exitCode
	if requested {
		r.exiting = true
	}
	r.mu.Unlock()

	if requested {
		r.finishExit(code)
	}

	return err
}

// During Run, Exit requests stopping and returns so lifecycle callbacks cannot
// deadlock waiting for themselves. The Run owner performs the requested process
// exit after rollback/draining, cleanup, exit callbacks, and logger finalization.
// Callers should return from their callback after requesting exit. Before and
// after Run, Exit delegates immediately to the original exit function.
func (r *appRunner) requestExit(code int) {
	r.mu.Lock()
	if r.exiting {
		r.mu.Unlock()
		return
	}
	if !r.running {
		r.exiting = true
		r.mu.Unlock()
		r.finishExit(code)
		return
	}
	if !r.exitRequested {
		r.exitRequested = true
		r.exitCode = code
	}
	r.mu.Unlock()
	r.app.Stop()
}

func (r *appRunner) finishExit(code int) {
	defer func() {
		r.mu.Lock()
		r.exiting = false
		r.mu.Unlock()
	}()
	r.exit(code)
}

func init() {
	defaultRunner.exit = runtimex.GetExitFunc()
	runtimex.SetExitFunc(defaultRunner.requestExit)
}
