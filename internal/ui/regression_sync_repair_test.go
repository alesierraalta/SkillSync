package ui

import (
	"os"
	"path/filepath"
	"skillsync/tui/internal/storage"
	"testing"
)

func TestSyncRepairsMissingSharedLib(t *testing.T) {
	// 1. Setup a temp directory representing an external project
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to tmpDir: %v", err)
	}

	// 2. Setup skill-sync without the shared lib
	// We need to create the directory structure but omit .agents/skills/lib/utils.sh
	syncScriptPath := filepath.Join(".agents", "skills", "skill-sync", "assets", "sync.sh")
	if err := os.MkdirAll(filepath.Dir(syncScriptPath), 0755); err != nil {
		t.Fatalf("failed to create sync script dir: %v", err)
	}

	// Create a dummy sync.sh that checks for utils.sh
	dummySyncContent := `#!/bin/bash
if [ ! -f "../../lib/utils.sh" ]; then
  echo "Error: Could not find lib/utils.sh"
  exit 1
fi
echo "Syncing Skills..."
exit 0
`
	if err := os.WriteFile(syncScriptPath, []byte(dummySyncContent), 0755); err != nil {
		t.Fatalf("failed to write dummy sync.sh: %v", err)
	}

	// Ensure .agents/skills/lib exists but is empty (or utils.sh is missing)
	libDir := filepath.Join(".agents", "skills", "lib")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		t.Fatalf("failed to create lib dir: %v", err)
	}

	// 3. Initialize Model and start sync
	// The startSync() method uses "./.agents/skills/skill-sync/assets/sync.sh" relative to cwd
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenSyncing

	cmd := m.startSync()
	msg := cmd()

	res, ok := msg.(syncReportMsg)
	if !ok {
		t.Fatalf("expected syncReportMsg, got %T", msg)
	}

	// 4. Assertions
	// Before the fix, this should fail with error
	// After the fix, it should succeed (err == nil) because startSync repaired it.
	if res.err != nil {
		t.Errorf("Sync failed with error: %v. Expected repair and success.", res.err)
	}

	// Verify shared lib was actually created
	sharedLib := filepath.Join(".agents", "skills", "lib", "utils.sh")
	if _, err := os.Stat(sharedLib); os.IsNotExist(err) {
		t.Error("shared lib was not repaired/created by startSync")
	}
}
