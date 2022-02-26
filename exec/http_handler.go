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
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/xgfone/go-exec"
	"github.com/xgfone/ship/v5"
	"github.com/xgfone/ship/v5/middleware"
)

var bufpool = sync.Pool{New: func() interface{} {
	return bytes.NewBuffer(make([]byte, 0, 512))
}}

func getBuffer() *bytes.Buffer    { return bufpool.Get().(*bytes.Buffer) }
func putBuffer(buf *bytes.Buffer) { buf.Reset(); bufpool.Put(buf) }

func init() {
	if os.PathSeparator == '/' {
		DefaultShellConfig.Dir = "/tmp"
	}
}

// DefaultShellConfig is the default ShellConfig.
var DefaultShellConfig = ShellConfig{Shell: "bash", Timeout: time.Minute}

// ShellResultHandler is used to handle the result of the shell command or script.
type ShellResultHandler func(w http.ResponseWriter, stdout, stderr string, err error)

// ShellConfig is used to configure the shell execution.
type ShellConfig struct {
	Timeout time.Duration // The timeout to execute the shell command.
	Shell   string        // The shell name or path, which is "bash" by default.
	Dir     string        // The directory to save and run the shell script.
}

type shellRequest struct {
	Cmd     string `json:"cmd,omitempty"`
	Script  string `json:"script,omitempty"`
	Shell   string `json:"shell,omitempty"`
	Timeout string `json:"timeout,omitempty"`
}

type shellResult struct {
	Stdout string `json:"stdout,omitempty"`
	Stderr string `json:"stderr,omitempty"`
	Error  string `json:"error,omitempty"`
}

// ExecuteShellHandler returns a http handler to execute a SHELL command or script.
//
// The request body is the command to be executed as JSON like this:
//
//    {
//        "cmd":     "BASE64_COMMAND",             // Optional
//        "script":  "BASE64_SCRIPT_FILE_CONTENT", // Optional
//        "shell":   "SHELL_COMMAND",              // Optional
//        "timeout": "10s"                         // Optional
//    }
//
// handle is used to handle the result of the command or script. If nil,
// it will use the default that returns a JSON as the response body like this,
//
//   {
//       "stdout": "BASE64_STD_OUTPUT",
//       "stderr": "BASE64_STD_ERR_OUTPUT",
//       "error": "BASE64 failure reason. If successfully, it is empty."
//   }
//
// Notice:
//   1. The executed command or script must be encoded by base64.
//   2. If shell is given, it will override the Shell in ShellConfig.
//   3. If timeout is given, it will override the Timeout in ShellConfig.
//
// The returned http handler is very dangerous, and should not be called
// by the non-trusted callers.
func ExecuteShellHandler(handler ShellResultHandler, config ...ShellConfig) http.Handler {
	var conf ShellConfig
	if len(config) > 0 {
		conf = config[0]
	}

	if conf.Shell == "" {
		if conf.Shell = exec.DefaultShell; conf.Shell == "" {
			conf.Shell = "bash"
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := getBuffer()
		defer putBuffer(buf)

		_, err := io.CopyBuffer(buf, r.Body, make([]byte, 1024))
		if err != nil {
			w.WriteHeader(400)
			io.WriteString(w, err.Error())
			return
		}

		var cmd shellRequest
		if err := json.NewDecoder(buf).Decode(&cmd); err != nil {
			w.WriteHeader(400)
			io.WriteString(w, err.Error())
			return
		}

		timeout := conf.Timeout
		if cmd.Timeout != "" {
			t, err := time.ParseDuration(cmd.Timeout)
			if err != nil {
				w.WriteHeader(400)
				io.WriteString(w, err.Error())
				return
			}
			timeout = t
		}

		var cancel func()
		c := context.Background()
		if timeout > 0 {
			c, cancel = context.WithTimeout(c, timeout)
			defer cancel()
		}

		shell := cmd.Shell
		if shell == "" {
			shell = conf.Shell
		}

		var stdout, stderr string
		if cmd.Cmd != "" {
			stdout, stderr, err = executeShellCommand(c, shell, cmd.Cmd)
		} else if cmd.Script != "" {
			stdout, stderr, err = executeShellScript(c, shell, conf.Dir, cmd.Script)
		}

		if handler != nil {
			handler(w, stdout, stderr, err)
		} else {
			defaultHandler(w, buf, stdout, stderr, err)
		}
	})
}

func defaultHandler(w http.ResponseWriter, buf *bytes.Buffer, stdout, stderr string, err error) {
	var result shellResult
	if len(stdout) > 0 {
		result.Stdout = base64.StdEncoding.EncodeToString([]byte(stdout))
	}
	if len(stderr) > 0 {
		result.Stderr = base64.StdEncoding.EncodeToString([]byte(stderr))
	}

	switch e := err.(type) {
	case nil:
	case exec.Result:
		result.Error = base64.StdEncoding.EncodeToString([]byte(e.Err.Error()))
	default:
		result.Error = base64.StdEncoding.EncodeToString([]byte(e.Error()))
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	buf.Reset()
	if err := json.NewEncoder(buf).Encode(result); err != nil {
		w.WriteHeader(500)
		io.WriteString(w, err.Error())
	} else {
		w.WriteHeader(200)
		w.Write(buf.Bytes())
	}
}

func executeShellCommand(c context.Context, shell, cmd string) (string, string, error) {
	bs, err := base64.StdEncoding.DecodeString(cmd)
	if err != nil {
		err = ship.ErrBadRequest.Newf("failed to decode base64 '%s': %v", cmd, err)
		return "", "", err
	}

	stdout, stderr, err := exec.Run(c, shell, "-c", string(bs))
	if err != nil {
		err = ship.ErrInternalServerError.New(err)
	}

	return stdout, stderr, err
}

var generateTmpFilename = middleware.GenerateToken(16)

func executeShellScript(c context.Context, shell, dir, script string) (string, string, error) {
	scriptContent, err := base64.StdEncoding.DecodeString(script)
	if err != nil {
		err = ship.ErrBadRequest.Newf("failed to decode base64 '%s': %v", script, err)
		return "", "", err
	}

	filename, err := exec.GetScriptFile(dir, string(scriptContent))
	if err != nil {
		return "", "", ship.ErrInternalServerError.New(err)
	}
	defer os.Remove(filename)

	stdout, stderr, err := exec.Run(c, shell, filename)
	if err != nil {
		err = ship.ErrInternalServerError.New(err)
	}
	return stdout, stderr, err
}
