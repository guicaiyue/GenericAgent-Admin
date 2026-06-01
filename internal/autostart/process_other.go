//go:build !windows

package autostart

import "os/exec"

func hideChildWindow(cmd *exec.Cmd) {}
