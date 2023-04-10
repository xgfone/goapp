// Copyright 2022 xgfone
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

// Package service provides some convenient service functions.
package service

import (
	"context"
	"time"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/service"
	"github.com/xgfone/go-apiserver/service/task"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-checker"
	"github.com/xgfone/go-wait"
)

var (
	taskservice = service.LogService(log.LevelInfo, "task", task.DefaultService)
	svcchecker  = checker.NewChecker("taskservice", checker.AlwaysTrue(), func(_ string, ok bool) {
		if ok {
			taskservice.Activate()
		} else {
			taskservice.Deactivate()
		}
	})
)

func init() {
	atexit.OnInitWithPriority(1000, func() { go svcchecker.Start(atexit.Context()) })
	atexit.OnExit(svcchecker.Stop)
}

// RunTask runs the task function synchronously if task.DefaultService is activated.
// Or, do nothing.
func RunTask(delay, interval time.Duration, taskFunc func(context.Context)) {
	runner := task.WrapRunner(nil, task.RunnerFunc(taskFunc))
	wait.RunForever(atexit.Context(), delay, interval, runner.Run)
}

// SetCheckCond resets the check condition of the monitor service.
func SetCheckCond(cond checker.Condition) { svcchecker.SetCondition(cond) }

// SetVipCheckCond is a convenient function to set the checker based on vip,
// which is equal to SetCheckCond(checker.NewVipCondition(vip, "")).
func SetVipCheckCond(vip string) { SetCheckCond(checker.NewVipCondition(vip, "")) }
