package ui

import (
	"os"
	"path/filepath"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/types"
	"strings"
	"testing"
)

func TestInstallCoreSkill(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(tmpDir)

	skills := []string{"skill-creator", "skill-sync", "find-skills"}

	for _, sk := range skills {
		err := InstallCoreSkill(sk)
		if err != nil {
			t.Fatalf("installCoreSkill(%s) failed: %v", sk, err)
		}

		skillFile := filepath.Join(".agents", "skills", sk, "SKILL.md")
		if _, err := os.Stat(skillFile); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", skillFile)
		}

		content, err := os.ReadFile(skillFile)
		if err != nil {
			t.Fatalf("failed to read %s: %v", skillFile, err)
		}

		// Basic validation that it's not the placeholder
		if !strings.Contains(string(content), "name: "+sk) {
			t.Errorf("SKILL.md for %s seems to be a placeholder or missing metadata", sk)
		}

		// Harden: metadata parsing check (description and scope)
		if strings.Contains(sk, "sync") || strings.Contains(sk, "creator") {
			if !strings.Contains(string(content), "scope:") {
				t.Errorf("SKILL.md for %s missing metadata scope", sk)
			}
		}
		if !strings.Contains(string(content), "description:") {
			t.Errorf("SKILL.md for %s missing description", sk)
		}

		if sk == "skill-sync" {
			assetFile := filepath.Join(".agents", "skills", sk, "assets", "sync.sh")
			if _, err := os.Stat(assetFile); os.IsNotExist(err) {
				t.Errorf("expected asset %s to exist", assetFile)
			}
		}
	}
}


func TestInstallFromStorage(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(tmpDir)

	// Setup fake global storage
	storageRoot := t.TempDir()
	sService := &storage.Service{RootPath: storageRoot}
	
	skillName := "stored-skill"
	skillContent := "name: stored-skill\ndescription: from storage"
	
	err := sService.Save(&types.Skill{
		Name:    skillName,
		RawBody: skillContent,
	}, storage.StoredMetadata{
		SkillName:   skillName,
		Description: "from storage",
	})
	if err != nil {
		t.Fatalf("setup storage failed: %v", err)
	}

	// Setup model for installer
	m := NewModel()
	m.storageService = sService
	m.storedSkills, _ = sService.List()
	m.installerStoredSkills = make([]bool, len(m.storedSkills))
	m.installerStoredSkills[0] = true // Select the one skill

	// Run step 0.5 (Install from storage)
	cmd := nextInstallerStep(m, 0.5)
	msg := cmd()

	progress, ok := msg.(installerProgressMsg)
	if !ok {
		t.Fatalf("expected installerProgressMsg, got %T", msg)
	}
	if progress.percent != 0.6 {
		t.Errorf("expected progress 0.6, got %f", progress.percent)
	}

	// Verify file was written
	destFile := filepath.Join(".agents", "skills", skillName, "SKILL.md")
	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read installed skill: %v", err)
	}
	if string(content) != skillContent {
		t.Errorf("content mismatch. expected %q, got %q", skillContent, string(content))
	}
}



