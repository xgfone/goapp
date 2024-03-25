//go:build unix

package goapp

import (
	"log/slog"
	"math"
	"syscall"

	"github.com/xgfone/go-atexit"
)

func init() {
	atexit.OnInitWithPriority(math.MaxInt32, printRlimitNOFILE)
}

func printRlimitNOFILE() {
	var r syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &r)
	if err != nil {
		slog.Error("fail to get the nofile limit on unix", "err", err)
	} else {
		slog.Info("print nofile limit on unix", "cur", r.Cur, "max", r.Max)
	}
}
