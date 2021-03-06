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

package goapp

import (
	"fmt"

	"github.com/go-stack/stack"
	"github.com/xgfone/go-log"
)

// PanicError is used to represent the panic error.
type PanicError struct {
	Panic interface{}
	Stack stack.CallStack
}

// NewPanicError returns a new PanicError.
func NewPanicError(panic interface{}, depth int) PanicError {
	return PanicError{Panic: panic, Stack: GetCallStack(depth + 1)}
}

func (pe PanicError) Error() string {
	if len(pe.Stack) == 0 {
		return fmt.Sprintf("panic: %v", pe.Panic)
	}
	return fmt.Sprintf("panic: %v, stacks=%+v", pe.Panic, pe.Stack)
}

// GetCallStack returns the stacks of the caller.
func GetCallStack(depth int) stack.CallStack {
	return stack.Trace().TrimBelow(stack.Caller(depth + 2)).TrimRuntime()
}

// Panic panics with the msg if err is not nil. Or do nothing.
func Panic(err error, msg string, args ...interface{}) {
	if err != nil {
		if len(args) != 0 {
			msg = fmt.Sprintf(msg, args...)
		}
		panic(fmt.Errorf("%s: %s", msg, err))
	}
}

// Must logs the error with the msg and the program exits if err is not nil.
// Or do nothing.
func Must(err error, msg string, args ...interface{}) {
	if err != nil {
		if len(args) != 0 {
			msg = fmt.Sprintf(msg, args...)
		}
		log.Fatal(msg, log.E(err))
	}
}
