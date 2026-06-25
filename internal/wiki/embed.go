package wiki

import (
	"embed"
	"os"
)

// //go:embed wiki_ingest.py
var wikiIngestScript embed.FS

// ExtractWikiIngestScript 将嵌入的 wiki_ingest.py 写出到目标路径。
// 用于在运行时将脚本释放到 wikiDir/scripts/ 后由 agentmain --reflect 调度。
func ExtractWikiIngestScript(dest string) error {
	data, err := wikiIngestScript.ReadFile("wiki_ingest.py")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(parentDir(dest), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0o755)
}

func parentDir(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[:i]
		}
	}
	return "."
}