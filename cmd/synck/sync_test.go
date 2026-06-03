package main

import (
	"context"
	"os"
	"testing"
	"skillsync/tui/internal/install"
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
