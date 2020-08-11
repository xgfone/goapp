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
	"net/http"
	"net/url"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/xgfone/ship/v3"
)

// OpenTracingOption is used to configure the OpenTracingServer.
type OpenTracingOption struct {
	Tracer        opentracing.Tracer // Default: opentracing.GlobalTracer()
	ComponentName string             // Default: use ComponentNameFunc(req)

	// ComponentNameFunc is used to get the component name if ComponentName
	// is empty.
	//
	// Default: "net/http"
	ComponentNameFunc func(*http.Request) string

	// URLTagFunc is used to get the value of the tag "http.url".
	//
	// Default: url.String()
	URLTagFunc func(*url.URL) string

	// SpanFilter is used to filter the span if returning true.
	//
	// Default: return false
	SpanFilter func(*http.Request) bool

	// OperationNameFunc is used to the operation name.
	//
	// Default: fmt.Sprintf("HTTP %s %s", r.Method, r.URL.Path)
	OperationNameFunc func(*http.Request) string

	// SpanObserver is used to do extra things of the span for the request.
	//
	// For example,
	//    OpenTracingOption {
	//        SpanObserver: func(*http.Request, opentracing.Span) {
	//            ext.PeerHostname.Set(span, req.Host)
	//        },
	//    }
	//
	// Default: Do nothing.
	SpanObserver func(*http.Request, opentracing.Span)
}

// Init initializes the OpenTracingOption.
func (o *OpenTracingOption) Init() {
	if o.ComponentNameFunc == nil {
		o.ComponentNameFunc = func(*http.Request) string { return "net/http" }
	}
	if o.URLTagFunc == nil {
		o.URLTagFunc = func(u *url.URL) string { return u.String() }
	}
	if o.SpanFilter == nil {
		o.SpanFilter = func(r *http.Request) bool { return false }
	}
	if o.SpanObserver == nil {
		o.SpanObserver = func(*http.Request, opentracing.Span) {}
	}
	if o.OperationNameFunc == nil {
		o.OperationNameFunc = func(r *http.Request) string {
			return fmt.Sprintf("HTTP %s %s", r.Method, r.URL.Path)
		}
	}
}

// GetComponentName returns ComponentName if it is not empty.
// Or ComponentNameFunc(req) instead.
func (o *OpenTracingOption) GetComponentName(req *http.Request) string {
	if o.ComponentName == "" {
		return o.ComponentNameFunc(req)
	}
	return o.ComponentName
}

// GetTracer returns the OpenTracing tracker.
func (o *OpenTracingOption) GetTracer() opentracing.Tracer {
	if o.Tracer == nil {
		return opentracing.GlobalTracer()
	}
	return o.Tracer
}

// OpenTracingRoundTripper is a RoundTripper to support OpenTracing,
// which extracts the parent span from the context of the sent http.Request,
// then creates a new span by the context of the parent span for http.Request.
type OpenTracingRoundTripper struct {
	http.RoundTripper
	OpenTracingOption
}

// NewOpenTracingRoundTripper returns a new OpenTracingRoundTripper.
func NewOpenTracingRoundTripper(rt http.RoundTripper, opt *OpenTracingOption) *OpenTracingRoundTripper {
	var o OpenTracingOption
	if opt != nil {
		o = *opt
	}
	o.Init()
	return &OpenTracingRoundTripper{RoundTripper: rt, OpenTracingOption: o}
}

// WrappedRoundTripper returns the wrapped http.RoundTripper.
func (rt *OpenTracingRoundTripper) WrappedRoundTripper() http.RoundTripper {
	return rt.RoundTripper
}

func (rt *OpenTracingRoundTripper) roundTrip(req *http.Request) (*http.Response, error) {
	if rt.RoundTripper == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return rt.RoundTripper.RoundTrip(req)
}

// RoundTrip implements the interface http.RounderTripper.
func (rt *OpenTracingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.SpanFilter(req) {
		return rt.roundTrip(req)
	}

	ctx := req.Context()
	opts := []opentracing.StartSpanOption{ext.SpanKindRPCClient}
	if pspan := opentracing.SpanFromContext(ctx); pspan != nil {
		opts = []opentracing.StartSpanOption{opentracing.ChildOf(pspan.Context())}
	}

	tracer := rt.GetTracer()
	sp := tracer.StartSpan(rt.OperationNameFunc(req), opts...)
	ext.HTTPUrl.Set(sp, rt.URLTagFunc(req.URL))
	ext.Component.Set(sp, rt.GetComponentName(req))
	ext.HTTPMethod.Set(sp, req.Method)
	rt.SpanObserver(req, sp)
	defer sp.Finish()

	carrier := opentracing.HTTPHeadersCarrier(req.Header)
	tracer.Inject(sp.Context(), opentracing.HTTPHeaders, carrier)

	return rt.roundTrip(req.WithContext(opentracing.ContextWithSpan(ctx, sp)))
}

// OpenTracing is a middleware to support OpenTracing, which extracts the span
// context from the http request header, creates a new span as the server span
// from the span context, and put it into the request context.
func OpenTracing(opt *OpenTracingOption) Middleware {
	var o OpenTracingOption
	if opt != nil {
		o = *opt
	}
	o.Init()

	const format = opentracing.HTTPHeaders
	return func(next ship.Handler) ship.Handler {
		return func(ctx *ship.Context) (err error) {
			req := ctx.Request()
			if o.SpanFilter(req) {
				return next(ctx)
			}

			tracer := o.GetTracer()
			sc, _ := tracer.Extract(format, opentracing.HTTPHeadersCarrier(req.Header))
			sp := tracer.StartSpan(o.OperationNameFunc(req), ext.RPCServerOption(sc))
			ext.HTTPMethod.Set(sp, req.Method)
			ext.Component.Set(sp, o.GetComponentName(req))
			ext.HTTPUrl.Set(sp, o.URLTagFunc(req.URL))
			o.SpanObserver(req, sp)

			req = req.WithContext(opentracing.ContextWithSpan(req.Context(), sp))
			ctx.SetRequest(req)

			defer func() {
				if e := recover(); e != nil {
					ext.Error.Set(sp, true)
					sp.Finish()
					panic(e)
				}

				statusCode := ctx.StatusCode()
				if !ctx.IsResponded() {
					switch e := err.(type) {
					case nil:
					case ship.HTTPError:
						statusCode = e.Code
					default:
						statusCode = 500
					}
				}

				ext.HTTPStatusCode.Set(sp, uint16(statusCode))
				if statusCode >= 500 {
					ext.Error.Set(sp, true)
				}
				sp.Finish()
			}()

			err = next(ctx)
			return err
		}
	}
}