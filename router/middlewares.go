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

package router

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/xgfone/go-log"
	"github.com/xgfone/goapp"
	"github.com/xgfone/ship/v5"
)

// Middleware is the type alias of ship.Middleware.
type Middleware = ship.Middleware

type reqctx uint8

// GetContext returns the http reqeust context from the context.
func getContext(ctx context.Context) *ship.Context {
	c, _ := ctx.Value(reqctx(255)).(*ship.Context)
	return c
}

// SetContext sets the http request context into the context.
func setContext(ctx context.Context, c *ship.Context) (newctx context.Context) {
	return context.WithValue(ctx, reqctx(255), c)
}

type mwHandler ship.Handler

func (h mwHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.HandleHTTP(w, r)
}

func (h mwHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) error {
	c := getContext(r.Context())
	if r != c.Request() {
		c.SetRequest(r)
	}
	if resp, ok := w.(*ship.Response); !ok || resp != c.Response() {
		c.SetResponse(w)
	}
	return ship.Handler(h)(c)
}

// NewMiddleware returns a common middleware with the handler.
//
// Notice: the wrapped http.Handler has implemented the interface
//
//   type interface {
//       HandleHTTP(http.ResponseWriter, *http.Request) error
//   }
//
// So it can be used to wrap the error returned by other middleware handlers.
func NewMiddleware(handle func(http.Handler, http.ResponseWriter, *http.Request) error) Middleware {
	return func(next ship.Handler) ship.Handler {
		return func(c *ship.Context) error {
			req := c.Request()
			if ctx := req.Context(); getContext(ctx) == nil {
				req = req.WithContext(setContext(ctx, c))
				c.SetRequest(req)
			}
			return handle(mwHandler(next), c.Response(), req)
		}
	}
}

// func

/// ----------------------------------------------------------------------- ///

// Recover is a ship middleware to recover the panic if exists.
func Recover(next Handler) Handler {
	return func(ctx *ship.Context) (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = goapp.NewPanicError(e, 0)
			}
		}()

		return next(ctx)
	}
}

type reqBodyWrapper struct {
	io.Closer
	*bytes.Buffer
}

// Logger returns a logger middleware to log the request.
func Logger(logReqBody bool) Middleware {
	return func(next ship.Handler) ship.Handler {
		return func(c *ship.Context) (err error) {
			req := c.Request()
			var reqbody string
			if logReqBody {
				buf := c.AcquireBuffer()
				defer c.ReleaseBuffer(buf)

				_, err = io.CopyBuffer(buf, req.Body, make([]byte, 1024))
				if err != nil {
					return err
				}

				reqbody = buf.String()
				req.Body = reqBodyWrapper{Closer: req.Body, Buffer: buf}
			}

			start := time.Now()
			err = next(c)
			cost := time.Since(start)

			code := c.StatusCode()
			if err != nil && !c.IsResponded() {
				if hse, ok := err.(ship.HTTPServerError); ok {
					code = hse.Code
				} else {
					code = http.StatusInternalServerError
				}
			}

			var logger *log.Emitter
			if code < 400 {
				logger = log.Info()
			} else if code < 500 {
				logger = log.Warn()
			} else {
				logger = log.Error()
			}

			logger.Kv("addr", req.RemoteAddr).
				Kv("method", req.Method).
				Kv("uri", req.RequestURI).
				Kv("code", code).
				Kv("start", start.Unix()).
				Kv("cost", cost)

			if logReqBody {
				logger.Kv("reqbody", reqbody)
			}
			if err != nil {
				logger.Kv("err", err)
			}
			logger.Printf("log request")

			return
		}
	}
}
