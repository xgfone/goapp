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

package internal

import (
	"os"
	"os/signal"
)

var exitsignals = []os.Signal{os.Interrupt}

// SignalForExit watches the exit signals and calls the Exit function
// when any exit signal occurs.
func SignalForExit() {
	ch := make(chan os.Signal, 1)
	defer signal.Stop(ch)

	signal.Notify(ch, exitsignals...)
	<-ch

	RunExit()
}
