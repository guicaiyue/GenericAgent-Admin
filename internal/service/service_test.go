package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverExcludesGoalModeFromGenericReflectList(t *testing.T) {
	root := t.TempDir()
	reflectDir := filepath.Join(root, "reflect")
	if err := os.MkdirAll(reflectDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"goal_mode.py", "autonomous.py", "custom_reflect.py"} {
		if err := os.WriteFile(filepath.Join(reflectDir, name), []byte("# test\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	items := NewManager(root, 100).Discover()
	seen := map[string]int{}
	for _, item := range items {
		seen[item.Name]++
	}
	if seen[filepath.ToSlash(filepath.Join("reflect", "goal_mode.py"))] != 1 {
		t.Fatalf("goal_mode.py should appear exactly once as the dedicated lifecycle entry, seen=%d items=%#v", seen[filepath.ToSlash(filepath.Join("reflect", "goal_mode.py"))], items)
	}
	if seen[filepath.ToSlash(filepath.Join("reflect", "autonomous.py"))] != 1 {
		t.Fatalf("autonomous.py should remain discoverable once, seen=%d", seen[filepath.ToSlash(filepath.Join("reflect", "autonomous.py"))])
	}
	if seen[filepath.ToSlash(filepath.Join("reflect", "custom_reflect.py"))] != 1 {
		t.Fatalf("custom reflect should remain discoverable once, seen=%d", seen[filepath.ToSlash(filepath.Join("reflect", "custom_reflect.py"))])
	}
}

func TestDiscoverSchedulerUsesReflectEntrypoint(t *testing.T) {
	root := t.TempDir()
	reflectDir := filepath.Join(root, "reflect")
	if err := os.MkdirAll(reflectDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(reflectDir, "scheduler.py"), []byte("# test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	items := NewManager(root, 100).Discover()
	seen := map[string]ServiceInfo{}
	for _, item := range items {
		seen[item.Name] = item
	}
	rel := filepath.ToSlash(filepath.Join("reflect", "scheduler.py"))
	item, ok := seen[rel]
	if !ok {
		t.Fatalf("missing scheduler service in %#v", items)
	}
	if item.Kind != "reflect" {
		t.Fatalf("scheduler kind=%s", item.Kind)
	}
	want := []string{"python", "agentmain.py", "--reflect", rel}
	if len(item.Command) != len(want) {
		t.Fatalf("scheduler command=%#v want %#v", item.Command, want)
	}
	for i := range want {
		if item.Command[i] != want[i] {
			t.Fatalf("scheduler command[%d]=%q want %q in %#v", i, item.Command[i], want[i], item.Command)
		}
	}
}

func TestDiscoverIncludesChannelFrontendApps(t *testing.T) {
	root := t.TempDir()
	frontendsDir := filepath.Join(root, "frontends")
	if err := os.MkdirAll(frontendsDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"fsapp.py", "wecomapp.py", "dingtalkapp.py", "notbot.py", "_hiddenapp.py"} {
		if err := os.WriteFile(filepath.Join(frontendsDir, name), []byte("# test\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	items := NewManager(root, 100).Discover()
	seen := map[string]ServiceInfo{}
	for _, item := range items {
		seen[item.Name] = item
	}
	for _, name := range []string{"fsapp.py", "wecomapp.py", "dingtalkapp.py"} {
		rel := filepath.ToSlash(filepath.Join("frontends", name))
		item, ok := seen[rel]
		if !ok {
			t.Fatalf("missing channel frontend %s in %#v", rel, items)
		}
		if item.Kind != "frontend" {
			t.Fatalf("%s kind=%s", rel, item.Kind)
		}
		if len(item.Command) != 2 || item.Command[1] != rel {
			t.Fatalf("%s command=%#v", rel, item.Command)
		}
	}
	if _, ok := seen[filepath.ToSlash(filepath.Join("frontends", "notbot.py"))]; ok {
		t.Fatalf("notbot.py should not be discovered: %#v", items)
	}
	if _, ok := seen[filepath.ToSlash(filepath.Join("frontends", "_hiddenapp.py"))]; ok {
		t.Fatalf("_hiddenapp.py should not be discovered: %#v", items)
	}
}

func TestCommandLineMatchesServiceRequiresExactScriptPath(t *testing.T) {
	root := filepath.Clean(filepath.Join(t.TempDir(), "ga-root"))
	py := filepath.Join(root, ".venv", "Scripts", "python.exe")
	cmd := []string{py, filepath.ToSlash(filepath.Join("reflect", "custom_reflect.py"))}

	ownAbs := py + " " + filepath.Join(root, "reflect", "custom_reflect.py")
	if !commandLineMatchesService(ownAbs, "", root, cmd) {
		t.Fatalf("expected absolute GA script command to match")
	}

	ownRel := py + " reflect/custom_reflect.py"
	if !commandLineMatchesService(ownRel, "", root, cmd) {
		t.Fatalf("expected relative GA script command to match")
	}

	relFromRoot := "python agentmain.py --reflect reflect/custom_reflect.py"
	if !commandLineMatchesService(relFromRoot, root, root, []string{py, "agentmain.py", "--reflect", "reflect/custom_reflect.py"}) {
		t.Fatalf("expected relative command from GA root to match")
	}
	if commandLineMatchesService(relFromRoot, filepath.Dir(root), root, []string{py, "agentmain.py", "--reflect", "reflect/custom_reflect.py"}) {
		t.Fatalf("relative command outside GA root must not match")
	}

	otherSameBase := py + " " + filepath.Join(filepath.Dir(root), "other-root", "reflect", "custom_reflect.py")
	if commandLineMatchesService(otherSameBase, "", root, cmd) {
		t.Fatalf("same basename under another root must not match")
	}

	otherSimilarRel := py + " other/reflect/custom_reflect.py"
	if commandLineMatchesService(otherSimilarRel, "", root, cmd) {
		t.Fatalf("relative path with extra prefix must not match")
	}
}

func TestCommandWithLLMNoAppendsOrReplacesWithoutMutatingOriginal(t *testing.T) {
	base := []string{"python", "agentmain.py", "--reflect", "reflect/autonomous.py"}
	withLLM := commandWithLLMNo(base, 2)
	want := []string{"python", "agentmain.py", "--reflect", "reflect/autonomous.py", "--llm_no", "2"}
	if len(withLLM) != len(want) {
		t.Fatalf("command length=%d want=%d: %#v", len(withLLM), len(want), withLLM)
	}
	for i := range want {
		if withLLM[i] != want[i] {
			t.Fatalf("command[%d]=%q want %q in %#v", i, withLLM[i], want[i], withLLM)
		}
	}
	if len(base) != 4 {
		t.Fatalf("base command mutated: %#v", base)
	}

	existing := []string{"python", "agentmain.py", "--reflect", "reflect/autonomous.py", "--llm_no", "1"}
	replaced := commandWithLLMNo(existing, 3)
	if replaced[len(replaced)-1] != "3" {
		t.Fatalf("existing llm_no was not replaced: %#v", replaced)
	}
	if existing[len(existing)-1] != "1" {
		t.Fatalf("existing command mutated: %#v", existing)
	}
}

func TestStartWithLLMRejectsInvalidOverridesBeforeStarting(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "reflect"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "reflect", "autonomous.py"), []byte("# test\n"), 0644); err != nil {
		t.Fatal(err)
	}
	frontendsDir := filepath.Join(root, "frontends")
	if err := os.MkdirAll(frontendsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(frontendsDir, "fsapp.py"), []byte("# test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m := NewManager(root, 100)
	negative := -1
	if _, err := m.StartWithLLM(filepath.ToSlash(filepath.Join("reflect", "autonomous.py")), &negative); err == nil || err.Error() != "llm_no must be non-negative" {
		t.Fatalf("negative llm_no err=%v", err)
	}
	llm := 2
	if _, err := m.StartWithLLM(filepath.ToSlash(filepath.Join("frontends", "fsapp.py")), &llm); err == nil || err.Error() != "llm_no is only supported for reflect services" {
		t.Fatalf("frontend llm_no err=%v", err)
	}
}
