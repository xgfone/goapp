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

package validate

import (
	"fmt"
)

func ExampleValidate() {
	type S struct {
		F1 int64  `validate:"min(100)"` // General Type
		F2 *int64 `validate:"max(200)"` // Pointer Type
		F3 []struct {
			F4 int64 `validate:"min(1)"`
			F5 int64
		}
	}

	var v S
	fmt.Println(Validate(v))
	fmt.Println(Validate(&v))

	v.F1 = 123
	v.F2 = new(int64)
	*v.F2 = 123
	v.F3 = []struct {
		F4 int64 `validate:"min(1)"`
		F5 int64
	}{
		{F4: 1},
		{F4: 2},
	}
	fmt.Println(Validate(v))
	fmt.Println(Validate(&v))

	// Output:
	// F1: the integer is less than 100
	// F1: the integer is less than 100
	// <nil>
	// <nil>
}
