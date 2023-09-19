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

type writerWrapper struct{ io.Writer }

// SwitchWriter is a writer proxy, which can switch the writer to another
// in running.
type SwitchWriter struct{ w atomic.Value }

// NewSwitchWriter returns a new SwitchWriter with w.
func NewSwitchWriter(writer io.Writer) *SwitchWriter {
	w := new(SwitchWriter)
	w.Set(writer)
	return w
}

// Write implements the interface io.Writer.
func (w *SwitchWriter) Write(b []byte) (int, error) {
	return w.Get().Write(b)
}

// Close implements the interface io.Closer.
func (w *SwitchWriter) Close() error {
	if c, ok := w.Get().(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// Set sets the writer to new.
func (w *SwitchWriter) Set(new io.Writer) {
	if new == nil {
		panic("SwitchWriter.Set: io.Writer is nil")
	}
	w.w.Store(writerWrapper{Writer: new})
}

// Get returns the wrapped writer, which is equal to Unwrap.
func (w *SwitchWriter) Get() io.Writer {
	return w.w.Load().(writerWrapper).Writer
}

// Unwrap returns the wrapped writer.
func (w *SwitchWriter) Unwrap() io.Writer {
	return w.w.Load().(writerWrapper).Writer
}

// Swap swaps the old writer with the new writer.
func (w *SwitchWriter) Swap(new io.Writer) (old io.Writer) {
	if new == nil {
		panic("SwitchWriter.Swap: io.Writer is nil")
	}
	return w.w.Swap(writerWrapper{Writer: new}).(writerWrapper).Writer
}
