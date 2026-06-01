package hatchpet

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	SkillName    = "hatch-pet"
	embeddedRoot = "skill"
)

// Skill embeds the complete Codex hatch-pet skill so GA Admin builds remain self-contained.
//
//go:embed skill
var Skill embed.FS

type FileEntry struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type Status struct {
	SkillName        string      `json:"skill_name"`
	EmbeddedRoot     string      `json:"embedded_root"`
	DefaultExportDir string      `json:"default_export_dir"`
	Exported         bool        `json:"exported"`
	Files            int         `json:"files"`
	Bytes            int64       `json:"bytes"`
	Manifest         []FileEntry `json:"manifest,omitempty"`
	Missing          []string    `json:"missing,omitempty"`
}

func DefaultExportDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex", "skills", SkillName), nil
}

func FS() (fs.FS, error) { return fs.Sub(Skill, embeddedRoot) }

func Manifest() ([]FileEntry, int64, error) {
	var entries []FileEntry
	var total int64
	err := fs.WalkDir(Skill, embeddedRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, embeddedRoot), "/")
		entries = append(entries, FileEntry{Path: rel, Size: info.Size()})
		total += info.Size()
		return nil
	})
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, total, err
}

func StatusAt(dest string) (Status, error) {
	if dest == "" {
		var err error
		dest, err = DefaultExportDir()
		if err != nil {
			return Status{}, err
		}
	}
	manifest, bytes, err := Manifest()
	if err != nil {
		return Status{}, err
	}
	st := Status{SkillName: SkillName, EmbeddedRoot: embeddedRoot, DefaultExportDir: dest, Files: len(manifest), Bytes: bytes, Manifest: manifest}
	for _, entry := range manifest {
		if _, err := os.Stat(filepath.Join(dest, filepath.FromSlash(entry.Path))); err != nil {
			st.Missing = append(st.Missing, entry.Path)
		}
	}
	st.Exported = len(st.Missing) == 0 && len(manifest) > 0
	return st, nil
}

func Export(dest string, overwrite bool) (Status, error) {
	if dest == "" {
		var err error
		dest, err = DefaultExportDir()
		if err != nil {
			return Status{}, err
		}
	}
	cleanDest := filepath.Clean(dest)
	if cleanDest == string(filepath.Separator) || cleanDest == "." {
		return Status{}, fmt.Errorf("refusing to export hatch-pet to unsafe path %q", dest)
	}
	err := fs.WalkDir(Skill, embeddedRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, embeddedRoot), "/")
		if rel == "" {
			return os.MkdirAll(cleanDest, 0o755)
		}
		out := filepath.Join(cleanDest, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}
		if !overwrite {
			if _, err := os.Stat(out); err == nil {
				return nil
			} else if !os.IsNotExist(err) {
				return err
			}
		}
		data, err := fs.ReadFile(Skill, p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		return os.WriteFile(out, data, 0o644)
	})
	if err != nil {
		return Status{}, err
	}
	return StatusAt(cleanDest)
}
