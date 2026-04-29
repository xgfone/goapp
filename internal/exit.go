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

package internal

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/xgfone/go-toolkit/runtimex"
)

/// ----------------------------------------------------------------------- ///

var (
	init0funcs []func()
	init1funcs []func()
)

// OnInitPre registers a pre-init function called before calling init functions
// when calling RunInit().
func OnInitPre(f func()) {
	init0funcs = append(init0funcs, f)
	_traceregister("init0")
}

// OnInit registers an init function called when calling RunInit().
func OnInit(f func()) {
	init1funcs = append(init1funcs, f)
	_traceregister("init1")
}

// RunInit calls the init functions in turn.
func RunInit() {
	iter(init0funcs, func(f func()) { f() })
	iter(init1funcs, func(f func()) { f() })
}

/// ----------------------------------------------------------------------- ///

var (
	exitfuncs  []func()
	cleanfuncs []func()
	exitonce   sync.Once
	exitedch   = make(chan struct{})

	exitctx, exitcancel = context.WithCancel(context.Background())
)

// OnExitPost registers a function called after calling exit functions.
func OnExitPost(f func()) {
	cleanfuncs = append(cleanfuncs, f)
	_traceregister("exitpost")
}

// OnExit registers a function called when calling RunExit().
func OnExit(f func()) {
	exitfuncs = append(exitfuncs, f)
	_traceregister("exit")
}

// ExitContext returns a context that it will be cancelled when calling RunExit.
func ExitContext() context.Context { return exitctx }

// WaitExit waits until all exit functions finish to be called.
func WaitExit() { <-exitedch }

// RunExit calls the exit functions in reverse turn.
func RunExit() {
	exitonce.Do(exit)
	WaitExit()
}

func exit() {
	exitcancel()
	reverseIter(exitfuncs, runexit)
	reverseIter(cleanfuncs, runexit)
	close(exitedch)
}

func runexit(f func()) {
	defer exitrecover()
	f()
}

func exitrecover() {
	if r := recover(); r != nil {
		slog.Error("exit func panics", "panic", r)
	}
}

func init() { OnExitPost(func() { time.Sleep(time.Millisecond * 10) }) }

func iter[S ~[]E, E any](s S, f func(E)) {
	for _, e := range s {
		f(e)
	}
}

func reverseIter[S ~[]E, E any](s S, f func(E)) {
	for _len := len(s) - 1; _len >= 0; _len-- {
		f(s[_len])
	}
}

func init() {
	runtimex.SetExitFunc(func(code int) {
		RunExit()
		os.Exit(code)
	})
}

/// ----------------------------------------------------------------------- ///

var DEBUG bool

func init() {
	DEBUG, _ = strconv.ParseBool(os.Getenv("DEBUG"))
}

func _traceregister(kind string) {
	if DEBUG {
		frame := runtimex.Caller(3)
		msg := fmt.Sprintf("register %s function", kind)
		slog.Info(msg, "file", frame.File, "line", frame.Line)
	}
}
