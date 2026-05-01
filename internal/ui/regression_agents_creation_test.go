package ui

import (
	"os"
	"path/filepath"
	"skillsync/tui/internal/runner"
	"testing"
)

func TestSyncCreatesMissingAgentsMD(t *testing.T) {
	// 1. Setup a temp directory representing an external project
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to tmpDir: %v", err)
	}

	// 2. Setup skill-sync assets but NO AGENTS.md
	syncScriptPath := filepath.Join(".agents", "skills", "skill-sync", "assets", "sync.sh")
	if err := os.MkdirAll(filepath.Dir(syncScriptPath), 0755); err != nil {
		t.Fatalf("failed to create sync script dir: %v", err)
	}
	
	// Create a dummy sync.sh that checks for AGENTS.md
	dummySyncContent := `#!/bin/bash
if [ ! -f "../../../../AGENTS.md" ]; then
   echo "Error: AGENTS.md missing"
  exit 1
fi
echo "Sync Successful"
exit 0
`
	if err := os.WriteFile(syncScriptPath, []byte(dummySyncContent), 0755); err != nil {
		t.Fatalf("failed to write dummy sync.sh: %v", err)
	}

	// 3. Initialize Model and start sync
	m := NewModel()
	m.Screen = ScreenSyncing
	
	// This should now call ensureAgentsMD() before running the script
	cmd := m.startSync()
	msg := cmd()

	res, ok := msg.(runner.SyncResult)
	if !ok {
		t.Fatalf("expected runner.SyncResult, got %T", msg)
	}

	// 4. Assertions
	if res.ExitCode != 0 {
		t.Errorf("Sync failed with exit code %d, stderr: %q. Expected AGENTS.md creation and success.", res.ExitCode, res.Stderr)
	}

	// Verify AGENTS.md was created
	if _, err := os.Stat("AGENTS.md"); os.IsNotExist(err) {
		t.Error("AGENTS.md was not created by startSync")
	}
}