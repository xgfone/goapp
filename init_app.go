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
	"os"
	"path/filepath"
	"strings"

	"github.com/xgfone/goapp/log"
)

// AppName is the name of app.
//
// Default: filepath.Base(os.Args[0]), but not contain the suffix ".exe"
var AppName string

func init() {
	if len(os.Args) > 0 {
		AppName = filepath.Base(os.Args[0])
		AppName = strings.TrimSuffix(AppName, ".exe")
		log.SetAppName(AppName)
	}
}
