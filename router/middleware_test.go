// Copyright 2021 xgfone
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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xgfone/goapp"
	"github.com/xgfone/ship/v5"
)

func TestRecover(t *testing.T) {
	err := Recover(ship.NothingHandler())(nil)
	if err != nil {
		t.Error(err)
	}

	panicStack := "middleware_test.go:33"
	err = Recover(func(ctx *ship.Context) (err error) { panic("testpanic") })(nil)
	switch e := err.(type) {
	case nil:
		t.Error(err)
	case goapp.PanicError:
		if e.Panic.(string) != "testpanic" {
			t.Error(e.Panic)
		} else if _len := len(e.Stack); _len == 0 {
			t.Error(e.Stack)
		} else if stack := e.Stack[_len-1].String(); stack != panicStack {
			t.Errorf("expect '%s', got '%s'", panicStack, stack)
		}
	default:
		t.Error(err)
	}
}

func TestNewMiddleware(t *testing.T) {
	r := ship.New()
	r.Use(NewMiddleware(func(h http.Handler, w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("test", "abc")
		h.ServeHTTP(w, r)
		return nil
	}))
	r.Route("/").GET(func(c *ship.Context) error { return c.NoContent(201) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Errorf("expect the status code %d, but got %d", 201, rec.Code)
	}
	if test := rec.Header().Get("test"); test != "abc" {
		t.Errorf("expect 'test' header '%s', but got '%s'", "abc", test)
	}
}

func BenchmarkNewMiddleware(b *testing.B) {
	r := ship.New()
	r.Use(NewMiddleware(func(h http.Handler, w http.ResponseWriter, r *http.Request) error {
		h.ServeHTTP(w, r)
		return nil
	}))
	r.Route("/").GET(func(c *ship.Context) error { return nil })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.ServeHTTP(rec, req)
		}
	})
}
