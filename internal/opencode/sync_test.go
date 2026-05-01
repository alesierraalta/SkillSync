package opencode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skillsync/tui/internal/types"
)

// ensureDir creates all directories in path.
func ensureDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

// ----------------------------------------------------------------
// TestSyncSkills_MirrorsFile — Task 1.1
// ----------------------------------------------------------------

func TestSyncSkills_MirrorsFile(t *testing.T) {
	tmp := t.TempDir()
	
	agentsDir := filepath.Join(tmp, ".agents", "skills")
	opencodeDir := filepath.Join(tmp, ".opencode")
	ensureDir(t, agentsDir)
	ensureDir(t, opencodeDir)

	fooDir := filepath.Join(agentsDir, "foo")
	ensureDir(t, fooDir)
	fooSrc := filepath.Join(fooDir, "SKILL.md")
	fooContent := "# Foo Skill\n---\nname: foo\ndescription: A test skill\n"
	if err := os.WriteFile(fooSrc, []byte(fooContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := SyncSkills(tmp, Options{})
	if err != nil {
		t.Fatalf("SyncSkills failed: %v", err)
	}

	fooDest := filepath.Join(tmp, ".opencode", "skills", "foo", "SKILL.md")
	if _, err := os.Stat(fooDest); os.IsNotExist(err) {
		t.Fatalf("expected %s to exist", fooDest)
	}

	got, _ := os.ReadFile(fooDest)
	if string(got) != fooContent {
		t.Errorf("content mismatch:\nwant: %s\ngot:  %s", fooContent, got)
	}
}

// ----------------------------------------------------------------
// TestSyncSkills_SkipsUnchanged — Task 1.2
// ----------------------------------------------------------------

func TestSyncSkills_SkipsUnchanged(t *testing.T) {
	tmp := t.TempDir()
	
	agentsDir := filepath.Join(tmp, ".agents", "skills")
	opencodeDir := filepath.Join(tmp, ".opencode")
	ensureDir(t, agentsDir)
	ensureDir(t, opencodeDir)

	fooDir := filepath.Join(agentsDir, "foo")
	ensureDir(t, fooDir)
	fooSrc := filepath.Join(fooDir, "SKILL.md")
	fooContent := "name: foo\ndescription: unchanged\n"
	if err := os.WriteFile(fooSrc, []byte(fooContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := SyncSkills(tmp, Options{})
	if err != nil {
		t.Fatalf("first SyncSkills failed: %v", err)
	}

	fooDest := filepath.Join(tmp, ".opencode", "skills", "foo", "SKILL.md")
	stat1, err := os.Stat(fooDest)
	if err != nil {
		t.Fatalf("dest file not created: %v", err)
	}
	mtime1 := stat1.ModTime()

	err = SyncSkills(tmp, Options{})
	if err != nil {
		t.Fatalf("second SyncSkills failed: %v", err)
	}

	stat2, err := os.Stat(fooDest)
	if err != nil {
		t.Fatalf("dest file missing after second sync: %v", err)
	}
	mtime2 := stat2.ModTime()

	if !mtime1.Equal(mtime2) {
		t.Error("expected second sync to skip writing unchanged file")
	}
}

// ----------------------------------------------------------------
// TestSyncSkills_PreservesSymlink — Task 1.3
// ----------------------------------------------------------------

func TestSyncSkills_PreservesSymlink(t *testing.T) {
	tmp := t.TempDir()
	
	agentsDir := filepath.Join(tmp, ".agents", "skills")
	opencodeDir := filepath.Join(tmp, ".opencode")
	ensureDir(t, agentsDir)
	ensureDir(t, opencodeDir)

	// Create target directory and file
	targetDir := filepath.Join(tmp, "shared")
	ensureDir(t, targetDir)
	targetFile := filepath.Join(targetDir, "bar.md")
	if err := os.WriteFile(targetFile, []byte("shared content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create symlink in .agents/skills/bar/SKILL.md
	// Use an absolute path or path relative to the symlink location
	barSrcDir := filepath.Join(agentsDir, "bar")
	ensureDir(t, barSrcDir)
	barSrc := filepath.Join(barSrcDir, "SKILL.md")
	
	// Create the symlink - on Windows we need to use the actual resolved path
	// because relative symlinks can be tricky with EvalSymlinks
	linkTarget := targetFile // absolute path to target
	if err := os.Symlink(linkTarget, barSrc); err != nil {
		t.Skipf("symlink creation failed (may need admin on Windows): %v", err)
	}

	err := SyncSkills(tmp, Options{})
	if err != nil {
		t.Fatalf("SyncSkills failed: %v", err)
	}

	barDest := filepath.Join(tmp, ".opencode", "skills", "bar", "SKILL.md")
	info, err := os.Lstat(barDest)
	if err != nil {
		t.Fatalf("expected symlink to exist at %s: %v", barDest, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink, got mode %v", info.Mode())
	}

	readLink, err := os.Readlink(barDest)
	if err != nil {
		t.Fatal(err)
	}
	// Compare normalized paths
	if readLink != linkTarget && filepath.ToSlash(readLink) != filepath.ToSlash(linkTarget) {
		t.Errorf("symlink target mismatch: want %s, got %s", linkTarget, readLink)
	}
}

// ----------------------------------------------------------------
// TestSyncSkills_CreatesMissingDir — Task 1.4
// ----------------------------------------------------------------

func TestSyncSkills_CreatesMissingDir(t *testing.T) {
	tmp := t.TempDir()
	
	agentsDir := filepath.Join(tmp, ".agents", "skills")
	opencodeDir := filepath.Join(tmp, ".opencode")
	ensureDir(t, agentsDir)
	ensureDir(t, opencodeDir)

	fooDir := filepath.Join(agentsDir, "foo")
	ensureDir(t, fooDir)
	fooSrc := filepath.Join(fooDir, "SKILL.md")
	if err := os.WriteFile(fooSrc, []byte("name: foo\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := SyncSkills(tmp, Options{})
	if err != nil {
		t.Fatalf("SyncSkills failed: %v", err)
	}

	fooDest := filepath.Join(tmp, ".opencode", "skills", "foo", "SKILL.md")
	if _, err := os.Stat(fooDest); os.IsNotExist(err) {
		t.Fatalf("expected %s to exist after first-run", fooDest)
	}
}

// ----------------------------------------------------------------
// TestSyncSkills_PruneRemovesOrphans — Task 1.5
// ----------------------------------------------------------------

func TestSyncSkills_PruneRemovesOrphans(t *testing.T) {
	tmp := t.TempDir()
	
	agentsDir := filepath.Join(tmp, ".agents", "skills")
	opencodeSkillsDir := filepath.Join(tmp, ".opencode", "skills")
	ensureDir(t, agentsDir)
	ensureDir(t, opencodeSkillsDir)

	orphanDir := filepath.Join(opencodeSkillsDir, "orphan")
	ensureDir(t, orphanDir)
	if err := os.WriteFile(filepath.Join(orphanDir, "SKILL.md"), []byte("orphan"), 0644); err != nil {
		t.Fatal(err)
	}

	fooDir := filepath.Join(agentsDir, "foo")
	ensureDir(t, fooDir)
	fooSrc := filepath.Join(fooDir, "SKILL.md")
	if err := os.WriteFile(fooSrc, []byte("name: foo\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := SyncSkills(tmp, Options{})
	if err != nil {
		t.Fatalf("SyncSkills without prune failed: %v", err)
	}
	if _, err := os.Stat(orphanDir); os.IsNotExist(err) {
		t.Error("orphan should NOT be deleted without --prune")
	}

	err = SyncSkills(tmp, Options{Prune: true})
	if err != nil {
		t.Fatalf("SyncSkills with prune failed: %v", err)
	}
	if _, err := os.Stat(orphanDir); !os.IsNotExist(err) {
		t.Error("orphan should be deleted with --prune")
	}
}

// ----------------------------------------------------------------
// TestSyncSkills_DryRunReportsNoWrite — Task 1.6
// ----------------------------------------------------------------

func TestSyncSkills_DryRunReportsNoWrite(t *testing.T) {
	tmp := t.TempDir()
	
	agentsDir := filepath.Join(tmp, ".agents", "skills")
	opencodeDir := filepath.Join(tmp, ".opencode")
	ensureDir(t, agentsDir)
	ensureDir(t, opencodeDir)

	fooDir := filepath.Join(agentsDir, "foo")
	ensureDir(t, fooDir)
	fooSrc := filepath.Join(fooDir, "SKILL.md")
	fooContent := "name: foo\n"
	if err := os.WriteFile(fooSrc, []byte(fooContent), 0644); err != nil {
		t.Fatal(err)
	}

	SyncSkills(tmp, Options{DryRun: true})

	fooDest := filepath.Join(tmp, ".opencode", "skills", "foo", "SKILL.md")
	if _, err := os.Stat(fooDest); !os.IsNotExist(err) {
		t.Error("dry-run should not write files")
	}
}

// ----------------------------------------------------------------
// TestRegenerateTools_GeneratesTools — Task 1.7
// ----------------------------------------------------------------

func TestRegenerateTools_GeneratesTools(t *testing.T) {
	tmp := t.TempDir()
	ensureDir(t, filepath.Join(tmp, ".opencode"))
	pkgPath := filepath.Join(tmp, ".opencode", "package.json")
	initPackageJSON(t, tmp)

	skills := []types.Skill{
		{Name: "foo", Metadata: types.Metadata{AutoInvoke: true, Description: "Foo skill"}},
		{Name: "bar", Metadata: types.Metadata{AutoInvoke: true, Description: "Bar skill"}},
		{Name: "baz", Metadata: types.Metadata{AutoInvoke: false, Description: "Baz skill"}},
	}

	err := RegenerateTools(tmp, skills, false)
	if err != nil {
		t.Fatalf("RegenerateTools failed: %v", err)
	}

	data, err := os.ReadFile(pkgPath)
	if err != nil {
		t.Fatal(err)
	}
	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		t.Fatal(err)
	}

	oc := pkg["opencode"].(map[string]interface{})
	tools := oc["tools"].([]interface{})

	if len(tools) != 0 {
		t.Errorf("expected 0 tools (handled by markdown), got %d", len(tools))
	}
}

// ----------------------------------------------------------------
// TestRegenerateTools_PreservesConfig — Task 1.8
// ----------------------------------------------------------------

func TestRegenerateTools_PreservesConfig(t *testing.T) {
	tmp := t.TempDir()
	ensureDir(t, filepath.Join(tmp, ".opencode"))
	pkgPath := filepath.Join(tmp, ".opencode", "package.json")

	customPkg := `{
  "name": "my-custom-project",
  "version": "1.0.0",
  "opencode": {
    "tools": [{"name": "old-tool", "command": "echo old"}]
  }
}`
	if err := os.WriteFile(pkgPath, []byte(customPkg), 0644); err != nil {
		t.Fatal(err)
	}

	skills := []types.Skill{
		{Name: "foo", Metadata: types.Metadata{AutoInvoke: true}},
	}

	err := RegenerateTools(tmp, skills, false)
	if err != nil {
		t.Fatalf("RegenerateTools failed: %v", err)
	}

	data, _ := os.ReadFile(pkgPath)
	var pkg map[string]interface{}
	json.Unmarshal(data, &pkg)

	if pkg["name"] != "my-custom-project" {
		t.Error("custom 'name' field was overwritten")
	}
	if pkg["version"] != "1.0.0" {
		t.Error("custom 'version' field was overwritten")
	}
}

// ----------------------------------------------------------------
// TestRegenerateTools_DryRun — Task 1.9
// ----------------------------------------------------------------

func TestRegenerateTools_DryRun(t *testing.T) {
	tmp := t.TempDir()
	ensureDir(t, filepath.Join(tmp, ".opencode"))
	pkgPath := filepath.Join(tmp, ".opencode", "package.json")
	initPackageJSON(t, tmp)

	originalContent, _ := os.ReadFile(pkgPath)

	skills := []types.Skill{
		{Name: "foo", Metadata: types.Metadata{AutoInvoke: true}},
	}

	err := RegenerateTools(tmp, skills, true)
	if err != nil {
		t.Fatalf("RegenerateTools dry-run failed: %v", err)
	}

	newContent, _ := os.ReadFile(pkgPath)
	if string(newContent) != string(originalContent) {
		t.Error("dry-run should not modify package.json")
	}
}

// ----------------------------------------------------------------
// TestRegenerateAgent_GeneratesMarkdown — Task 1.10
// ----------------------------------------------------------------

func TestRegenerateAgent_GeneratesMarkdown(t *testing.T) {
	tmp := t.TempDir()
	ensureDir(t, filepath.Join(tmp, ".opencode"))

	skills := []types.Skill{
		{Name: "foo", Metadata: types.Metadata{Description: "Foo skill"}},
		{Name: "bar", Metadata: types.Metadata{Description: "Bar skill"}},
	}

	err := RegenerateAgent(tmp, skills, false)
	if err != nil {
		t.Fatalf("RegenerateAgent failed: %v", err)
	}

	agentPath := filepath.Join(tmp, ".opencode", "agents", "skill-manager.md")
	content, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("skill-manager.md not created: %v", err)
	}

	s := string(content)
	for _, name := range []string{"foo", "bar"} {
		if !strings.Contains(s, name) {
			t.Errorf("skill %q not found in agent markdown", name)
		}
	}
}

// ----------------------------------------------------------------
// TestRegenerateAgent_DryRun — Task 1.11
// ----------------------------------------------------------------

func TestRegenerateAgent_DryRun(t *testing.T) {
	tmp := t.TempDir()
	ensureDir(t, filepath.Join(tmp, ".opencode"))

	agentPath := filepath.Join(tmp, ".opencode", "agents", "skill-manager.md")

	skills := []types.Skill{
		{Name: "foo", Metadata: types.Metadata{Description: "Foo"}},
	}

	err := RegenerateAgent(tmp, skills, true)
	if err != nil {
		t.Fatalf("RegenerateAgent dry-run failed: %v", err)
	}

	if _, err := os.Stat(agentPath); !os.IsNotExist(err) {
		t.Error("dry-run should not create skill-manager.md")
	}
}

// ----------------------------------------------------------------
// TestCopyAgentsMD — Task 1.12
// ----------------------------------------------------------------

func TestCopyAgentsMD(t *testing.T) {
	tmp := t.TempDir()
	ensureDir(t, filepath.Join(tmp, ".opencode"))

	agentsPath := filepath.Join(tmp, "AGENTS.md")
	agentsContent := "# Agent Skills\n\nTest content"
	if err := os.WriteFile(agentsPath, []byte(agentsContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := CopyAgentsMD(tmp)
	if err != nil {
		t.Fatalf("CopyAgentsMD failed: %v", err)
	}

	opencodePath := filepath.Join(tmp, "OPENCODE.md")
	got, _ := os.ReadFile(opencodePath)
	if string(got) != agentsContent {
		t.Errorf("OPENCODE.md content mismatch:\nwant: %s\ngot:  %s", agentsContent, got)
	}
}

// ----------------------------------------------------------------
// TestSyncOpencode_FullSync — Task 4.1 (verification)
// ----------------------------------------------------------------

func TestSyncOpencode_FullSync(t *testing.T) {
	tmp := t.TempDir()
	
	agentsDir := filepath.Join(tmp, ".agents", "skills")
	opencodeDir := filepath.Join(tmp, ".opencode")
	ensureDir(t, agentsDir)
	ensureDir(t, opencodeDir)

	agentsPath := filepath.Join(tmp, "AGENTS.md")
	agentsContent := "# Agent Skills\n\nTest content"
	os.WriteFile(agentsPath, []byte(agentsContent), 0644)

	fooDir := filepath.Join(agentsDir, "foo")
	ensureDir(t, fooDir)
	fooSrc := filepath.Join(fooDir, "SKILL.md")
	os.WriteFile(fooSrc, []byte("name: foo\ndescription: Foo skill\nauto_invoke: true\n"), 0644)

	err := SyncSkills(tmp, Options{})
	if err != nil {
		t.Fatalf("SyncSkills failed: %v", err)
	}

	skills := []types.Skill{
		{Name: "foo", Metadata: types.Metadata{AutoInvoke: true, Description: "Foo skill"}},
	}
	err = RegenerateTools(tmp, skills, false)
	if err != nil {
		t.Fatalf("RegenerateTools failed: %v", err)
	}
	err = RegenerateAgent(tmp, skills, false)
	if err != nil {
		t.Fatalf("RegenerateAgent failed: %v", err)
	}
	err = CopyAgentsMD(tmp)
	if err != nil {
		t.Fatalf("CopyAgentsMD failed: %v", err)
	}

	fooDest := filepath.Join(tmp, ".opencode", "skills", "foo", "SKILL.md")
	if _, err := os.Stat(fooDest); os.IsNotExist(err) {
		t.Fatal("mirrored skill missing")
	}

	pkgPath := filepath.Join(tmp, ".opencode", "package.json")
	data, _ := os.ReadFile(pkgPath)
	var pkg map[string]interface{}
	json.Unmarshal(data, &pkg)
	oc := pkg["opencode"].(map[string]interface{})
	tools := oc["tools"].([]interface{})
	if len(tools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(tools))
	}

	agentPath := filepath.Join(tmp, ".opencode", "agents", "skill-manager.md")
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		t.Fatal("skill-manager.md missing")
	}

	opencodePath := filepath.Join(tmp, "OPENCODE.md")
	got, _ := os.ReadFile(opencodePath)
	if string(got) != agentsContent {
		t.Fatal("OPENCODE.md not synced")
	}
}

// ----------------------------------------------------------------
// TestSyncOpencode_Idempotent — Task 4.2
// ----------------------------------------------------------------

func TestSyncOpencode_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	
	agentsDir := filepath.Join(tmp, ".agents", "skills")
	opencodeDir := filepath.Join(tmp, ".opencode")
	ensureDir(t, agentsDir)
	ensureDir(t, opencodeDir)

	agentsPath := filepath.Join(tmp, "AGENTS.md")
	os.WriteFile(agentsPath, []byte("# Agent Skills\n"), 0644)

	fooDir := filepath.Join(agentsDir, "foo")
	ensureDir(t, fooDir)
	fooSrc := filepath.Join(fooDir, "SKILL.md")
	os.WriteFile(fooSrc, []byte("name: foo\nauto_invoke: true\n"), 0644)

	skills := []types.Skill{{Name: "foo", Metadata: types.Metadata{AutoInvoke: true}}}

	SyncSkills(tmp, Options{})
	RegenerateTools(tmp, skills, false)
	RegenerateAgent(tmp, skills, false)
	CopyAgentsMD(tmp)

	fooDest := filepath.Join(tmp, ".opencode", "skills", "foo", "SKILL.md")
	stat1, err := os.Stat(fooDest)
	if err != nil {
		t.Fatalf("dest file not created: %v", err)
	}
	mtime1 := stat1.ModTime()

	SyncSkills(tmp, Options{})
	RegenerateTools(tmp, skills, false)
	RegenerateAgent(tmp, skills, false)
	CopyAgentsMD(tmp)

	stat2, err := os.Stat(fooDest)
	if err != nil {
		t.Fatalf("dest file missing after second sync: %v", err)
	}
	mtime2 := stat2.ModTime()

	if !mtime1.Equal(mtime2) {
		t.Error("idempotent sync should produce zero writes on second run")
	}
}

// ----------------------------------------------------------------
// TestSyncOpencode_RemovedSkill — Task 4.3
// ----------------------------------------------------------------

func TestSyncOpencode_RemovedSkill(t *testing.T) {
	tmp := t.TempDir()
	
	agentsDir := filepath.Join(tmp, ".agents", "skills")
	opencodeDir := filepath.Join(tmp, ".opencode")
	ensureDir(t, agentsDir)
	ensureDir(t, opencodeDir)

	fooDir := filepath.Join(agentsDir, "foo")
	ensureDir(t, fooDir)
	fooSrc := filepath.Join(fooDir, "SKILL.md")
	os.WriteFile(fooSrc, []byte("name: foo\nauto_invoke: true\n"), 0644)

	SyncSkills(tmp, Options{})
	skills := []types.Skill{{Name: "foo", Metadata: types.Metadata{AutoInvoke: true}}}
	RegenerateTools(tmp, skills, false)

	os.Remove(fooSrc)

	SyncSkills(tmp, Options{})
	RegenerateTools(tmp, []types.Skill{}, false)

	pkgPath := filepath.Join(tmp, ".opencode", "package.json")
	data, _ := os.ReadFile(pkgPath)
	var pkg map[string]interface{}
	json.Unmarshal(data, &pkg)
	oc := pkg["opencode"].(map[string]interface{})
	ocTools, ok := oc["tools"]
	if !ok {
		t.Fatal("tools key missing from opencode")
	}
	tools, ok := ocTools.([]interface{})
	if !ok {
		t.Fatal("tools is not a slice")
	}

	if len(tools) != 0 {
		t.Errorf("expected 0 tools after skill removal, got %d", len(tools))
	}
}

// ----------------------------------------------------------------
// Helper: initPackageJSON
// ----------------------------------------------------------------

func initPackageJSON(t *testing.T, tmp string) {
	t.Helper()
	pkgPath := filepath.Join(tmp, ".opencode", "package.json")
	pkg := map[string]interface{}{
		"name": "test-project",
		"opencode": map[string]interface{}{
			"tools": []interface{}{},
		},
	}
	data, err := json.Marshal(pkg)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(pkgPath, data, 0644)
}
