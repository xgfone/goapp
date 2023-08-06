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

// Package services provides some convenient service functions.
package services

import (
	"context"
	"time"

	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/service"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-checker"
	"github.com/xgfone/go-wait"
)

var (
	taskservice = service.LogService(log.LevelInfo, "task", service.DefaultProxy)
	allservices = service.Services{taskservice}
	svcchecker  = checker.NewChecker("services", checker.AlwaysTrue(), func(_ string, ok bool) {
		if ok {
			allservices.Activate()
		} else {
			allservices.Deactivate()
		}
	})
)

func init() {
	atexit.OnInitWithPriority(10000, func() { go svcchecker.Start(atexit.Context()) })
	atexit.OnExit(svcchecker.Stop)
}

// RunTaskOrElse runs the task function synchronously if service.DefaultProxy is activated.
// Or, run the elsef function.
func RunTaskOrElse(delay, interval time.Duration, taskf func(context.Context), elsef func()) {
	wait.RunForever(atexit.Context(), delay, interval, func(context.Context) {
		service.DefaultProxy.RunFunc(taskf, elsef)
	})
}

// RunTask is equal to RunTask(delay, interval, taskf, nil).
func RunTask(delay, interval time.Duration, taskf func(context.Context)) {
	RunTaskOrElse(delay, interval, taskf, nil)
}

// RunTaskAlways always runs the task function synchronously and periodically.
func RunTaskAlways(delay, interval time.Duration, task func(context.Context)) {
	wait.RunForever(atexit.Context(), delay, interval, task)
}

// Append appends the new services.
func Append(services ...service.Service) { allservices = allservices.Append(services...) }

// SetCheckCond resets the check condition of the monitor service.
func SetCheckCond(cond checker.Condition) { svcchecker.SetCondition(cond) }

// SetVipCheckCond is a convenient function to set the checker based on vip,
// which is equal to SetCheckCond(checker.NewVipCondition(vip, "")).
func SetVipCheckCond(vip string) { SetCheckCond(checker.NewVipCondition(vip, "")) }
