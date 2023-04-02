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
	"github.com/xgfone/go-wait"
)

// TaskService is the default task service.
var TaskService = task.DefaultService

// Monitor is the default service monitor.
var Monitor = service.NewMonitor(
	service.LogService(log.LevelInfo, "task", TaskService),
	service.NothingChecker(),
	nil)

func init() {
	atexit.OnInit(Monitor.Activate)
	atexit.OnExit(Monitor.Deactivate)
}

// RunTask runs the task function synchronously if TaskService is activated.
// Or, do nothing.
func RunTask(delay, interval time.Duration, taskFunc func(context.Context)) {
	runner := task.WrapRunner(TaskService, task.RunnerFunc(taskFunc))
	wait.RunForever(atexit.Context(), delay, interval, runner.Run)
}

// SetChecker resets the checker of the monitor service.
func SetChecker(checker service.Checker) { Monitor.SetChecker(checker) }

// SetVipChecker is a convenient function to set the checker based on vip,
// which is equal to SetChecker(service.NewVipChecker(vip, "")).
func SetVipChecker(vip string) { SetChecker(service.NewVipChecker(vip, "")) }
