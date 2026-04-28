package ui

import (
	"os"
	"strings"
	"testing"
)

func TestModelRefreshAfterInstall(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(tmpDir)

	// Create a dummy AGENTS.md to satisfy loadSkills logic if needed
	_ = os.WriteFile("AGENTS.md", []byte("# Agents"), 0644)

	m := NewModel()
	m.rootPath = tmpDir
	m.Screen = ScreenInstaller

	// 1. Simulate successful installation of core skills
	// First, manually install one so we can verify it's loaded
	err := InstallCoreSkill("skill-sync")
	if err != nil {
		t.Fatalf("failed to install core skill: %v", err)
	}

	// 2. Send installerFinishedMsg
	msg := installerFinishedMsg{err: nil}
	newModel, cmd := m.Update(msg)
	
	m2 := newModel.(Model)

	// 3. Verify screen transition
	if m2.Screen != ScreenHome {
		t.Errorf("expected screen to be ScreenHome, got %v", m2.Screen)
	}

	if m2.StatusMsg != "Instalación completada con éxito" {
		t.Errorf("unexpected status msg: %s", m2.StatusMsg)
	}

	// 4. Verify command is returned (loadSkills)
	if cmd == nil {
		t.Fatal("expected loadSkills command, got nil")
	}

	// 5. Execute command and verify skillsLoadedMsg
	rawMsg := cmd()
	loadedMsg, ok := rawMsg.(skillsLoadedMsg)
	if !ok {
		t.Fatalf("expected skillsLoadedMsg, got %T", rawMsg)
	}

	// 6. Update model with loaded skills
	newModel, _ = m2.Update(loadedMsg)
	m3 := newModel.(Model)

	// 7. Verify skills are in the list
	items := m3.list.Items()
	found := false
	for _, it := range items {
		if si, ok := it.(item); ok {
			if strings.Contains(si.skill.ID, "skill-sync") || strings.Contains(si.skill.Path, "skill-sync") {
				found = true
				if si.skill.Metadata.Description == "" {
					t.Errorf("skill-sync loaded but has empty description")
				}
				break
			}
		}
	}

	if !found {
		t.Errorf("skill-sync not found in model after refresh")
	}
}
