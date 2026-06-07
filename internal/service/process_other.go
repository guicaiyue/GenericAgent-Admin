//go:build !windows

package service

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func hideChildWindow(cmd *exec.Cmd) {}

type processRow struct {
	pid         int
	commandLine string
	workingDir  string
}

func (m *Manager) stopConflictingService(s ServiceInfo) ([]int, error) {
	// Admin is the authority for all GA services. Before starting a managed service,
	// terminate any already-running external instance that matches the discovered
	// GA-root command/script, then restart it under Admin so PID/status/logs are tracked.
	if len(s.Command) < 2 {
		return nil, nil
	}
	rows, err := listPythonProcesses()
	if err != nil {
		return nil, err
	}
	self := os.Getpid()
	var killed []int
	for _, row := range rows {
		if row.pid <= 0 || row.pid == self {
			continue
		}
		if !commandLineMatchesService(row.commandLine, row.workingDir, m.GARoot, s.Command) {
			continue
		}
		p, err := os.FindProcess(row.pid)
		if err != nil {
			continue
		}
		if err := p.Kill(); err != nil {
			return killed, err
		}
		killed = append(killed, row.pid)
	}
	return killed, nil
}

func listPythonProcesses() ([]processRow, error) {
	out, err := exec.Command("ps", "-eo", "pid=,args=").Output()
	if err != nil {
		return nil, err
	}
	var rows []processRow
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		cmdline := strings.TrimSpace(strings.TrimPrefix(line, fields[0]))
		lower := strings.ToLower(cmdline)
		if !strings.Contains(lower, "python") {
			continue
		}
		cwd, _ := os.Readlink(filepath.Join("/proc", strconv.Itoa(pid), "cwd"))
		rows = append(rows, processRow{pid: pid, commandLine: cmdline, workingDir: cwd})
	}
	return rows, nil
}

func commandLineMatchesService(commandLine, workingDir, gaRoot string, serviceCommand []string) bool {
	if len(serviceCommand) < 2 {
		return false
	}
	cmd := strings.ToLower(normalizePathText(commandLine))
	root := strings.ToLower(normalizePathText(filepath.Clean(gaRoot)))
	cwd := strings.ToLower(normalizePathText(filepath.Clean(workingDir)))
	py := strings.ToLower(normalizePathText(filepath.Clean(serviceCommand[0])))
	inGARoot := cwd == root

	// Require either the configured GA python executable, the GA root path in the
	// command line, or cwd equal to GA root. The cwd exception lets Admin recognize
	// existing services started as relative commands from the GA root without also
	// matching unrelated python processes in other directories.
	if !strings.Contains(cmd, py) && !strings.Contains(cmd, root) && !inGARoot {
		return false
	}

	for _, arg := range serviceCommand[1:] {
		if strings.HasPrefix(arg, "-") {
			if !strings.Contains(cmd, strings.ToLower(arg)) {
				return false
			}
			continue
		}
		if !strings.HasSuffix(strings.ToLower(arg), ".py") && !strings.HasSuffix(strings.ToLower(arg), ".pyw") {
			continue
		}
		if !commandLineContainsScript(cmd, gaRoot, arg) {
			return false
		}
	}
	return true
}

func commandLineContainsScript(cmd, gaRoot, scriptArg string) bool {
	script := filepath.Clean(filepath.Join(gaRoot, filepath.FromSlash(scriptArg)))
	scriptAbs := strings.ToLower(normalizePathText(script))
	scriptRel := strings.ToLower(normalizePathText(filepath.ToSlash(mustRel(gaRoot, script))))
	return commandLineContainsPathToken(cmd, scriptAbs) || commandLineContainsPathToken(cmd, scriptRel)
}

func commandLineContainsPathToken(cmd, token string) bool {
	if token == "" {
		return false
	}
	for _, field := range strings.Fields(cmd) {
		field = normalizePathText(strings.Trim(field, `'"`))
		if field == token {
			return true
		}
	}
	return false
}

func normalizePathText(s string) string {
	s = strings.ReplaceAll(s, "\\", "/")
	for strings.Contains(s, "//") {
		s = strings.ReplaceAll(s, "//", "/")
	}
	return strings.Trim(s, " \t\r\n\"")
}

func mustRel(base, target string) string {
	if rel, err := filepath.Rel(base, target); err == nil {
		return rel
	}
	return target
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil
}
