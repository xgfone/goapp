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
	"io"
	"net/http"
	"time"

	"github.com/xgfone/go-log"
	"github.com/xgfone/goapp"
	"github.com/xgfone/ship/v5"
)

// Middleware is the type alias of ship.Middleware.
type Middleware = ship.Middleware

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
