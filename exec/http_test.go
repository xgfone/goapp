// Copyright 2022 xgfone
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

package exec

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExecuteShell(t *testing.T) {
	handler := ExecuteShellHandler(func(w http.ResponseWriter, stdout, stderr string, err error) {
		w.WriteHeader(200)
		io.WriteString(w, stdout+"|"+stderr)
	})

	router := http.NewServeMux()
	router.Handle("/shell", handler)

	// For Command
	cmd := base64.StdEncoding.EncodeToString([]byte("echo -n abc123"))
	buf := bytes.NewBufferString(fmt.Sprintf(`{"cmd": "%s", "shell": "bash"}`, cmd))
	req, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1/shell", buf)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("expected StatusCode '200', but got '%d'", rec.Code)
	} else if body := rec.Body.String(); body != "abc123|" {
		t.Errorf("expected Body 'abc123|', but got '%s'", body)
	}

	// For Script
	cmd = base64.StdEncoding.EncodeToString([]byte("echo -n abc123"))
	buf = bytes.NewBufferString(fmt.Sprintf(`{"script": "%s", "shell": "bash", "timeout": "10s"}`, cmd))
	req, _ = http.NewRequest(http.MethodPost, "http://127.0.0.1/shell", buf)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Errorf("expected StatusCode '200', but got '%d'", rec.Code)
	} else if body := rec.Body.String(); body != "abc123|" {
		t.Errorf("expected Body 'abc123|', but got '%s'", body)
	}
}
