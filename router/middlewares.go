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
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

			return
		}
	}
}

// HistogramBuckets is used to replace the default Histogram buckets.
var HistogramBuckets = []float64{.005, .01, .025, .05, .075, .1, .25, .5, .75, 1, 1.5, 2}

// Prometheus returns a middleware to handle the prometheus metrics.
//
// The first argument is the namespace, and the second is the subsystem.
// Both of them are optional.
func Prometheus(namespaceAndSubsystem ...string) Middleware {
	var namespace, subsystem string
	switch len(namespaceAndSubsystem) {
	case 0:
	case 1:
		namespace = namespaceAndSubsystem[0]
	default:
		namespace = namespaceAndSubsystem[0]
		subsystem = namespaceAndSubsystem[1]
	}

	requestNumber := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,

			Name: "http_request_number",
			Help: "The number of the current http request",
		},
		[]string{"method", "path"})

	requestTotal := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,

			Name: "http_request_total",
			Help: "The total number of the http request",
		},
		[]string{"method", "path", "code"})

	requestDurations := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,

			Name: "http_request_duration_seconds",
			Help: "The duration to handle the request",

			Buckets: HistogramBuckets,
		},
		[]string{"method", "path", "code"})

	return func(next Handler) Handler {
		return func(ctx *ship.Context) (err error) {
			var start time.Time

			code := 200
			path := ctx.Request().URL.Path
			method := ctx.Method()

			requestNumberGuage := requestNumber.With(prometheus.Labels{"method": method, "path": path})
			requestNumberGuage.Inc()

			defer func() {
				if e := recover(); e != nil {
					err = goapp.NewPanicError(e, 0)
					code = 500
				} else {
					switch e := err.(type) {
					case nil:
						code = ctx.StatusCode()
					case ship.HTTPServerError:
						code = e.Code
					default:
						code = 500
					}
				}

				labels := prometheus.Labels{
					"method": method,
					"path":   path,
					"code":   fmt.Sprintf("%d", code),
				}

				requestTotal.With(labels).Inc()
				requestDurations.With(labels).Observe(time.Since(start).Seconds())
				requestNumberGuage.Dec()
			}()

			start = time.Now()
			err = next(ctx)
			return
		}
	}
}
