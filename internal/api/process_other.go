//go:build !windows

package api

import (
	"os/exec"
	"syscall"
)

func hideChildWindow(cmd *exec.Cmd) {}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}
