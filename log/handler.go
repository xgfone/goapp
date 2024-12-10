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

package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/xgfone/go-toolkit/runtimex"
)

// NewJSONHandler returns a new json handler.
//
// If w is nil, use Writer instead.
func NewJSONHandler(w io.Writer, level slog.Leveler) slog.Handler {
	if w == nil {
		w = os.Stderr
	}

	return slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:       level,
		AddSource:   true,
		ReplaceAttr: replaceSourceAttr,
	})
}

func replaceSourceAttr(groups []string, a slog.Attr) slog.Attr {
	switch {
	case a.Key == slog.SourceKey:
		if src, ok := a.Value.Any().(*slog.Source); ok {
			a.Value = slog.StringValue(fmt.Sprintf("%s:%d", runtimex.TrimPkgFile(src.File), src.Line))
		}

	case a.Key == slog.LevelKey:
		if lvl, ok := a.Value.Any().(slog.Level); ok {
			switch lvl {
			case LevelTrace:
				a.Value = slog.StringValue("TRACE")
			case LevelFatal:
				a.Value = slog.StringValue("FATAL")
			}
		}

	case a.Value.Kind() == slog.KindDuration:
		a.Value = slog.StringValue(a.Value.Duration().String())
	}

	return a
}

type OptionHandler struct {
	slog.Handler

	// Options
	EnableFunc  func(context.Context, slog.Level) bool         // Default: nil
	FilterFunc  func(context.Context, slog.Record) bool        // Default: nil
	ReplaceFunc func(context.Context, slog.Record) slog.Record // Default: nil
}

// NewOptionHandler returns a new OptionHandler wrapping the given handler.
func NewOptionHandler(handler slog.Handler) *OptionHandler {
	return &OptionHandler{Handler: handler}
}

func (h *OptionHandler) clone() *OptionHandler {
	nh := *h
	return &nh
}

// Unwrap returns the inner wrapped slog handler.
func (h *OptionHandler) Unwrap() slog.Handler { return h.Handler }

// Enabled implements the interface Handler#Enabled.
func (h *OptionHandler) Enabled(c context.Context, l slog.Level) bool {
	if h.EnableFunc != nil {
		return h.EnableFunc(c, l)
	}
	return h.Handler.Enabled(c, l)
}

// Handle implements the interface Handler#Handle.
func (h *OptionHandler) Handle(c context.Context, r slog.Record) error {
	if h.ReplaceFunc != nil {
		r = h.ReplaceFunc(c, r)
	}
	if h.FilterFunc != nil && h.FilterFunc(c, r) {
		return nil
	}
	return h.Handler.Handle(c, r)
}

// WithAttrs implements the interface Handler#WithAttrs.
func (h *OptionHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nh := h.clone()
	nh.Handler = h.Handler.WithAttrs(attrs)
	return nh
}

// WithGroup implements the interface Handler#WithGroup.
func (h *OptionHandler) WithGroup(name string) slog.Handler {
	nh := h.clone()
	nh.Handler = h.Handler.WithGroup(name)
	return nh
}
