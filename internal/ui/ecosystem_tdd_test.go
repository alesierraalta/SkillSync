package ui

import (
	"os"
	"path/filepath"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/types"
	"testing"
)

type mockBackend struct {
	AppService // embed interface to satisfy it partially
}

func (m *mockBackend) InstallCoreSkill(name string) error        { return nil }
func (m *mockBackend) RegisterOpenCodeTools() error              { return nil }
func (m *mockBackend) RegisterSkillManagerAgent() error          { return nil }
func (m *mockBackend) RegisterProjectInitial(path string) error  { return nil }
func (m *mockBackend) EnsureAgentsMD(root string) error          { return nil }
func (m *mockBackend) LoadFromStorage(id string) (string, error) { return "", nil }
func (m *mockBackend) ParseSkillContent(content string) (*types.Skill, error) {
	return &types.Skill{}, nil
}

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
	m.Installer.AllStored = m.storedSkills
	m.Installer.StoredSkills = make([]bool, len(m.storedSkills))

	// Select both
	for i := range m.Installer.StoredSkills {
		m.Installer.StoredSkills[i] = true
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

func TestAutoskillsSequencing(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(tmpDir)

	// Mock backend to avoid panics
	backend := &mockBackend{}
	m := NewModel(backend)
	m.Installer.Autoskills = true
	// Disable providers to avoid file writing and more backend calls
	for i := range m.Installer.Providers {
		m.Installer.Providers[i] = false
	}

	// Step 0.5 -> 0.6
	// In the WRONG sequence, 0.6 is "Sincronizando configuraciones..."
	// In the RIGHT sequence, 0.6 should be "Ejecutando Smart Scan (autoskills)..."
	cmd := nextInstallerStep(m, 0.5)
	msg := cmd()
	progress, _ := msg.(installerProgressMsg)
	if progress.percent != 0.6 {
		t.Fatalf("expected 0.6, got %f", progress.percent)
	}
	if progress.task != "Ejecutando Smart Scan (autoskills)..." {
		t.Errorf("expected task 'Ejecutando Smart Scan (autoskills)...' at 0.6, got %q", progress.task)
	}

	// Step 0.6 -> 0.8
	// In the WRONG sequence, 0.6 leads to 0.8 which is "Ejecutando Smart Scan (autoskills)..."
	// In the RIGHT sequence, 0.6 leads to 0.8 which should be "Sincronizando configuraciones..."
	cmd = nextInstallerStep(m, 0.6)
	msg = cmd()
	progress, _ = msg.(installerProgressMsg)
	if progress.percent != 0.8 {
		t.Fatalf("expected 0.8, got %f", progress.percent)
	}
	if progress.task != "Sincronizando configuraciones..." {
		t.Errorf("expected task 'Sincronizando configuraciones...' at 0.8, got %q", progress.task)
	}
}
