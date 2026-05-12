package ui

import (
	"os"
	"path/filepath"
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/storage"
	"strings"
	"testing"
)

func TestInstallFromStorageAndSyncCmd_MalformedYAML(t *testing.T) {
	// Setup temp storage
	tempDir, err := os.MkdirTemp("", "skillsync-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	skillID := "malformed-skill"
	skillDir := filepath.Join(tempDir, skillID)
	os.MkdirAll(skillDir, 0755)

	// Malformed YAML
	malformedContent := "---\nname: Test\n  invalid: mapping: values: not: allowed\n---"
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(malformedContent), 0644)

	m := NewModel(NewBackend(storage.NewService(tempDir)))
	m.rootPath = tempDir // Use tempDir as project root too for safety

	stored := storage.StoredSkill{
		ID: skillID,
		Metadata: storage.StoredMetadata{
			SkillName: "Malformed Skill",
		},
	}

	cmd := m.installFromStorageAndSyncCmd(stored)
	msg := cmd()

	res, ok := msg.(runner.SyncResult)
	if !ok {
		t.Fatalf("expected runner.SyncResult, got %T", msg)
	}

	if res.ExitCode == 0 {
		t.Error("expected non-zero exit code for malformed YAML")
	}

	expectedErr := "Malformed YAML: Edit and fix"
	if !strings.Contains(res.Stderr, expectedErr) {
		t.Errorf("expected error message to contain %q, got %q", expectedErr, res.Stderr)
	}
}
