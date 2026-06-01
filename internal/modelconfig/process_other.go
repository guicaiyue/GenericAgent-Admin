//go:build !windows

package modelconfig

import "os/exec"

func hideChildWindow(cmd *exec.Cmd) {}
