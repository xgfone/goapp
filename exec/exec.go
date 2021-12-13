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

// Package exec supplies some convenient execution functions.
package exec

import (
	"context"
	"sync"
	"time"

	"github.com/xgfone/go-exec"
	"github.com/xgfone/go-log"
)

func init() { exec.DefaultTimeout = time.Second * 3 }

func fatalError(name string, args []string, err error) {
	ce := err.(exec.Result)
	logger := log.Fatal().Kv("cmd", ce.Name).Kv("args", ce.Args)

	if len(ce.Stdout) != 0 {
		logger.Kv("stdout", string(ce.Stdout))
	}
	if len(ce.Stderr) != 0 {
		logger.Kv("stderr", string(ce.Stderr))
	}
	if ce.Err != nil {
		logger.Kv("err", ce.Err)
	}

	logger.Printf("failed to execute the command")
}

// Execute executes a command name with args.
func Execute(name string, args ...string) error {
	return exec.Execute(context.Background(), name, args...)
}

// Output executes a command name with args, and get the stdout.
func Output(name string, args ...string) (string, error) {
	return exec.Output(context.Background(), name, args...)
}

// Executes is equal to Execute(cmds[0], cmds[1:]...).
func Executes(cmds ...string) error { return Execute(cmds[0], cmds[1:]...) }

// Outputs is equal to Output(cmds[0], cmds[1:]...).
func Outputs(cmds ...string) (string, error) { return Output(cmds[0], cmds[1:]...) }

// MustExecute is the same as Execute, but the program exits if there is an error.
func MustExecute(name string, args ...string) {
	if err := Execute(name, args...); err != nil {
		fatalError(name, args, err)
	}
}

// MustOutput is the same as Execute, but the program exits if there is an error.
func MustOutput(name string, args ...string) string {
	out, err := Output(name, args...)
	if err != nil {
		fatalError(name, args, err)
	}
	return out
}

// MustExecutes is the equal to MustExecute(cmds[0], cmds[1:]...).
func MustExecutes(cmds ...string) { MustExecute(cmds[0], cmds[1:]...) }

// MustOutputs is the equal to MustOutput(cmds[0], cmds[1:]...).
func MustOutputs(cmds ...string) string { return MustOutput(cmds[0], cmds[1:]...) }

// PanicExecute is the same as MustExecute, but panic instead of exiting.
func PanicExecute(name string, args ...string) {
	if err := Execute(name, args...); err != nil {
		panic(err)
	}
}

// PanicOutput is the same as MustOutput, but panic instead of exiting.
func PanicOutput(name string, args ...string) string {
	out, err := Output(name, args...)
	if err != nil {
		panic(err)
	}
	return out
}

// PanicExecutes is the equal to PanicExecute(cmds[0], cmds[1:]...).
func PanicExecutes(cmds ...string) { PanicExecute(cmds[0], cmds[1:]...) }

// PanicOutputs is the equal to PanicOutput(cmds[0], cmds[1:]...).
func PanicOutputs(cmds ...string) string { return PanicOutput(cmds[0], cmds[1:]...) }

//////////////////////////////////////////////////////////////////////////////

// SetDefaultCmdLock sets the lock of the default command executor to lock.
func SetDefaultCmdLock(lock *sync.Mutex) { exec.DefaultCmd.Lock = lock }

// SetDefaultCmdLogHook sets the log hook for the default command executor.
func SetDefaultCmdLogHook() { exec.DefaultCmd.ResultHook = LogExecutedCmdResultHook }

// LogExecutedCmdResultHook returns a hook to log the executed command.
func LogExecutedCmdResultHook(r exec.Result) {
	if r.Err == nil {
		log.Info().Kv("cmd", r.Name).Kv("args", r.Args).
			Printf("successfully execute the command")
		return
	}

	logger := log.Error()
	if e, ok := r.Err.(exec.Result); ok {
		logger.Kv("cmd", e.Name).Kv("args", e.Args)
		if len(e.Stdout) != 0 {
			logger.Kv("stdout", string(e.Stdout))
		}
		if len(e.Stderr) != 0 {
			logger.Kv("stderr", string(e.Stderr))
		}
		if e.Err != nil {
			logger.Kv("err", e.Err)
		}
	} else {
		logger.Kv("cmd", r.Name).Kv("args", r.Args)
		if len(r.Stdout) != 0 {
			logger.Kv("stdout", string(r.Stdout))
		}
		if len(r.Stderr) != 0 {
			logger.Kv("stderr", string(r.Stderr))
		}
		logger.Kv("err", r.Err)
	}

	logger.Printf("failed to execute the command")
}
