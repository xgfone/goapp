// Copyright 2022~2023 xgfone
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

// Package middleware registers some log config options to configure the log
// middleware, which just needs to be imported to do some initializations.
//
// Example:
//
//	import _ "github.com/xgfone/goapp/log/middleware"
package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"unsafe"

	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/go-apiserver/http/header"
	"github.com/xgfone/go-apiserver/http/middlewares"
	"github.com/xgfone/go-apiserver/http/reqresp"
	"github.com/xgfone/go-apiserver/log"
	"github.com/xgfone/go-apiserver/middleware/logger"
	"github.com/xgfone/go-generics/slices"
	"github.com/xgfone/go-rawjson"
)

var (
	group          = gconf.Group("log")
	logQuery       = group.NewBool("query", false, "If true, log the request query.")
	logReqBody     = group.NewBool("reqbody", false, "If true, log the request body.")
	logRespBody    = group.NewBool("respbody", false, "If true, log the response body.")
	logReqHeaders  = group.NewBool("reqheaders", false, "If true, log the request headers.")
	logRespHeaders = group.NewBool("respheaders", false, "If true, log the response headers.")

	logBodyMaxLen = group.NewInt("bodymaxlen", 1024,
		"The maximum length of the request or response body to log.")
	logBodyTypes = group.NewStringSlice("bodytypes", []string{
		header.MIMEApplicationJSON, header.MIMEApplicationForm,
	}, "The content types of the request or response body to log.")
)

var bufpool = sync.Pool{New: func() interface{} { return bytes.NewBuffer(make([]byte, 0, 512)) }}

func getbuffer() *bytes.Buffer  { return bufpool.Get().(*bytes.Buffer) }
func putbuffer(b *bytes.Buffer) { b.Reset(); bufpool.Put(b) }

type bufferCloser struct {
	*bytes.Buffer
	io.Closer
}

func init() {
	logger.Start = start
	logger.Enabled = enabled
	middlewares.WrapLoggerResponse = wrapResponse
}

func enabled(ctx context.Context, req interface{}) bool {
	r, ok := req.(*http.Request)
	return !ok || r.URL.Path != "/"
}

func start(ctx context.Context, req interface{}) logger.Collector {
	r, ok := req.(*http.Request)
	if !ok {
		return nil
	}

	logquery := logQuery.Get()
	logreqbody := logReqBody.Get()
	logreqheader := logReqHeaders.Get()
	logresheader := logRespHeaders.Get()

	var reqct string
	var reqbody []byte
	var reqbodybuf *bytes.Buffer
	if logreqbody {
		reqct = header.ContentType(r.Header)
		if logreqbody = slices.Contains(logBodyTypes.Get(), reqct); logreqbody {
			reqbodybuf = getbuffer()
			_, err := io.CopyBuffer(reqbodybuf, r.Body, make([]byte, 512))
			if err != nil {
				log.Error("fail to read the request body", "raddr", r.RemoteAddr,
					"method", r.Method, "path", r.RequestURI, "err", err)
			}

			reqbody = reqbodybuf.Bytes()
			r.Body = bufferCloser{Buffer: reqbodybuf, Closer: r.Body}
		}
	}

	if !logreqbody && !logquery && !logreqheader && !logresheader {
		return nil
	}

	return func(kvs []interface{}) (newkvs []interface{}, clean func()) {
		if logquery {
			kvs = append(kvs, "query", r.URL.RawQuery)
		}

		if logreqbody {
			clean = func() { putbuffer(reqbodybuf) }
			kvs = append(kvs, "reqbodylen", len(reqbody))
			if maxlen := logBodyMaxLen.Get(); maxlen > 0 && len(reqbody) <= maxlen {
				if strings.HasSuffix(reqct, "json") {
					kvs = append(kvs, "reqbody", rawjson.Bytes(reqbody))
				} else {
					reqbodystr := unsafe.String(unsafe.SliceData(reqbody), len(reqbody))
					kvs = append(kvs, "reqbody", reqbodystr)
				}
			}
		}

		if logreqheader {
			kvs = append(kvs, "reqheaders", r.Header)
		}

		if logresheader {
			if c := reqresp.GetContextFromCtx(ctx); c != nil {
				kvs = append(kvs, "respheaders", c.ResponseWriter.Header())
			}
		}

		newkvs = kvs
		return
	}
}

func wrapResponse(w http.ResponseWriter, r *http.Request) (new http.ResponseWriter, kvs []interface{}, clean func()) {
	if !logRespBody.Get() {
		return w, nil, nil
	}

	respbuf := getbuffer()
	new = reqresp.NewResponseWriter(w, reqresp.DisableReaderFrom(),
		reqresp.WriteWithResponse(func(w http.ResponseWriter, p []byte) (n int, err error) {
			if n, err = w.Write(p); n > 0 {
				respbuf.Write(p[:n])
			}
			return
		}))

	kvs = []interface{}{
		"respbodylen", respbodylen{b: respbuf},
		"respbody", respbodycnt{w: w, b: respbuf},
	}
	clean = func() { putbuffer(respbuf) }
	return
}

type respbodylen struct{ b *bytes.Buffer }

func (v respbodylen) LogValue() log.Value { return log.AnyValue(v.b.Len()) }

type respbodycnt struct {
	w http.ResponseWriter
	b *bytes.Buffer
}

func (v respbodycnt) LogValue() log.Value {
	if maxlen := logBodyMaxLen.Get(); maxlen > 0 && v.b.Len() > maxlen {
		return log.AnyValue(nil)
	}

	ct := header.ContentType(v.w.Header())
	if !slices.Contains(logBodyTypes.Get(), ct) {
		return log.AnyValue(nil)
	}

	respbody := v.b.Bytes()
	if strings.HasSuffix(ct, "json") {
		return log.AnyValue(rawjson.Bytes(respbody))
	}

	respbodystr := unsafe.String(unsafe.SliceData(respbody), len(respbody))
	return log.AnyValue(respbodystr)
}
