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

package exec

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/xgfone/go-exec"
)

// SSHUser is the user to execute the shell command by SSH.
var SSHUser = "root"

// SSHOptions is the options of ssh/scp command.
var SSHOptions = "-o StrictHostKeyChecking=no"

// ExecuteCmdBySSH executes the shell command by SSH.
func ExecuteCmdBySSH(host, cmd string) (stdout, stderr string, err error) {
	cmd = fmt.Sprintf(`ssh %s %s@%s "%s"`, SSHOptions, SSHUser, host, cmd)
	_stdout, _stderr, err := exec.RunShellCmd(context.Background(), cmd)
	if err == nil {
		stdout, stderr = string(_stdout), string(_stderr)
	}
	return
}

// CopyFilesToRemoteBySSH copies the files from the local to the remote.
func CopyFilesToRemoteBySSH(remoteHost, remoteDirOrFile string, localFiles ...string) error {
	if len(localFiles) == 0 {
		return nil
	}

	files := strings.Join(localFiles, " ")
	cmd := fmt.Sprintf("scp %s %s %s@%s:%s", SSHOptions, files, SSHUser, remoteHost, remoteDirOrFile)
	_, _, err := exec.RunShellCmd(context.Background(), cmd)
	return err
}

// CopyFilesFromRemoteBySSH copies the files from the remote to the local.
func CopyFilesFromRemoteBySSH(remoteHost, localDirOrFile string, remoteFiles ...string) error {
	if len(remoteFiles) == 0 {
		return nil
	}

	for i, file := range remoteFiles {
		remoteFiles[i] = fmt.Sprintf("%s@%s:%s", SSHUser, remoteHost, file)
	}

	files := strings.Join(remoteFiles, " ")
	cmd := fmt.Sprintf("scp %s %s %s", SSHOptions, files, localDirOrFile)
	_, _, err := exec.RunShellCmd(context.Background(), cmd)
	return err
}

// ExecuteScriptBySSH executes the shell script by SSH.
func ExecuteScriptBySSH(host, script string) (stdout, stderr string, err error) {
	filename1 := getScriptFile(script)
	filename2 := filename1
	if exec.ShellScriptDir != "" {
		filename1 = filepath.Join(exec.ShellScriptDir, filename1)
		filename2 = filename1
	} else {
		filename2 = filepath.Join("~", filename1)
	}

	if err = ioutil.WriteFile(filename1, []byte(script), 0700); err != nil {
		return
	}
	defer os.Remove(filename1)

	if err = CopyFilesToRemoteBySSH(host, filename2, filename1); err != nil {
		return
	}
	defer ExecuteCmdBySSH(host, fmt.Sprintf("rm -f %s", filename2))

	shell := exec.DefaultCmd.Shell
	if shell == "" {
		if shell = exec.DefaultShell; shell == "" {
			shell = "sh"
		}
	}
	return ExecuteCmdBySSH(host, fmt.Sprintf("%s %s", shell, filename2))
}

func getScriptFile(script string) (filename string) {
	data := []byte(script)
	md5sum := md5.Sum(data)
	hexsum := hex.EncodeToString(md5sum[:])
	return fmt.Sprintf("__execution_run_shell_script_%s.sh", hexsum)
}
