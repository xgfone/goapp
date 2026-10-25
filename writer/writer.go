// Copyright 2023~2026 xgfone
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

// Switcher proxies writes to an atomically replaceable writer.
//
// Atomic access protects only the writer reference. It does not serialize
// underlying Write and Close calls or wait for writes using a previous writer.
// Callers must coordinate writer lifetimes and any required I/O synchronization.
//
// With a single slog.JSONHandler or slog.TextHandler and its derived handlers,
// writes are serialized by that handler. Initialize logging before concurrent
// use, and close its writer only after all log users have stopped. Concurrent
// replacement followed by closing the old writer requires separate synchronization.
//
// Initialize a Switcher with NewSwitcher or Set before reading or using it.
// A Switcher must not be copied after its first use.
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

// Close closes the currently selected writer if it implements io.Closer.
//
// It does not wait for writes or prevent future writes; callers must first
// stop all users of that writer.
func (w *Switcher) Close() error {
	if c, ok := w.Get().(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// Set atomically selects new. It neither closes the previous writer nor waits
// for operations using it.
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

// Swap atomically selects new and returns the previous writer.
//
// It does not wait for operations using the returned writer.
// The caller must ensure those operations finish before closing it.
func (w *Switcher) Swap(new io.Writer) (old io.Writer) {
	if new == nil {
		panic("Switcher.Swap: io.Writer is nil")
	}
	return w.w.Swap(wrapper{Writer: new}).(wrapper).Writer
}
