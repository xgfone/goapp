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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/xgfone/go-exec"
	"github.com/xgfone/ship/v4"
	"github.com/xgfone/ship/v4/middleware"
)

// Handler is the type alias of ship.Handler.
type Handler = ship.Handler

// DisableBuiltinPrometheusCollector removes the collectors that the default
// prometheus register registers
func DisableBuiltinPrometheusCollector() {
	prometheus.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.Unregister(prometheus.NewGoCollector())
}

// PrometheusHandler returns a prometheus handler.
//
// if missing gatherer, it is prometheus.DefaultGatherer.
func PrometheusHandler(gatherer ...prometheus.Gatherer) Handler {
	gather := prometheus.DefaultGatherer
	if len(gatherer) > 0 && gatherer[0] != nil {
		gather = gatherer[0]
	}

	return func(ctx *ship.Context) error {
		mfs, err := gather.Gather()
		if err != nil {
			return err
		}

		ct := expfmt.Negotiate(ctx.Request().Header)
		ctx.SetContentType(string(ct))
		enc := expfmt.NewEncoder(ctx, ct)

		for _, mf := range mfs {
			if err = enc.Encode(mf); err != nil {
				ctx.Logger().Errorf("failed to encode prometheus metric: %s", err)
			}
		}

		return nil
	}
}

// DefaultShellConfig is the default ShellConfig.
var DefaultShellConfig = ShellConfig{Shell: "bash", Timeout: time.Minute}

// ShellConfig is used to configure the shell execution.
type ShellConfig struct {
	Dir     string        // The directory to save and run the shell script.
	Shell   string        // The shell name or path, which is "sh" by default.
	Timeout time.Duration // The timeout to execute the shell command.
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

// ExecuteShell returns a handler to execute a SHELL command or script.
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
//       "error": "failure reason. If successfully, it is empty."
//   }
//
// Notice:
//   1. The executed command or script must be encoded by base64.
//   2. If shell is given, it will override the Shell in ShellConfig.
//   3. If timeout is given, it will override the Timeout in ShellConfig.
//
// The returned handler is very dangerous, and should not be called
// by the non-trusted callers.
func ExecuteShell(handle func(ctx *ship.Context, stdout, stderr []byte, err error) error,
	config ...ShellConfig) Handler {
	var conf ShellConfig
	if len(config) > 0 {
		conf = config[0]
	}

	if conf.Shell == "" {
		conf.Shell = exec.DefaultShell
	}

	if handle == nil {
		handle = func(c *ship.Context, stdout, stderr []byte, err error) error {
			var result shellResult
			if len(stdout) > 0 {
				result.Stdout = base64.StdEncoding.EncodeToString(stdout)
			}
			if len(stderr) > 0 {
				result.Stderr = base64.StdEncoding.EncodeToString(stderr)
			}
			if err != nil {
				he := err.(ship.HTTPServerError)
				if ce, ok := he.Err.(exec.CmdError); ok {
					result.Error = ce.Err.Error()
				} else {
					result.Error = he.Err.Error()
				}
			}

			return c.JSON(200, result)
		}
	}

	return func(ctx *ship.Context) error {
		var cmd shellRequest
		buf, err := ctx.GetBodyReader()
		if err != nil {
			return ship.ErrBadRequest.New(err)
		}
		defer ctx.ReleaseBuffer(buf)

		if err := json.NewDecoder(buf).Decode(&cmd); err != nil {
			return ship.ErrBadRequest.New(err)
		}

		timeout := conf.Timeout
		if cmd.Timeout != "" {
			t, err := time.ParseDuration(cmd.Timeout)
			if err != nil {
				return ship.ErrBadRequest.New(err)
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

		return handle(ctx, []byte(stdout), []byte(stderr), err)
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
		return "", "", ship.ErrInternalServerError.New(err)
	}

	return stdout, stderr, nil
}

var generateTmpFilename = middleware.GenerateToken(16)

func executeShellScript(c context.Context, shell, dir, script string) (string, string, error) {
	bs, err := base64.StdEncoding.DecodeString(script)
	if err != nil {
		err = ship.ErrBadRequest.Newf("failed to decode base64 '%s': %v", script, err)
		return "", "", err
	}

	filename := fmt.Sprintf("__run_shell_script_%s.sh", generateTmpFilename())
	if dir != "" {
		filename = filepath.Join(dir, filename)
	}

	if err = ioutil.WriteFile(filename, bs, 0700); err != nil {
		return "", "", ship.ErrInternalServerError.New(err)
	}
	defer os.Remove(filename)

	stdout, stderr, err := exec.Run(c, shell, filename)
	if err != nil {
		return "", "", ship.ErrInternalServerError.New(err)
	}
	return stdout, stderr, nil
}
