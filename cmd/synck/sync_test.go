package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"skillsync/tui/internal/install"
	"skillsync/tui/internal/opencode"
	"skillsync/tui/internal/runner"
	"strings"
	"testing"
)

func TestHandleSync_AutoskillsFlag(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Set up minimal project
	_ = os.MkdirAll(".agents", 0755)
	_ = os.WriteFile("AGENTS.md", []byte("# Agent Skills\n\n## Available Skills\n\n| Skill | Description | Location |\n| ----- | ----------- | -------- |\n\n### Auto-invoke Skills\n\nWhen performing these actions, ALWAYS invoke the corresponding skill FIRST:\n\n| Action | Skill |\n| --- | --- |\n"), 0644)

	// Mock AutoskillsInstaller
	called := false
	AutoskillsInstaller = func(ctx context.Context) install.AutoskillsResult {
		called = true
		return install.AutoskillsResult{Success: true, Output: "mocked output"}
	}

	// 1. Run sync with --autoskills
	err := handleSync([]string{"--autoskills"})

	// Should fail now because flag is not defined in handleSync
	if err != nil {
		t.Fatalf("handleSync(--autoskills) failed: %v", err)
	}

	if !called {
		t.Error("expected AutoskillsInstaller to be called when --autoskills is present")
	}
}

func TestSyncOpenCodeForRoot_DryRunReportsWithoutWriting(t *testing.T) {
	tmpDir := t.TempDir()

	writeFile := func(path, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	agentsContent := "# Agents\n\n## Available Skills\n\n| Skill | Description | Location |\n| --- | --- | --- |\n\n### Auto-invoke Skills\n\n| Action | Skill |\n| --- | --- |\n"
	writeFile(filepath.Join(tmpDir, "AGENTS.md"), agentsContent)
	writeFile(filepath.Join(tmpDir, ".agents", "skills", "foo", "SKILL.md"), "---\nname: foo\ndescription: Foo skill\nauto_invoke:\n  - Use foo\n---\n")
	writeFile(filepath.Join(tmpDir, ".opencode", "skills", "foo", "SKILL.md"), "old mirrored skill\n")
	writeFile(filepath.Join(tmpDir, ".opencode", "skills", "orphan", "SKILL.md"), "orphan skill\n")
	writeFile(filepath.Join(tmpDir, ".opencode", "package.json"), "{\"opencode\":{\"tools\":[{\"name\":\"old\"}]}}")
	writeFile(filepath.Join(tmpDir, ".opencode", "agents", "skill-manager.md"), "old agent\n")
	writeFile(filepath.Join(tmpDir, ".opencode", "commands", "foo.md"), "---\nmanaged_by: skillsync\ndescription: Old foo\n---\n\nold\n")
	writeFile(filepath.Join(tmpDir, ".opencode", "commands", "orphan.md"), "---\nmanaged_by: skillsync\ndescription: Orphan\n---\n\norphan\n")
	writeFile(filepath.Join(tmpDir, "OPENCODE.md"), "old opencode\n")

	before := snapshotFiles(t, tmpDir, []string{
		"AGENTS.md",
		"OPENCODE.md",
		".opencode/package.json",
		".opencode/commands/foo.md",
		".opencode/commands/orphan.md",
		".opencode/agents/skill-manager.md",
		".opencode/skills/foo/SKILL.md",
		".opencode/skills/orphan/SKILL.md",
	})

	report, err := syncOpenCodeForRoot(tmpDir, opencode.Options{DryRun: true, Prune: true})
	if err != nil {
		t.Fatalf("syncOpenCodeForRoot dry-run failed: %v", err)
	}

	after := snapshotFiles(t, tmpDir, []string{
		"AGENTS.md",
		"OPENCODE.md",
		".opencode/package.json",
		".opencode/commands/foo.md",
		".opencode/commands/orphan.md",
		".opencode/agents/skill-manager.md",
		".opencode/skills/foo/SKILL.md",
		".opencode/skills/orphan/SKILL.md",
	})
	for path, beforeContent := range before {
		if after[path] != beforeContent {
			t.Fatalf("dry-run wrote %s", path)
		}
	}
	if _, err := os.Stat(filepath.Join(tmpDir, ".opencode", "commands", "skill.md")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create command files, stat error: %v", err)
	}

	assertSyncReportHasChange(t, report, ".opencode/skills/foo/SKILL.md", "modified")
	assertSyncReportHasChange(t, report, ".opencode/skills/orphan", "deleted")
	assertSyncReportHasChange(t, report, ".opencode/package.json", "modified")
	assertSyncReportHasChange(t, report, ".opencode/agents/skill-manager.md", "modified")
	assertSyncReportHasChange(t, report, "OPENCODE.md", "modified")
	assertSyncReportHasChange(t, report, ".opencode/commands/foo.md", "modified")
	assertSyncReportHasChange(t, report, ".opencode/commands/orphan.md", "deleted")
	assertSyncReportHasChange(t, report, ".opencode/commands/skill.md", "created")
}

func TestHandleSyncDryRunReportsOPENCODEFromPlannedAgents(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	writeFile := func(path, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	staleAgents := "# Agent Skills\n\n## Available Skills\n\n| Skill | Description | Location |\n| ----- | ----------- | -------- |\n\n### Auto-invoke Skills\n\nWhen performing these actions, ALWAYS invoke the corresponding skill FIRST:\n\n| Action | Skill |\n| --- | --- |\n"
	writeFile("AGENTS.md", staleAgents)
	writeFile("OPENCODE.md", staleAgents)
	writeFile(filepath.Join(".agents", "skills", "foo", "SKILL.md"), "---\nname: foo\ndescription: Foo skill\nauto_invoke:\n  - Use foo\n---\n")

	var syncErr error
	output := captureStdout(t, func() {
		syncErr = handleSync([]string{"--dry-run", "--verbose"})
	})
	if syncErr != nil {
		t.Fatalf("handleSync dry-run failed: %v", syncErr)
	}

	if got := readFileString(t, "AGENTS.md"); got != staleAgents {
		t.Fatalf("dry-run wrote AGENTS.md")
	}
	if got := readFileString(t, "OPENCODE.md"); got != staleAgents {
		t.Fatalf("dry-run wrote OPENCODE.md")
	}
	if !strings.Contains(output, "OPENCODE.md (modified)") {
		t.Fatalf("expected OPENCODE.md modified in report, got:\n%s", output)
	}
	if !strings.Contains(output, "| `foo` | Foo skill | [SKILL.md](.agents/skills/foo/SKILL.md) |") {
		t.Fatalf("expected OPENCODE.md dry-run report to include planned AGENTS.md content, got:\n%s", output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	var buf bytes.Buffer
	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(&buf, r)
		done <- err
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = oldStdout
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func readFileString(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func snapshotFiles(t *testing.T, root string, paths []string) map[string]string {
	t.Helper()
	snapshot := make(map[string]string, len(paths))
	for _, path := range paths {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		snapshot[path] = string(content)
	}
	return snapshot
}

func assertSyncReportHasChange(t *testing.T, report *runner.SyncReport, path, status string) {
	t.Helper()
	for _, change := range report.Changes {
		if change.Path == path && change.Status == status {
			return
		}
	}
	t.Fatalf("expected report change %s %s, got %#v", path, status, report.Changes)
}
