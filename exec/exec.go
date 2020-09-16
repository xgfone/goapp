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

	"github.com/xgfone/go-tools/v7/execution"
	"github.com/xgfone/goapp/log"
)

func fatalError(name string, args []string, err error) {
	if ce, ok := err.(execution.CmdError); ok {
		fields := make([]log.Field, 2, 5)
		fields[0] = log.F("cmd", ce.Name)
		fields[1] = log.F("args", ce.Args)

		if len(ce.Stdout) != 0 {
			fields = append(fields, log.F("stdout", string(ce.Stdout)))
		}
		if len(ce.Stderr) != 0 {
			fields = append(fields, log.F("stderr", string(ce.Stderr)))
		}
		if ce.Err != nil {
			fields = append(fields, log.E(ce.Err))
		}

		log.Fatal("fail to execute the command", fields...)
	}

	log.Fatal("fail to execute the command", log.F("cmd", name),
		log.F("args", args), log.E(err))
}

// Execute executes a command name with args.
func Execute(name string, args ...string) error {
	return execution.Execute(context.Background(), name, args...)
}

// Output executes a command name with args, and get the stdout.
func Output(name string, args ...string) (string, error) {
	return execution.Output(context.Background(), name, args...)
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

// PanicExecutes is the equal to MustExecute(cmds[0], cmds[1:]...).
func PanicExecutes(cmds ...string) { MustExecute(cmds[0], cmds[1:]...) }

// PanicOutputs is the equal to MustOutput(cmds[0], cmds[1:]...).
func PanicOutputs(cmds ...string) string { return MustOutput(cmds[0], cmds[1:]...) }

//////////////////////////////////////////////////////////////////////////////

// SetDefaultCmdLogHook sets the log hook for the default command executor.
func SetDefaultCmdLogHook() {
	execution.DefaultCmd.AppendResultHooks(LogExecutedCmdResultHook)
}

// LogExecutedCmdResultHook returns a hook to log the executed command.
func LogExecutedCmdResultHook(name string, args []string, stdout, stderr []byte,
	err error) ([]byte, []byte, error) {
	if err == nil {
		log.Info("successfully execute the command", log.F("cmd", name), log.F("args", args))
	} else {
		fields := make([]log.Field, 2, 5)
		if e, ok := err.(execution.CmdError); ok {
			fields[0] = log.F("cmd", e.Name)
			fields[1] = log.F("args", e.Args)

			if len(e.Stdout) != 0 {
				fields = append(fields, log.F("stdout", string(e.Stdout)))
			}
			if len(e.Stderr) != 0 {
				fields = append(fields, log.F("stderr", string(e.Stderr)))
			}
			if e.Err != nil {
				fields = append(fields, log.E(e.Err))
			}
		} else {
			fields[0] = log.F("cmd", name)
			fields[1] = log.F("args", args)
			if len(stdout) != 0 {
				fields = append(fields, log.F("stdout", string(stdout)))
			}
			if len(stderr) != 0 {
				fields = append(fields, log.F("stderr", string(stderr)))
			}
			fields = append(fields, log.E(err))
		}

		log.Error("fail to execute the command", fields...)
	}

	return stdout, stderr, err
}