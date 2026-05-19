package ui

import (
	"os"
	"path/filepath"
	"skillsync/tui/internal/storage"
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
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenSyncing

	// This should now call ensureAgentsMD() before running the script
	cmd := m.startSync()
	msg := cmd()

	res, ok := msg.(syncReportMsg)
	if !ok {
		t.Fatalf("expected syncReportMsg, got %T", msg)
	}

	// 4. Assertions
	if res.err != nil {
		t.Errorf("Sync failed with error: %v. Expected AGENTS.md creation and success.", res.err)
	}

	// Verify AGENTS.md was created
	if _, err := os.Stat("AGENTS.md"); os.IsNotExist(err) {
		t.Error("AGENTS.md was not created by startSync")
	}
}
