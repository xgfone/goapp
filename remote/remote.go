// Copyright 2021~2022 xgfone
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

// Package remote provides the command execution functions on the remote host.
package remote

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xgfone/go-exec"
)

// SSHUser is the user to execute the shell command by SSH.
var SSHUser = "root"

// SSHOptions is the options of ssh/scp command.
var SSHOptions = "-o StrictHostKeyChecking=no"

// ExecuteCmdBySSHContext executes the command on the remote host by SSH.
func ExecuteCmdBySSHContext(ctx context.Context, remoteHost, cmd string) (
	stdout, stderr string, err error) {
	cmd = fmt.Sprintf(`ssh %s %s@%s "%s"`, SSHOptions, SSHUser, remoteHost, cmd)
	_stdout, _stderr, err := exec.RunShellCmd(ctx, cmd)
	if err == nil {
		stdout, stderr = string(_stdout), string(_stderr)
	}
	return
}

// CopyFilesToRemoteBySSHContext copies the files from the local to the remote.
func CopyFilesToRemoteBySSHContext(ctx context.Context, remoteHost,
	remoteDirOrFile string, localFiles ...string) error {
	if len(localFiles) == 0 {
		return nil
	}

	files := strings.Join(localFiles, " ")
	cmd := fmt.Sprintf("scp %s %s %s@%s:%s", SSHOptions, files, SSHUser, remoteHost, remoteDirOrFile)
	_, _, err := exec.RunShellCmd(ctx, cmd)
	return err
}

// CopyFilesFromRemoteBySSHContext copies the files from the remote to the local.
func CopyFilesFromRemoteBySSHContext(ctx context.Context, remoteHost,
	localDirOrFile string, remoteFiles ...string) error {
	if len(remoteFiles) == 0 {
		return nil
	}

	for i, file := range remoteFiles {
		remoteFiles[i] = fmt.Sprintf("%s@%s:%s", SSHUser, remoteHost, file)
	}

	files := strings.Join(remoteFiles, " ")
	cmd := fmt.Sprintf("scp %s %s %s", SSHOptions, files, localDirOrFile)
	_, _, err := exec.RunShellCmd(ctx, cmd)
	return err
}

// ExecuteCmdBySSH is equal to
//
//	ExecuteCmdBySSHContext(context.Background(), remoteHost, cmd)
func ExecuteCmdBySSH(remoteHost, cmd string) (stdout, stderr string, err error) {
	return ExecuteCmdBySSHContext(context.Background(), remoteHost, cmd)
}

// CopyFilesToRemoteBySSH is equal to
//
//	CopyFilesToRemoteBySSHContext(context.Background(), remoteHost,
//	                              remoteDirOrFile, localFiles...)
func CopyFilesToRemoteBySSH(remoteHost, remoteDirOrFile string, localFiles ...string) error {
	return CopyFilesToRemoteBySSHContext(context.Background(), remoteHost,
		remoteDirOrFile, localFiles...)
}

// CopyFilesFromRemoteBySSH is equal to
//
//	CopyFilesFromRemoteBySSHContext(context.Background(), remoteHost,
//	                                localDirOrFile, remoteFiles...)
func CopyFilesFromRemoteBySSH(remoteHost, localDirOrFile string, remoteFiles ...string) error {
	return CopyFilesFromRemoteBySSHContext(context.Background(), remoteHost,
		localDirOrFile, remoteFiles...)
}

// ExecuteScriptBySSH is equal to
//
//	ExecuteScriptBySSHContext(context.Background(), remoteHost, script)
func ExecuteScriptBySSH(remoteHost, script string) (stdout, stderr string, err error) {
	return ExecuteScriptBySSHContext(context.Background(), remoteHost, script)
}

// ExecuteScriptBySSHContext executes the shell script by SSH.
func ExecuteScriptBySSHContext(ctx context.Context, remoteHost, script string) (
	stdout, stderr string, err error) {

	localfile, err := exec.GetScriptFile("", script)
	if err != nil {
		return
	}
	defer os.Remove(localfile)

	remotefile := localfile
	if dir := filepath.Dir(localfile); dir != "" {
		remotefile = filepath.Join(dir, localfile)
	}

	err = CopyFilesToRemoteBySSHContext(ctx, remoteHost, remotefile, localfile)
	if err != nil {
		return
	}
	defer ExecuteCmdBySSHContext(ctx, remoteHost, fmt.Sprintf("rm -f %s", remotefile))

	shell := exec.DefaultCmd.Shell
	if shell == "" {
		if shell = exec.DefaultShell; shell == "" {
			shell = "bash"
		}
	}

	return ExecuteCmdBySSHContext(ctx, remoteHost, fmt.Sprintf("%s %s", shell, remotefile))
}

func getScriptFile(script string) (filename string) {
	data := []byte(script)
	md5sum := md5.Sum(data)
	hexsum := hex.EncodeToString(md5sum[:])
	return fmt.Sprintf("__execution_run_shell_script_%s.sh", hexsum)
}
