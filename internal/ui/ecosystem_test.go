package ui

import (
	"os"
	"path/filepath"
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



