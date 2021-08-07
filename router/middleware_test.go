// Copyright 2021 xgfone
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

package router

import (
	"testing"

	"github.com/xgfone/goapp"
	"github.com/xgfone/ship/v5"
)

func TestRecover(t *testing.T) {
	err := Recover(ship.NothingHandler())(nil)
	if err != nil {
		t.Error(err)
	}

	panicStack := "middleware_test.go:31"
	err = Recover(func(ctx *ship.Context) (err error) { panic("testpanic") })(nil)
	switch e := err.(type) {
	case nil:
		t.Error(err)
	case goapp.PanicError:
		if e.Panic.(string) != "testpanic" {
			t.Error(e.Panic)
		} else if _len := len(e.Stack); _len == 0 {
			t.Error(e.Stack)
		} else if stack := e.Stack[_len-1].String(); stack != panicStack {
			t.Errorf("expect '%s', got '%s'", panicStack, stack)
		}
	default:
		t.Error(err)
	}
}
