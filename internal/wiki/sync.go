package wiki

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// SyncResult 同步变更摘要
type SyncResult struct {
	Added    int      `json:"added"`
	Removed  int      `json:"removed"`
	Modified int      `json:"modified"`
	Errors   []string `json:"errors,omitempty"`
}

var syncMu sync.Mutex

// SyncMemoryToRaw 将 GA memory/ 同步到 wiki/raw/。
//
// 同步策略：
//   - 源目录：gaRoot/memory/ 下所有 .md 文件（含子目录）
//   - 目标：wikiDir/raw/，保持相对路径结构
//   - 新增/修改：覆盖写入
//   - 删除追踪：先扫源集合，再遍历 raw/ 删除不在集合中的文件
//   - 并发安全：加写锁
//   - 大文件（>10MB）记录警告但继续
func SyncMemoryToRaw(gaRoot, wikiDir string) (*SyncResult, error) {
	if strings.TrimSpace(gaRoot) == "" {
		return nil, errors.New("gaRoot is empty")
	}
	if strings.TrimSpace(wikiDir) == "" {
		return nil, errors.New("wikiDir is empty")
	}
	syncMu.Lock()
	defer syncMu.Unlock()

	result := &SyncResult{Errors: []string{}}
	memoryDir := filepath.Join(gaRoot, "memory")
	rawDir := filepath.Join(wikiDir, "raw")

	if err := os.MkdirAll(rawDir, 0o755); err != nil {
		return result, err
	}

	// 1) 收集源文件集合（相对路径 -> 源绝对路径）
	srcFiles := make(map[string]string) // rel -> abs
	if info, err := os.Stat(memoryDir); err == nil && info.IsDir() {
		err := filepath.WalkDir(memoryDir, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				result.Errors = append(result.Errors, "walk "+path+": "+walkErr.Error())
				return nil
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
				return nil
			}
			rel, err := filepath.Rel(memoryDir, path)
			if err != nil {
				return nil
			}
			srcFiles[filepath.ToSlash(rel)] = path
			return nil
		})
		if err != nil {
			return result, err
		}
	}

	// 2) 复制新增/修改
	for rel, src := range srcFiles {
		dst := filepath.Join(rawDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			result.Errors = append(result.Errors, "mkdir "+filepath.Dir(dst)+": "+err.Error())
			continue
		}
		stat, statErr := os.Stat(src)
		existed := false
		if statErr == nil && stat.Size() > 10*1024*1024 {
			result.Errors = append(result.Errors, "warn: large file "+rel+" >10MB")
		}
		if dstStat, err := os.Stat(dst); err == nil {
			existed = true
			if statErr == nil && dstStat.Size() == stat.Size() && dstStat.ModTime().Equal(stat.ModTime()) {
				continue
			}
		}
		if err := copyFile(src, dst); err != nil {
			result.Errors = append(result.Errors, "copy "+rel+": "+err.Error())
			continue
		}
		if existed {
			result.Modified++
		} else {
			result.Added++
		}
	}

	// 3) 删除 raw/ 中不再存在的源文件
	if err := filepath.WalkDir(rawDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(rawDir, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if _, ok := srcFiles[rel]; !ok {
			if err := os.Remove(path); err == nil {
				result.Removed++
			} else {
				result.Errors = append(result.Errors, "remove "+rel+": "+err.Error())
			}
		}
		return nil
	}); err != nil {
		result.Errors = append(result.Errors, "walk raw: "+err.Error())
	}

	return result, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	if stat, err := os.Stat(src); err == nil {
		_ = os.Chtimes(dst, stat.ModTime(), stat.ModTime())
	}
	return nil
}
