// Copyright 2023 xgfone
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

// Package bytestr is used to convert the value between string and []byte.
package bytestr

import "unsafe"

// String converts the value b from []byte to string.
func String(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// Bytes converts the value s from string to []byte.
func Bytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
