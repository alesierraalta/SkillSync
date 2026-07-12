package parser

import (
	"os"
	"path/filepath"
	"skillsync/tui/internal/types"
	"testing"
)

func TestSaveTempFileCleanupOnPanicOrErrorTDD(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "SKILL.md")

	skill := &types.Skill{
		Name: "test-skill",
	}

	// We pass a path that resides inside a non-existent directory. os.WriteFile will fail.
	invalidPath := filepath.Join(tmpDir, "non-existent-dir", "SKILL.md")

	err := Save(invalidPath, skill)
	if err == nil {
		t.Fatal("Expected error when saving to invalid path")
	}

	// Verify the .tmp file is cleaned up (doesn't exist)
	tmpPath := invalidPath + ".tmp"
	if _, statErr := os.Stat(tmpPath); !os.IsNotExist(statErr) {
		t.Error("Expected temporary file to be deleted/cleaned up")
	}

	// Verify standard successful flow cleans it up
	if err := Save(path, skill); err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Stat(path + ".tmp"); !os.IsNotExist(statErr) {
		t.Error("Expected temporary file to be deleted/cleaned up after successful save")
	}
}
