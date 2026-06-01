package api

import (
	"net/http"
	"os/exec"
	"runtime"

	"genericagent-admin-go/internal/hatchpet"
)

type hatchPetExportRequest struct {
	Path      string `json:"path"`
	Overwrite bool   `json:"overwrite"`
}

func (s *Server) hatchPetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		bad(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	st, err := hatchpet.StatusAt("")
	if err != nil {
		bad(w, 500, err.Error())
		return
	}
	writeJSON(w, st)
}

func (s *Server) hatchPetExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		bad(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req hatchPetExportRequest
	if err := decode(r, &req); err != nil {
		bad(w, 400, err.Error())
		return
	}
	st, err := hatchpet.Export(req.Path, req.Overwrite)
	if err != nil {
		bad(w, 500, err.Error())
		return
	}
	writeJSON(w, st)
}

func (s *Server) hatchPetOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		bad(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req hatchPetExportRequest
	if err := decode(r, &req); err != nil {
		bad(w, 400, err.Error())
		return
	}
	dir := req.Path
	if dir == "" {
		var err error
		dir, err = hatchpet.DefaultExportDir()
		if err != nil {
			bad(w, 500, err.Error())
			return
		}
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer.exe", dir)
	case "darwin":
		cmd = exec.Command("open", dir)
	default:
		cmd = exec.Command("xdg-open", dir)
	}
	if err := cmd.Start(); err != nil {
		bad(w, 500, err.Error())
		return
	}
	writeJSON(w, map[string]interface{}{"ok": true, "path": dir})
}
