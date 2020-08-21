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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/xgfone/go-tools/v7/execution"
	"github.com/xgfone/ship/v3"
	"github.com/xgfone/ship/v3/middleware"
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
	Shell   string        // The shell name or path, which is "sh" by default.
	Timeout time.Duration // The timeout to execute the shell command.
}

type shellResult struct {
	Stdout string `json:"stdout,omitempty"`
	Stderr string `json:"stderr,omitempty"`
	Error  error  `json:"error,omitempty"`
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
		conf.Shell = "sh"
	}

	if handle == nil {
		handle = func(c *ship.Context, stdout []byte, stderr []byte, err error) error {
			return c.JSON(200, shellResult{
				Stdout: base64.StdEncoding.EncodeToString(stdout),
				Stderr: base64.StdEncoding.EncodeToString(stderr),
				Error:  err,
			})
		}
	}

	return func(ctx *ship.Context) error {
		type Cmd struct {
			Cmd     string `json:"cmd"`
			Script  string `json:"script"`
			Shell   string `json:"shell"`
			Timeout string `json:"timeout"`
		}

		var cmd Cmd
		buf, err := ctx.GetBodyReader()
		if err != nil {
			return ship.ErrBadRequest.NewError(err)
		}
		defer ctx.ReleaseBuffer(buf)

		if err := json.NewDecoder(buf).Decode(&cmd); err != nil {
			return ship.ErrBadRequest.NewError(err)
		}

		timeout := conf.Timeout
		if cmd.Timeout != "" {
			t, err := time.ParseDuration(cmd.Timeout)
			if err != nil {
				return ship.ErrBadRequest.NewError(err)
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

		var stdout, stderr []byte
		if cmd.Cmd != "" {
			stdout, stderr, err = executeShellCommand(c, shell, cmd.Cmd)
		} else if cmd.Script != "" {
			stdout, stderr, err = executeShellScript(c, shell, cmd.Script)
		}

		return handle(ctx, stdout, stderr, err)
	}
}

func executeShellCommand(c context.Context, shell, cmd string) ([]byte, []byte, error) {
	bs, err := base64.StdEncoding.DecodeString(cmd)
	if err != nil {
		err = ship.ErrBadRequest.NewMsg("failed to decode base64 '%s': %v", cmd, err)
		return nil, nil, err
	}

	stdout, stderr, err := execution.Run(c, shell, "-c", string(bs))
	if err != nil {
		return nil, nil, ship.ErrInternalServerError.NewError(err)
	}

	return stdout, stderr, nil
}

var generateTmpFilename = middleware.GenerateToken(16)

func executeShellScript(c context.Context, shell, script string) ([]byte, []byte, error) {
	bs, err := base64.StdEncoding.DecodeString(script)
	if err != nil {
		err = ship.ErrBadRequest.NewMsg("failed to decode base64 '%s': %v", script, err)
		return nil, nil, err
	}

	filename := fmt.Sprintf("__run_shell_script_%s.sh", generateTmpFilename())
	if err = ioutil.WriteFile(filename, bs, 0700); err != nil {
		return nil, nil, ship.ErrInternalServerError.NewError(err)
	}
	defer os.Remove(filename)

	stdout, stderr, err := execution.Run(c, shell, filename)
	if err != nil {
		return nil, nil, ship.ErrInternalServerError.NewError(err)
	}
	return stdout, stderr, nil
}