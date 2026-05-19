//go:build windows

package service

import (
	"encoding/csv"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func hideChildWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}

type processRow struct {
	pid         int
	commandLine string
}

func (m *Manager) stopConflictingService(s ServiceInfo) ([]int, error) {
	// Frontend singleton services such as fsapp.py may already be running outside Admin.
	// Force-restart only exact same GA-root script instances; never kill arbitrary python.
	if s.Kind != "frontend" || len(s.Command) < 2 {
		return nil, nil
	}
	script := filepath.Clean(filepath.Join(m.GARoot, filepath.FromSlash(s.Name)))
	if !strings.HasSuffix(strings.ToLower(script), ".py") {
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
		if !commandLineMatchesService(row.commandLine, m.GARoot, script, s.Command[0]) {
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
	ps := `$ErrorActionPreference='Stop'; Get-CimInstance Win32_Process | Where-Object { $_.CommandLine -and ($_.Name -match '^(python|pythonw)\.exe$' -or $_.CommandLine -match '(?i)python') } | Select-Object ProcessId,CommandLine | ConvertTo-Csv -NoTypeInformation`
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", ps)
	hideChildWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(strings.NewReader(string(out)))
	recs, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	var rows []processRow
	for i, rec := range recs {
		if i == 0 || len(rec) < 2 {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimSpace(rec[0]))
		if err != nil {
			continue
		}
		rows = append(rows, processRow{pid: pid, commandLine: rec[1]})
	}
	return rows, nil
}

func commandLineMatchesService(commandLine, gaRoot, script, pythonExe string) bool {
	cmd := strings.ToLower(normalizePathText(commandLine))
	root := strings.ToLower(normalizePathText(filepath.Clean(gaRoot)))
	scriptAbs := strings.ToLower(normalizePathText(filepath.Clean(script)))
	scriptRel := strings.ToLower(normalizePathText(filepath.ToSlash(mustRel(gaRoot, script))))
	py := strings.ToLower(normalizePathText(filepath.Clean(pythonExe)))
	base := strings.ToLower(filepath.Base(script))

	if !strings.Contains(cmd, base) {
		return false
	}
	if strings.Contains(cmd, scriptAbs) {
		return true
	}
	if strings.Contains(cmd, scriptRel) {
		// Most Admin-started services use a relative script path after setting Cmd.Dir=m.GARoot.
		// Treat the exact discovered relative script as conflicting even when WMI does not expose cwd.
		return true
	}
	if strings.Contains(cmd, base) && (strings.Contains(cmd, root) || strings.Contains(cmd, py)) {
		return true
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
