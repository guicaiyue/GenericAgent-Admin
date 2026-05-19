//go:build !windows

package service

import "os/exec"

func hideChildWindow(cmd *exec.Cmd) {}

func (m *Manager) stopConflictingService(s ServiceInfo) ([]int, error) {
	return nil, nil
}
