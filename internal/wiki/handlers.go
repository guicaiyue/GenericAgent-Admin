package wiki

import (
	"encoding/json"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// IngestState tracks the async ingest task status.
type IngestState struct {
	Status    string    `json:"status"` // idle, running, done, error
	PID       int       `json:"pid,omitempty"`
	StartedAt time.Time `json:"started_at,omitempty"`
	EndedAt   time.Time `json:"ended_at,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// WikiFile represents a file in the wiki directory.
type WikiFile struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	MTime  string `json:"mtime"`
}

// IngestRequest is the request body for wiki ingest.
type IngestRequest struct {
	LLMNo int `json:"llm_no"`
}

func stateFile(wikiDir string) string {
	return filepath.Join(wikiDir, "ingest_state.json")
}

// LoadState reads ingest state from disk.
func LoadState(wikiDir string) *IngestState {
	data, err := os.ReadFile(stateFile(wikiDir))
	if err != nil {
		return &IngestState{Status: "idle"}
	}
	var s IngestState
	if err := json.Unmarshal(data, &s); err != nil {
		return &IngestState{Status: "idle"}
	}
	return &s
}

// SaveState persists ingest state to disk.
func SaveState(wikiDir string, s *IngestState) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(stateFile(wikiDir), data, 0644)
}

// IsProcessAlive checks if a process with the given PID is still running.
func IsProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

// EnsureDir creates dir if it doesn't exist.
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// ListFiles recursively lists files under a directory.
func ListFiles(root string) ([]WikiFile, error) {
	var out []WikiFile
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out = append(out, WikiFile{
			Name:  info.Name(),
			Path:  rel,
			Size:  info.Size(),
			MTime: info.ModTime().UTC().Format(time.RFC3339),
		})
		return nil
	})
	return out, err
}

// SyncMemory walks GA memory dir and counts files not yet in raw/.
func SyncMemory(memoryDir, rawDir string) (*SyncResult, error) {
	if err := EnsureDir(rawDir); err != nil {
		return nil, err
	}
	result := &SyncResult{}
	err := filepath.Walk(memoryDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(memoryDir, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(rawDir, rel)
		if _, err := os.Stat(dest); os.IsNotExist(err) {
			result.Added++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
