package ui

import (
	"os"
	"path/filepath"
	"skillsync/tui/internal/storage"
	"testing"
)

func TestNextInstallerStep_BypassMalformedYAML(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(tmpDir)

	// Setup fake global storage with TWO skills:
	// 1. Malformed
	// 2. Good
	storageRoot := t.TempDir()
	sService := &storage.Service{RootPath: storageRoot}

	// Malformed skill
	malformedName := "malformed-skill"
	malformedContent := "---\nname: Malformed\n  invalid: mapping\n---"
	os.MkdirAll(filepath.Join(storageRoot, malformedName), 0755)
	os.WriteFile(filepath.Join(storageRoot, malformedName, "SKILL.md"), []byte(malformedContent), 0644)
	os.WriteFile(filepath.Join(storageRoot, malformedName, "METADATA.json"), []byte(`{"skill_name": "Malformed Skill"}`), 0644)

	// Good skill
	goodName := "good-skill"
	goodContent := "---\nname: Good\ndescription: I am good\n---\nBody content"
	os.MkdirAll(filepath.Join(storageRoot, goodName), 0755)
	os.WriteFile(filepath.Join(storageRoot, goodName, "SKILL.md"), []byte(goodContent), 0644)
	os.WriteFile(filepath.Join(storageRoot, goodName, "METADATA.json"), []byte(`{"skill_name": "Good Skill"}`), 0644)

	// Setup model for installer
	backendSvc := NewBackend(sService)
	m := NewModel(backendSvc)
	m.storedSkills, _ = sService.List()
	m.installerStoredSkills = make([]bool, len(m.storedSkills))

	// Select both
	for i := range m.installerStoredSkills {
		m.installerStoredSkills[i] = true
	}

	// Run step 0.5
	cmd := nextInstallerStep(m, 0.5)
	msg := cmd()

	// Verify it didn't fail (RED if it returns installerFinishedMsg{err: ...})
	if finished, ok := msg.(installerFinishedMsg); ok && finished.err != nil {
		t.Fatalf("expected nextInstallerStep to bypass malformed YAML, but it failed: %v", finished.err)
	}

	progress, ok := msg.(installerProgressMsg)
	if !ok {
		t.Fatalf("expected installerProgressMsg, got %T", msg)
	}
	if progress.percent != 0.6 {
		t.Errorf("expected progress 0.6, got %f", progress.percent)
	}

	// Verify Good skill was written
	goodFile := filepath.Join(".agents", "skills", "Good Skill", "SKILL.md")
	if _, err := os.Stat(goodFile); os.IsNotExist(err) {
		t.Errorf("expected good skill to be written to %s", goodFile)
	}

	// Verify Malformed skill was NOT written (or at least we didn't crash)
	malformedFile := filepath.Join(".agents", "skills", "Malformed Skill", "SKILL.md")
	if _, err := os.Stat(malformedFile); err == nil {
		t.Errorf("expected malformed skill to be bypassed (not written to %s)", malformedFile)
	}
}
