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
	"context"
	"encoding/base64"
	"errors"
	"net/http"

	httpclient "github.com/xgfone/go-http-client"
	"github.com/xgfone/ship/v5"
)

// ExecShellByHTTP executes the shell command or script by HTTP.
//
// Notice: it uses the default interface implementation of ExecuteShell.
func ExecShellByHTTP(url, cmd, script string) (stdout, stderr string, err error) {
	var req shellRequest
	var resp shellResult

	if cmd != "" {
		req.Cmd = base64.StdEncoding.EncodeToString([]byte(cmd))
	}
	if script != "" {
		req.Script = base64.StdEncoding.EncodeToString([]byte(script))
	}

	err = httpclient.Post(url).
		SetContentType("application/json; charset=UTF-8").
		SetAccepts("application/json").
		SetBody(req).
		Do(context.Background(), &resp).
		Close().
		Unwrap()
	if err != nil {
		return
	} else if resp.Error != "" {
		err = ship.NewHTTPClientError(http.MethodPost, url, 200, errors.New(resp.Error))
		return
	}

	var sout, serr []byte
	if resp.Stdout != "" {
		if sout, err = base64.StdEncoding.DecodeString(resp.Stdout); err != nil {
			err = ship.NewHTTPClientError(http.MethodPost, url, 200, err)
			return
		}
	}
	if resp.Stderr != "" {
		if serr, err = base64.StdEncoding.DecodeString(resp.Stderr); err != nil {
			err = ship.NewHTTPClientError(http.MethodPost, url, 200, err)
			return
		}
	}

	return string(sout), string(serr), nil
}
