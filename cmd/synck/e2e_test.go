package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestHandleSync_HappyPath exercises the happy path of the synck CLI:
// setting up a minimal project with .agents skills and running a dry-run sync.
// This tests AUDIT-06a/b: e2e coverage of the main flow without terminal requirement.
func TestHandleSync_HappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Set up minimal project structure
	if err := os.MkdirAll(".agents/skills/test-skill", 0755); err != nil {
		t.Fatal(err)
	}

	// Create a minimal SKILL.md
	skillContent := `---
name: test-skill
description: Test skill for e2e
---
Test skill content.
`
	if err := os.WriteFile(".agents/skills/test-skill/SKILL.md", []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create initial AGENTS.md
	agentsContent := `# Agent Skills

## Available Skills

| Skill | Description | Location |
| ----- | ----------- | -------- |

### Auto-invoke Skills

| Action | Skill |
| --- | --- |
`
	if err := os.WriteFile("AGENTS.md", []byte(agentsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Run dry-run sync
	err = handleSync([]string{"--dry-run"})
	if err != nil {
		t.Fatalf("handleSync dry-run failed: %v", err)
	}

	// Verify AGENTS.md still has original content (dry-run, no write)
	raw, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("ReadFile AGENTS.md: %v", err)
	}
	if string(raw) != agentsContent {
		t.Error("dry-run should not modify AGENTS.md")
	}

	// Verify the .agents fixture directory still exists and is readable
	if _, err := os.Stat(filepath.Join(".agents", "skills", "test-skill", "SKILL.md")); err != nil {
		t.Errorf("expected .agents fixture to remain intact: %v", err)
	}
}
