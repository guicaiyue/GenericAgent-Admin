package wiki

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// IngestResult ingest 操作结果
type IngestResult struct {
	Status    string    `json:"status"` // idle, running, done, error
	StartedAt time.Time `json:"started_at,omitempty"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
	PID       int       `json:"pid,omitempty"`
	Error    string    `json:"error,omitempty"`
	Summary  string    `json:"summary,omitempty"`
	Files    int       `json:"files,omitempty"`
}

// StartIngest 启动异步 ingest 任务。
//
// 工作流程：
// 1. 检查 wikiDir 存在且包含 raw/ 目录
// 2. 写入 ingest state（status=running）
// 3. fork 子进程运行 reflect/wiki_ingest.py
//    - GA_ROOT 环境变量传入
//    - 将 prompt 输出到子进程 stdin
// 4. 返回子进程 PID
func StartIngest(wikiDir, gaRoot string, llmNo int) (*IngestResult, error) {
	if strings.TrimSpace(wikiDir) == "" {
		return nil, errors.New("wikiDir is empty")
	}
	if strings.TrimSpace(gaRoot) == "" {
		return nil, errors.New("gaRoot is empty")
	}
	rawDir := filepath.Join(wikiDir, "raw")
	if _, err := os.Stat(rawDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("raw directory does not exist: %s", rawDir)
	}

	// Step 1: SyncMemoryToRaw (plan step 2.3)
	if _, err := SyncMemoryToRaw(gaRoot, wikiDir); err != nil {
		return nil, fmt.Errorf("sync failed: %w", err)
	}

	// 加载当前状态
	state, err := loadIngestState(wikiDir)
	if err != nil {
		return nil, err
	}
	if state.Status == "running" {
		return nil, fmt.Errorf("ingest already running (pid=%d)", state.PID)
	}

	// 写入 ingest state
	state.Status = "running"
	state.StartedAt = time.Now()
	state.PID = 0
	state.Error = ""
	if err := saveIngestState(wikiDir, state); err != nil {
		return nil, err
	}

	// 准备子进程：优先使用 embed 提取的脚本
	scriptPath := filepath.Join(wikiDir, "scripts", "wiki_ingest.py")
	if err := ExtractWikiIngestScript(scriptPath); err != nil {
		// embed 失败，回退到 gaRoot/reflect/wiki_ingest.py
		scriptPath = filepath.Join(gaRoot, "reflect", "wiki_ingest.py")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			scriptPath = ""
		}
	}

	var cmd *exec.Cmd
	if scriptPath != "" {
		cmd = exec.Command("python3", scriptPath)
	} else {
		cmd = exec.Command("python3", "-c", generateIngestPrompt(wikiDir))
	}
	cmd.Dir = gaRoot

	// 设置环境变量
	env := os.Environ()
	env = append(env,
		"WIKI_DIR="+wikiDir,
		"RAW_DIR="+rawDir,
		"GA_ROOT="+gaRoot,
		"INGEST_STATE="+ingestStatePath(wikiDir),
		fmt.Sprintf("MODEL_NO=%d", llmNo),
		"PYTHONIOENCODING=utf-8",
		"PYTHONUTF8=1",
	)
	cmd.Env = env

	// 隐藏子进程窗口（Linux 下为空实现）
	hideChildWindow(cmd)

	if err := cmd.Start(); err != nil {
		state.Status = "error"
		state.Error = err.Error()
		saveIngestState(wikiDir, state)
		return state, err
	}

	state.PID = cmd.Process.Pid
	saveIngestState(wikiDir, state)
	return state, nil
}

// generateIngestPrompt 生成 ingest prompt
func generateIngestPrompt(wikiDir string) string {
	rawDir := filepath.Join(wikiDir, "raw")
	return fmt.Sprintf(`
请对 %s 执行 llm-wiki ingest 操作。
读取该目录下的所有源文件，使用 llm_wiki_skill SOP 
将其编译为交叉链接的 wiki 知识页面，写入 %s。
必须生成 %s/index.md 作为入口页。
`, rawDir, wikiDir, wikiDir)
}

// CheckIngestStatus 检查 ingest 任务状态
func CheckIngestStatus(wikiDir string) (*IngestResult, error) {
	state, err := loadIngestState(wikiDir)
	if err != nil {
		return nil, err
	}
	if state.Status == "running" && state.PID > 0 {
		// 检查进程是否存活
		if proc, err := os.FindProcess(state.PID); err == nil {
			if proc.Signal(syscall.Signal(0)) != nil {
				// 进程已退出
				state.Status = "done"
				state.FinishedAt = time.Now()
				saveIngestState(wikiDir, state)
			}
		}
	}
	return state, nil
}

// 生成 ingest state 文件路径
func ingestStatePath(wikiDir string) string {
	return filepath.Join(wikiDir, "ingest_state.json")
}

func loadIngestState(wikiDir string) (*IngestResult, error) {
	path := ingestStatePath(wikiDir)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &IngestResult{Status: "idle"}, nil
	}
	if err != nil {
		return nil, err
	}
	var state IngestResult
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func saveIngestState(wikiDir string, state *IngestResult) error {
	path := ingestStatePath(wikiDir)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// hideChildWindow 隐藏子进程窗口（Linux 空实现）
func hideChildWindow(cmd *exec.Cmd) {
	// Linux 下无操作
	_ = cmd
}