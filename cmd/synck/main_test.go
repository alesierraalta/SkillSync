package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skillsync/tui/internal/bundle"
	"skillsync/tui/internal/diff"
	"skillsync/tui/internal/install"
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/vault"
)

func TestCLIFlags(t *testing.T) {
	// 1. Test -g flag triggers GlobalInstaller
	calledGlobal := false
	GlobalInstaller = func() install.Result {
		calledGlobal = true
		return install.Result{Success: true, Message: "Mocked install"}
	}
	// Mock TUI runner to avoid launching it
	TUIRunner = func() error {
		return nil
	}

	err := run([]string{"synck", "-g"})
	if err != nil {
		t.Fatalf("run(-g) failed: %v", err)
	}
	if !calledGlobal {
		t.Error("expected GlobalInstaller to be called when -g is present")
	}

	// 2. Test 'install -g' triggers GlobalInstaller
	calledGlobal = false
	err = run([]string{"synck", "install", "-g"})
	if err != nil {
		t.Fatalf("run(install -g) failed: %v", err)
	}
	if !calledGlobal {
		t.Error("expected GlobalInstaller to be called when 'install -g' is present")
	}

	// 3. Test no flags triggers TUIRunner
	calledTUI := false
	TUIRunner = func() error {
		calledTUI = true
		return nil
	}
	calledGlobal = false

	err = run([]string{"synck"})
	if err != nil {
		t.Fatalf("run() failed: %v", err)
	}
	if !calledTUI {
		t.Error("expected TUIRunner to be called when no flags are present")
	}
	if calledGlobal {
		t.Error("expected GlobalInstaller NOT to be called when no flags are present")
	}
}

func TestNewCLICommands(t *testing.T) {
	calledTUI := false
	TUIRunner = func() error {
		calledTUI = true
		return nil
	}
	calledGlobal := false
	GlobalInstaller = func() install.Result {
		calledGlobal = true
		return install.Result{Success: true}
	}

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		checkFn  func() bool
		checkMsg string
	}{
		{
			name:    "skill command succeeds",
			args:    []string{"synck", "skill"},
			wantErr: false,
			checkFn: func() bool { return !calledTUI && !calledGlobal },
			checkMsg: "skill should not trigger TUIRunner or GlobalInstaller",
		},
		{
			name:    "find command succeeds",
			args:    []string{"synck", "find"},
			wantErr: false,
			checkFn: func() bool { return !calledTUI && !calledGlobal },
			checkMsg: "find should not trigger TUIRunner or GlobalInstaller",
		},
		{
			name:    "fullskills command succeeds",
			args:    []string{"synck", "fullskills"},
			wantErr: false,
			checkFn: func() bool { return !calledTUI && !calledGlobal },
			checkMsg: "fullskills should not trigger TUIRunner or GlobalInstaller",
		},
		{
			name:    "create alias enforces strict validation",
			args:    []string{"synck", "create", "test prompt"},
			wantErr: true,
			checkFn: func() bool { return !calledTUI && !calledGlobal },
			checkMsg: "create should not trigger TUIRunner or GlobalInstaller",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calledTUI = false
			calledGlobal = false
			err := run(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("run(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
			if !tt.checkFn() {
				t.Error(tt.checkMsg)
			}
		})
	}
}

func TestCreateSkill_GeneratesPack(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	_ = os.MkdirAll(".agents", 0755)
	_ = os.WriteFile("AGENTS.md", []byte("# Agent Skills\n"), 0644)

	err := run([]string{"synck", "create", "Go test quality guardrails"})
	if err != nil {
		t.Fatalf("run(create) failed: %v", err)
	}

	skillDir := filepath.Join(".agents", "skills", "go-test-quality-guardrails")
	for _, p := range []string{
		filepath.Join(skillDir, "SKILL.md"),
		filepath.Join(skillDir, "assets", "SKILL-TEMPLATE.md"),
		filepath.Join(skillDir, "references", "README.md"),
		filepath.Join(skillDir, "CHECKLIST.md"),
	} {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Fatalf("expected file to exist: %s", p)
		}
	}
}

func TestCreateSkill_RejectsWeakPrompt(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	_ = os.MkdirAll(".agents", 0755)
	err := run([]string{"synck", "create", "auth"})
	if err == nil {
		t.Fatal("expected strict validation error for weak prompt")
	}
	if !strings.Contains(err.Error(), "strict") {
		t.Fatalf("expected strict validation message, got: %v", err)
	}
}

func TestFindProjectRoot_MultiProvider(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(tmpDir string) (startDir string, wantRoot string)
	}{
		{
			name: "detects .opencode as root",
			setup: func(tmpDir string) (string, string) {
				subdir := filepath.Join(tmpDir, "src", "utils")
				_ = os.MkdirAll(subdir, 0755)
				opencodeDir := filepath.Join(tmpDir, ".opencode")
				_ = os.MkdirAll(opencodeDir, 0755)
				return subdir, tmpDir
			},
		},
		{
			name: "detects .qwen as root",
			setup: func(tmpDir string) (string, string) {
				subdir := filepath.Join(tmpDir, "src")
				_ = os.MkdirAll(subdir, 0755)
				qwenDir := filepath.Join(tmpDir, ".qwen")
				_ = os.MkdirAll(qwenDir, 0755)
				return subdir, tmpDir
			},
		},
		{
			name: "detects .claude as root",
			setup: func(tmpDir string) (string, string) {
				subdir := filepath.Join(tmpDir, "lib")
				_ = os.MkdirAll(subdir, 0755)
				claudeDir := filepath.Join(tmpDir, ".claude")
				_ = os.MkdirAll(claudeDir, 0755)
				return subdir, tmpDir
			},
		},
		{
			name: "detects AGENTS.md as root",
			setup: func(tmpDir string) (string, string) {
				subdir := filepath.Join(tmpDir, "docs")
				_ = os.MkdirAll(subdir, 0755)
				_ = os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte("# Agents"), 0644)
				return subdir, tmpDir
			},
		},
		{
			name: "detects OPENCODE.md as root",
			setup: func(tmpDir string) (string, string) {
				subdir := filepath.Join(tmpDir, "config")
				_ = os.MkdirAll(subdir, 0755)
				_ = os.WriteFile(filepath.Join(tmpDir, "OPENCODE.md"), []byte("# OpenCode"), 0644)
				return subdir, tmpDir
			},
		},
		{
			name: "finds closest root",
			setup: func(tmpDir string) (string, string) {
				// .agents at tmpDir, .opencode in nested/deep — closest is nested/deep
				subdir := filepath.Join(tmpDir, "nested", "deep")
				_ = os.MkdirAll(subdir, 0755)
				_ = os.MkdirAll(filepath.Join(tmpDir, ".agents"), 0755)
				_ = os.MkdirAll(filepath.Join(subdir, ".opencode"), 0755)
				return subdir, subdir // closest root is in subdir itself
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			startDir, wantRoot := tt.setup(tmpDir)
			got := findProjectRoot(startDir)
			if got != wantRoot {
				t.Errorf("findProjectRoot(%q) = %q, want %q", startDir, got, wantRoot)
			}
		})
	}
}

func TestVersionCommand(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// Mock TUI runner to avoid launching TUI
	TUIRunner = func() error { return nil }
	GlobalInstaller = func() install.Result { return install.Result{Success: true} }

	err := run([]string{"synck", "version"})
	if err != nil {
		t.Fatalf("run(synck version) failed: %v", err)
	}

	// Restore stdout and read
	w.Close()
	var buf [1024]byte
	n, _ := r.Read(buf[:])
	output := string(buf[:n])

	// Version output must contain "version" keyword and version number
	if !strings.Contains(output, "version") {
		t.Errorf("version output must contain 'version', got: %s", output)
	}
	if !strings.Contains(output, "0.") && !strings.Contains(output, "1.") {
		t.Errorf("version output must contain version number, got: %s", output)
	}
}

func TestSetupOpenCodeCommand_CreatesFiles(t *testing.T) {
	// Test that `synck setup-opencode` creates .opencode/package.json with tools
	// and .opencode/agents/skill-manager.md
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Ensure .opencode dir exists
	_ = os.MkdirAll(".opencode", 0755)

	TUIRunner = func() error { return nil }
	GlobalInstaller = func() install.Result { return install.Result{Success: true} }

	err := run([]string{"synck", "setup-opencode"})
	if err != nil {
		t.Fatalf("run(synck setup-opencode) failed: %v", err)
	}

	// Verify .opencode/commands has base command markdown files
	for _, tool := range []string{"skill", "find", "create", "sync", "fullskills"} {
		cmdPath := filepath.Join(".opencode", "commands", tool+".md")
		if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
			t.Errorf("missing command file %q", cmdPath)
		}
	}

	// Verify .opencode/package.json exists and has empty tools
	pkgPath := ".opencode/package.json"
	content, _ := os.ReadFile(pkgPath)
	if strings.Contains(string(content), `"name": "skill"`) {
		t.Errorf("tool 'skill' should not be in package.json tools array")
	}

	// Verify .opencode/agents/skill-manager.md exists
	agentPath := ".opencode/agents/skill-manager.md"
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		t.Errorf("expected skill-manager.md to exist after setup-opencode")
	}
}

// ----------------------------------------------------------------
// TestMain_SyncOpencodeFlags — Task 1.13
// ----------------------------------------------------------------

func TestMain_SyncOpencodeFlags(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	TUIRunner = func() error { return nil }
	GlobalInstaller = func() install.Result { return install.Result{Success: true} }

	// Create a minimal .agents directory and AGENTS.md so findProjectRoot and copyAgentsMD work
	_ = os.MkdirAll(".agents", 0755)
	_ = os.WriteFile("AGENTS.md", []byte("# Agent Skills\n"), 0644)

	// Create a skill in .agents/skills so there's something to mirror
	_ = os.MkdirAll(".agents/skills/foo", 0755)
	_ = os.WriteFile(".agents/skills/foo/SKILL.md", []byte("name: foo\n"), 0644)

	// Test --prune flag
	err := run([]string{"synck", "sync-opencode", "--prune"})
	if err != nil {
		t.Fatalf("run(synck sync-opencode --prune) failed: %v", err)
	}

	// Verify .opencode/skills directory was created (skill was mirrored)
	skillsPath := ".opencode/skills"
	if _, err := os.Stat(skillsPath); os.IsNotExist(err) {
		t.Errorf("expected .opencode/skills to exist after sync-opencode --prune")
	}

	// Verify the skill was mirrored
	if _, err := os.Stat(".opencode/skills/foo/SKILL.md"); os.IsNotExist(err) {
		t.Errorf("expected .opencode/skills/foo/SKILL.md to exist after mirroring")
	}

	// Test --dry-run flag - remove .opencode first
	_ = os.RemoveAll(".opencode")

	err = run([]string{"synck", "sync-opencode", "--dry-run"})
	if err != nil {
		t.Fatalf("run(synck sync-opencode --dry-run) failed: %v", err)
	}

	// Verify .opencode was NOT created (dry-run)
	if _, err := os.Stat(".opencode"); !os.IsNotExist(err) {
		t.Errorf("dry-run should not create .opencode directory")
	}

	// Test combined flags - recreate .opencode
	_ = os.MkdirAll(".opencode", 0755)
	err = run([]string{"synck", "sync-opencode", "--prune", "--dry-run"})
	if err != nil {
		t.Fatalf("run(synck sync-opencode --prune --dry-run) failed: %v", err)
	}
}

// ----------------------------------------------------------------
// TestMain_AutoChain — Task 1.14
// ----------------------------------------------------------------

func TestMain_AutoChain(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	TUIRunner = func() error { return nil }
	GlobalInstaller = func() install.Result { return install.Result{Success: true} }

	// Create .agents directory so findProjectRoot works
	_ = os.MkdirAll(".agents", 0755)
	// Create AGENTS.md so copyAgentsMD works
	_ = os.WriteFile("AGENTS.md", []byte("# Agent Skills\n\n## Available Skills\n\n| Skill | Description | Location |\n| ----- | ----------- | -------- |\n\n### Auto-invoke Skills\n\nWhen performing these actions, ALWAYS invoke the corresponding skill FIRST:\n\n| Action                              | Skill      |\n| ----------------------------------- | ---------- |\n"), 0644)

	// Create the legacy sync script path for compatibility coverage.
	// This is where runner.LegacySyncScriptPath points to.
	_ = os.MkdirAll(".agents/skills/skill-sync/assets", 0755)
	syncScript := `#!/bin/bash
echo "sync done"
exit 0
`
	_ = os.WriteFile(".agents/skills/skill-sync/assets/sync.sh", []byte(syncScript), 0755)

	// Run sync (which should auto-chain sync-opencode)
	err := run([]string{"synck", "sync"})
	if err != nil {
		t.Fatalf("run(synck sync) failed: %v", err)
	}

	// Verify sync-opencode was chained (OPENCODE.md should exist)
	if _, err := os.Stat("OPENCODE.md"); os.IsNotExist(err) {
		t.Errorf("expected OPENCODE.md to exist after sync auto-chained sync-opencode")
	}
}

// ----------------------------------------------------------------
// renderReport tests — PR3 TDD
// ----------------------------------------------------------------

func TestRenderReport_NoChanges(t *testing.T) {
	report := &runner.SyncReport{}
	got := renderReport(report, false)
	want := "No files changed.\n"
	if got != want {
		t.Errorf("renderReport(no changes) = %q, want %q", got, want)
	}
}

func TestRenderReport_Modified(t *testing.T) {
	before := "line1\nline2\n"
	after := "line1\nline2modified\n"
	diffStr, summary := diff.UnifiedDiff(before, after, 50)
	report := &runner.SyncReport{
		Changes: []runner.FileChange{
			{Path: "AGENTS.md", Status: "modified", Before: before, After: after, Diff: diffStr, Summary: summary},
		},
	}
	got := renderReport(report, false)
	if !strings.Contains(got, "AGENTS.md (modified)") {
		t.Errorf("expected status line in output, got:\n%s", got)
	}
	if !strings.Contains(got, summary) {
		t.Errorf("expected summary %q in output, got:\n%s", summary, got)
	}
	if !strings.Contains(got, "@@") {
		t.Errorf("expected diff hunk in output, got:\n%s", got)
	}
}

func TestRenderReport_CreatedAndDeleted(t *testing.T) {
	newDiff, newSummary := diff.UnifiedDiff("", "new content\n", 50)
	delDiff, delSummary := diff.UnifiedDiff("old content\n", "", 50)
	report := &runner.SyncReport{
		Changes: []runner.FileChange{
			{Path: "new.md", Status: "created", Before: "", After: "new content\n", Diff: newDiff, Summary: newSummary},
			{Path: "old.md", Status: "deleted", Before: "old content\n", After: "", Diff: delDiff, Summary: delSummary},
			{Path: "link.md", Status: "symlinked", Before: "", After: "", Diff: "", Summary: ""},
		},
	}
	got := renderReport(report, false)
	if !strings.Contains(got, "new.md (created)") {
		t.Errorf("expected new.md created, got:\n%s", got)
	}
	if !strings.Contains(got, "old.md (deleted)") {
		t.Errorf("expected old.md deleted, got:\n%s", got)
	}
	if !strings.Contains(got, "link.md (symlinked)") {
		t.Errorf("expected link.md symlinked, got:\n%s", got)
	}
}

func TestRenderReport_VerboseFullDiff(t *testing.T) {
	var beforeLines, afterLines []string
	for i := 0; i < 30; i++ {
		beforeLines = append(beforeLines, fmt.Sprintf("line %d", i))
		afterLines = append(afterLines, fmt.Sprintf("line %d", i))
	}
	for i := 30; i < 80; i++ {
		afterLines = append(afterLines, fmt.Sprintf("new line %d", i))
	}
	before := strings.Join(beforeLines, "\n")
	after := strings.Join(afterLines, "\n")

	cappedDiff, summary := diff.UnifiedDiff(before, after, 50)
	report := &runner.SyncReport{
		Changes: []runner.FileChange{
			{Path: "big.md", Status: "modified", Before: before, After: after, Diff: cappedDiff, Summary: summary},
		},
	}

	capped := renderReport(report, false)
	verbose := renderReport(report, true)

	// Verbose should contain more diff lines than capped
	cappedLines := strings.Count(capped, "\n")
	verboseLines := strings.Count(verbose, "\n")
	if verboseLines <= cappedLines {
		t.Errorf("expected verbose (%d lines) > capped (%d lines)", verboseLines, cappedLines)
	}
}

// ----------------------------------------------------------------
// handleSync integration tests — PR3
// ----------------------------------------------------------------

func TestHandleSync_ProgressAndReport(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Set up minimal project
	_ = os.MkdirAll(".agents/skills/foo", 0755)
	_ = os.WriteFile(".agents/skills/foo/SKILL.md", []byte("name: foo\ndescription: Foo skill\nauto_invoke: true\n"), 0644)
	_ = os.WriteFile("AGENTS.md", []byte("# Agent Skills\n\n## Available Skills\n\n| Skill | Description | Location |\n| ----- | ----------- | -------- |\n\n### Auto-invoke Skills\n\nWhen performing these actions, ALWAYS invoke the corresponding skill FIRST:\n\n| Action                              | Skill      |\n| ----------------------------------- | ---------- |\n"), 0644)

	// Capture stdout concurrently to avoid pipe buffer deadlock
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var buf strings.Builder
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	err := handleSync(nil)

	w.Close()
	<-done
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("handleSync failed: %v", err)
	}

	output := buf.String()

	// Must contain progress lines for all 8 stages
	for i := 1; i <= 8; i++ {
		if !strings.Contains(output, fmt.Sprintf("→ [%d/8]", i)) {
			t.Errorf("missing progress line for stage %d in output:\n%s", i, output)
		}
	}

	// Must contain completion message
	if !strings.Contains(output, "Synchronization complete.") {
		t.Errorf("missing completion message in output:\n%s", output)
	}

	// Must contain report header
	if !strings.Contains(output, "Changed files:") {
		t.Errorf("missing report header in output:\n%s", output)
	}
}

func TestHandleSync_NoChanges(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Set up minimal project with no skills
	_ = os.MkdirAll(".agents", 0755)
	_ = os.WriteFile("AGENTS.md", []byte("# Agent Skills\n\n## Available Skills\n\n| Skill | Description | Location |\n| ----- | ----------- | -------- |\n\n### Auto-invoke Skills\n\nWhen performing these actions, ALWAYS invoke the corresponding skill FIRST:\n\n| Action                              | Skill      |\n| ----------------------------------- | ---------- |\n"), 0644)

	// First run to set everything up
	_ = handleSync(nil)

	// Second run should be idempotent and report no changes
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var buf strings.Builder
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	err := handleSync(nil)

	w.Close()
	<-done
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("handleSync failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "No files changed.") {
		t.Errorf("expected 'No files changed.' in output:\n%s", output)
	}
}

func TestHandleSync_VerboseFlag(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Set up project with a skill that produces a large diff
	_ = os.MkdirAll(".agents/skills/foo", 0755)
	_ = os.WriteFile(".agents/skills/foo/SKILL.md", []byte("name: foo\ndescription: Foo skill\nauto_invoke: true\n"), 0644)

	// Create AGENTS.md with required sections so it gets modified
	_ = os.WriteFile("AGENTS.md", []byte("# Agent Skills\n\n## Available Skills\n\n| Skill | Description | Location |\n| ----- | ----------- | -------- |\n\n### Auto-invoke Skills\n\nWhen performing these actions, ALWAYS invoke the corresponding skill FIRST:\n\n| Action                              | Skill      |\n| ----------------------------------- | ---------- |\n"), 0644)

	// Capture stdout (non-verbose) concurrently
	oldStdout := os.Stdout
	r1, w1, _ := os.Pipe()
	os.Stdout = w1
	var buf1 strings.Builder
	done1 := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf1, r1)
		close(done1)
	}()
	_ = handleSync(nil)
	w1.Close()
	<-done1
	os.Stdout = oldStdout
	cappedOutput := buf1.String()

	// Capture stdout (verbose) concurrently
	r2, w2, _ := os.Pipe()
	os.Stdout = w2
	var buf2 strings.Builder
	done2 := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf2, r2)
		close(done2)
	}()
	_ = handleSync([]string{"--verbose"})
	w2.Close()
	<-done2
	os.Stdout = oldStdout
	verboseOutput := buf2.String()

	// Both should have progress and completion
	if !strings.Contains(cappedOutput, "→ [1/8]") {
		t.Error("capped output missing progress line")
	}
	if !strings.Contains(verboseOutput, "→ [1/8]") {
		t.Error("verbose output missing progress line")
	}
}

// ----------------------------------------------------------------
// Bundle CLI tests — PR1 Phase 4
// ----------------------------------------------------------------

func TestCLIBundleExport_Dispatch(t *testing.T) {
	called := false
	var capturedNames []string
	var capturedOutput string

	oldExport := BundleExport
	BundleExport = func(svc *vault.Service, names []string, outputPath string) (string, error) {
		called = true
		capturedNames = names
		capturedOutput = outputPath
		return outputPath, nil
	}
	defer func() { BundleExport = oldExport }()

	TUIRunner = func() error { return nil }
	GlobalInstaller = func() install.Result { return install.Result{Success: true} }

	tmpDir := t.TempDir()
	err := run([]string{"synck", "bundle", "export", "--output", filepath.Join(tmpDir, "out.skillsync"), "skill-a", "skill-b"})
	if err != nil {
		t.Fatalf("bundle export failed: %v", err)
	}

	if !called {
		t.Error("BundleExport was not called")
	}
	if len(capturedNames) != 2 || capturedNames[0] != "skill-a" || capturedNames[1] != "skill-b" {
		t.Errorf("export names = %v, want [skill-a skill-b]", capturedNames)
	}
	if !strings.HasSuffix(capturedOutput, "out.skillsync") {
		t.Errorf("output path = %s, should end with out.skillsync", capturedOutput)
	}
}

func TestCLIBundleList_Dispatch(t *testing.T) {
	// Create a real bundle to list
	tmpDir := t.TempDir()
	bundlePath := filepath.Join(tmpDir, "listtest.skillsync")

	svc := vault.NewServiceWithRoot(tmpDir)
	skillDir := filepath.Join(tmpDir, "listskill")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("name: listskill\n"), 0644)
	os.WriteFile(filepath.Join(skillDir, "METADATA.json"), []byte(`{"skill_name":"listskill"}`), 0644)

	_, err := bundle.Export(svc, []string{"listskill"}, bundlePath)
	if err != nil {
		t.Fatal(err)
	}

	TUIRunner = func() error { return nil }
	GlobalInstaller = func() install.Result { return install.Result{Success: true} }

	err = run([]string{"synck", "bundle", "list", bundlePath})
	if err != nil {
		t.Fatalf("bundle list failed: %v", err)
	}
}

func TestCLIInstallBundle_Dispatch(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create .agents so findProjectRoot works
	os.MkdirAll(".agents", 0755)

	bundlePath := filepath.Join(tmpDir, "test-bundle.skillsync")
	// Create a valid bundle file
	os.WriteFile(bundlePath, []byte("fake bundle"), 0644)

	called := false
	var capturedBundlePath string

	oldImport := BundleImport
	BundleImport = func(bp string, opts bundle.ImportOptions) ([]bundle.ImportResult, error) {
		called = true
		capturedBundlePath = bp
		return []bundle.ImportResult{{Skill: "test", Status: "installed"}}, nil
	}
	defer func() { BundleImport = oldImport }()

	err := run([]string{"synck", "install", "--bundle", bundlePath})
	if err != nil {
		t.Fatalf("install --bundle failed: %v", err)
	}

	if !called {
		t.Error("BundleImport was not called via install --bundle")
	}
	if capturedBundlePath != bundlePath {
		t.Errorf("bundle path = %s, want %s", capturedBundlePath, bundlePath)
	}
}

func TestCLIBundleImport_Dispatch(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create .agents/skills so findProjectRoot works
	os.MkdirAll(".agents/skills", 0755)

	bundlePath := filepath.Join(tmpDir, "import-test.skillsync")
	// Write a dummy bundle file (not actually opened, just needs to exist for Stat check)
	os.WriteFile(bundlePath, []byte("dummy"), 0644)

	called := false
	oldImport := BundleImport
	BundleImport = func(bp string, opts bundle.ImportOptions) ([]bundle.ImportResult, error) {
		called = true
		return []bundle.ImportResult{{Skill: "mocked", Status: "installed"}}, nil
	}
	defer func() { BundleImport = oldImport }()

	TUIRunner = func() error { return nil }
	GlobalInstaller = func() install.Result { return install.Result{Success: true} }

	err := run([]string{"synck", "bundle", "import", bundlePath})
	if err != nil {
		t.Fatalf("bundle import failed: %v", err)
	}

	if !called {
		t.Error("BundleImport was not called via bundle import")
	}
}

func TestCLIBundleExport_MissingSkill_ReturnsError(t *testing.T) {
	oldExport := BundleExport
	BundleExport = func(svc *vault.Service, names []string, outputPath string) (string, error) {
		return "", fmt.Errorf("skill %q not found", names[1])
	}
	defer func() { BundleExport = oldExport }()

	TUIRunner = func() error { return nil }
	GlobalInstaller = func() install.Result { return install.Result{Success: true} }

	err := run([]string{"synck", "bundle", "export", "existing", "missing"})
	if err == nil {
		t.Fatal("expected error for missing skill in bundle export")
	}
}

func TestCLIBundleImport_ParseFailure_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.MkdirAll(".agents/skills", 0755)
	bundlePath := filepath.Join(tmpDir, "bad.skillsync")
	os.WriteFile(bundlePath, []byte("not a zip"), 0644)

	TUIRunner = func() error { return nil }
	GlobalInstaller = func() install.Result { return install.Result{Success: true} }

	err := run([]string{"synck", "bundle", "import", bundlePath})
	if err == nil {
		t.Fatal("expected error for invalid bundle import")
	}
}

func TestCLIInstallBundle_FailurePropagates(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.MkdirAll(".agents", 0755)
	bundlePath := filepath.Join(tmpDir, "fail.skillsync")
	os.WriteFile(bundlePath, []byte("dummy"), 0644)

	oldImport := BundleImport
	BundleImport = func(bp string, opts bundle.ImportOptions) ([]bundle.ImportResult, error) {
		return []bundle.ImportResult{
			{Skill: "fail", Status: "failed", Error: fmt.Errorf("write error")},
		}, nil
	}
	defer func() { BundleImport = oldImport }()

	TUIRunner = func() error { return nil }
	GlobalInstaller = func() install.Result { return install.Result{Success: true} }

	err := run([]string{"synck", "install", "--bundle", bundlePath})
	if err == nil {
		t.Fatal("expected error for failed import results")
	}
	if !strings.Contains(err.Error(), "failed") {
		t.Errorf("error = %q, want 'failed'", err.Error())
	}
}
