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
	"io"
	"os"
	"time"

	"github.com/xgfone/go-exec"
	"github.com/xgfone/ship/v5"
	"github.com/xgfone/ship/v5/middleware"
)

// Handler is the type alias of ship.Handler.
type Handler = ship.Handler

// DefaultShellConfig is the default ShellConfig.
var DefaultShellConfig = ShellConfig{Shell: "bash", Timeout: time.Minute}

func init() {
	if os.PathSeparator == '/' {
		DefaultShellConfig.Dir = "/tmp"
	}
}

// ShellConfig is used to configure the shell execution.
type ShellConfig struct {
	Dir     string        // The directory to save and run the shell script.
	Shell   string        // The shell name or path, which is "bash" by default.
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
//       "error": "BASE64 failure reason. If successfully, it is empty."
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
		if conf.Shell = exec.DefaultShell; conf.Shell == "" {
			conf.Shell = "bash"
		}
	}

	if handle == nil {
		handle = func(c *ship.Context, stdout, stderr string, err error) error {
			var result shellResult
			if len(stdout) > 0 {
				result.Stdout = base64.StdEncoding.EncodeToString([]byte(stdout))
			}
			if len(stderr) > 0 {
				result.Stderr = base64.StdEncoding.EncodeToString([]byte(stderr))
			}
			if err != nil {
				he := err.(ship.HTTPServerError)
				if ce, ok := he.Err.(exec.Result); ok {
					result.Error = base64.StdEncoding.EncodeToString([]byte(ce.Err.Error()))
				} else {
					result.Error = base64.StdEncoding.EncodeToString([]byte(he.Err.Error()))
				}
			}

			return c.JSON(200, result)
		}
	}

	return func(ctx *ship.Context) error {
		buf := ctx.AcquireBuffer()
		defer ctx.ReleaseBuffer(buf)
		_, err := io.CopyBuffer(buf, ctx.Body(), make([]byte, 1024))
		if err != nil {
			return ship.ErrBadRequest.New(err)
		}

		var cmd shellRequest
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

		return handle(ctx, stdout, stderr, err)
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

	filename, err := exec.GetScriptFile(dir, scriptContent)
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
