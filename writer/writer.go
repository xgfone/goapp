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

package writer

import (
	"io"
	"sync/atomic"
)

type wrapper struct{ io.Writer }

// Switcher is a writer proxy, which can switch the writer to another
// in running.
type Switcher struct{ w atomic.Value }

// NewSwitcher returns a new Switcher with w.
func NewSwitcher(writer io.Writer) *Switcher {
	w := new(Switcher)
	w.Set(writer)
	return w
}

// Write implements the interface io.Writer.
func (w *Switcher) Write(b []byte) (int, error) {
	return w.Get().Write(b)
}

// Close implements the interface io.Closer.
func (w *Switcher) Close() error {
	if c, ok := w.Get().(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// Set sets the writer to new.
func (w *Switcher) Set(new io.Writer) {
	if new == nil {
		panic("Switcher.Set: io.Writer is nil")
	}
	w.w.Store(wrapper{Writer: new})
}

// Get returns the wrapped writer, which is equal to Unwrap.
func (w *Switcher) Get() io.Writer {
	return w.w.Load().(wrapper).Writer
}

// Unwrap returns the wrapped writer.
func (w *Switcher) Unwrap() io.Writer {
	return w.w.Load().(wrapper).Writer
}

// Swap swaps the old writer with the new writer.
func (w *Switcher) Swap(new io.Writer) (old io.Writer) {
	if new == nil {
		panic("Switcher.Swap: io.Writer is nil")
	}
	return w.w.Swap(wrapper{Writer: new}).(wrapper).Writer
}
