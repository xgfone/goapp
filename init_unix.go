//go:build unix

package goapp

import (
	"log/slog"
	"syscall"

	"github.com/xgfone/go-defaults"
)

func init() {
	defaults.OnInit(printRlimitNOFILE)
}

func printRlimitNOFILE() {
	var r syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &r)
	if err != nil {
		slog.Error("fail to get the nofile limit on unix", "err", err)
	} else {
		slog.Debug("print nofile limit on unix", "cur", r.Cur, "max", r.Max)
	}
}
