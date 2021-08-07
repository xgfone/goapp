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

package opentracing_test

import (
	"context"
	"io/ioutil"
	"net/http"

	// tr "github.com/opentracing/opentracing-go"
	// "github.com/uber/jaeger-client-go/config"
	"github.com/xgfone/goapp/opentracing"
	"github.com/xgfone/ship/v5"
)

func initOpenTracing() {
	// cfg, err := config.FromEnv()
	// if err != nil {
	//     panic(err)
	// }
	// cfg.ServiceName = "app1"
	// tracer, _, err := cfg.NewTracer()
	// if err != nil {
	//     panic(err)
	// }
	// tr.SetGlobalTracer(tracer)
}

func init() {
	// Initialize the Jaeger Tracer implementation.
	initOpenTracing()

	// HTTPRoundTripper will extract the parent span from the context
	// of the sent http.Request, then create a new span for the current request.
	http.DefaultTransport = opentracing.NewHTTPRoundTripper(http.DefaultTransport, nil)
}

func ExampleOpenTracing() { // Main function
	app := ship.Default()

	// OpenTracing middleware extracts the span context from the http request
	// header, creates a new span as the server parent span from span context
	// and put it into the request context.
	app.Use(opentracing.OpenTracing(nil))

	app.Route("/app1").GET(func(c *ship.Context) (err error) {
		// ctx contains the parent span, which is extracted by OpenTracing middleware
		// from the HTTP request header.
		ctx := c.Request().Context()

		data1, err := request(ctx, "http://127.0.0.1:8002/app2")
		if err != nil {
			return
		}

		data2, err := request(ctx, "http://127.0.0.1:8003/app3")
		if err != nil {
			return
		}

		return c.Text(200, "app1:%s,%s", data1, data2)
	})

	ship.StartServer(":8001", app)
}

func request(ctx context.Context, url string) (data string, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}

	// Pass the parent span to the new http.Request.
	req = req.WithContext(ctx)

	// http.DefaultTransport will extract the parent span from ctx,
	// then create a child span for the current request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	return string(body), err
}
