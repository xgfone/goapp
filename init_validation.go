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
	"reflect"

	"github.com/xgfone/go-structs"
	"github.com/xgfone/go-toolkit/validation"
)

func init() {
	validation.SetValidateFunc(func(_ context.Context, value any) error {
		if value == nil {
			return nil
		}
		return validate(reflect.ValueOf(value))
	})
}

func validate(vf reflect.Value) (err error) {
	switch vf.Kind() {
	case reflect.Struct:
		err = structs.ReflectValue(vf)

	case reflect.Pointer:
		if !vf.IsNil() {
			err = validate(vf.Elem())
		}

	case reflect.Array, reflect.Slice:
		for i, _len := 0, vf.Len(); i < _len; i++ {
			if err = validate(vf.Index(i)); err != nil {
				return
			}
		}
	}

	return
}
