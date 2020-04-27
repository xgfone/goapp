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

package goapp

import (
	"fmt"
	"time"

	"github.com/go-stack/stack"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/xgfone/ship/v2"
)

// Middleware is the type alias of ship.Middleware.
type Middleware = ship.Middleware

// Recover is a ship middleware to recover the panic if exists.
func Recover(next Handler) Handler {
	return func(ctx *ship.Context) (err error) {
		defer func() {
			switch e := recover().(type) {
			case nil:
			default:
				s := stack.Trace().TrimBelow(stack.Caller(1)).TrimRuntime()
				if len(s) == 0 {
					err = fmt.Errorf("panic: %v", e)
				} else {
					err = fmt.Errorf("panic: %v, stack=%v", e, s)
				}
			}
		}()

		return next(ctx)
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
					err = fmt.Errorf("%v", e)
					code = 500
				} else {
					code = ctx.StatusCode()
				}

				labels := prometheus.Labels{
					"method": method,
					"path":   path,
					"code":   fmt.Sprintf("%d", code),
				}

				requestTotal.With(labels).Inc()
				requestDurations.With(labels).Observe(time.Now().Sub(start).Seconds())
				requestNumberGuage.Dec()
			}()

			start = time.Now()
			err = next(ctx)
			return
		}
	}
}
