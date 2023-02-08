// Copyright 2020~2022 xgfone
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

// Package exec provides some convenient execution functions.
package exec

import (
	"sync"
	"time"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-exec"
)

func init() {
	exec.DefaultTimeout = time.Second * 3
	exec.DefaultCmd.Lock = new(sync.Mutex)
	exec.DefaultCmd.ResultHook = LogExecutedCmdResultHook
}

// LogExecutedCmdResultHook is a hook to log the executed command.
func LogExecutedCmdResultHook(r exec.Result) {
	if r.Err == nil {
		log.Info("successfully execute the command", "cmd", r.Name, "args", r.Args)
	} else {
		log.Error("fail to execute the command", "cmd", r.Name, "args", r.Args,
			"stdout", r.Stdout, "stderr", r.Stderr, "err", r.Err)
	}
}
