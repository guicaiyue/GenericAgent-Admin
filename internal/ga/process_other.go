//go:build !windows

package ga

import "os/exec"

func hideChildWindow(cmd *exec.Cmd) {}
