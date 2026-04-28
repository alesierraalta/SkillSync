package parser_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skillsync/tui/internal/parser"
	"skillsync/tui/internal/ui"
)

func TestCoreSkillInstallAndParse(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(tmpDir)

	skills := []string{"skill-creator", "skill-sync", "find-skills"}

	for _, sk := range skills {
		t.Run(sk, func(t *testing.T) {
			err := ui.InstallCoreSkill(sk)
			if err != nil {
				t.Fatalf("InstallCoreSkill(%s) failed: %v", sk, err)
			}

			skillFile := filepath.Join(".agents", "skills", sk, "SKILL.md")
			
			// 1. Assert file exists and content is full
			content, err := os.ReadFile(skillFile)
			if err != nil {
				t.Fatalf("failed to read installed SKILL.md for %s: %v", sk, err)
			}

			// Ensure it's not a stub (stub usually doesn't have metadata sections)
			if !strings.Contains(string(content), "description:") {
				t.Errorf("SKILL.md for %s appears to be a stub (missing description)", sk)
			}

			// 2. Assert assets for skill-sync
			if sk == "skill-sync" {
				assetPath := filepath.Join(".agents", "skills", sk, "assets", "sync.sh")
				if _, err := os.Stat(assetPath); os.IsNotExist(err) {
					t.Errorf("missing asset: %s", assetPath)
				}
			}

			// 3. Parse and assert metadata
			parsed, err := parser.Parse(skillFile)
			if err != nil {
				t.Fatalf("Parse(%s) failed: %v", skillFile, err)
			}

			if parsed.Metadata.Description == "" || parsed.Metadata.Description == "No description provided" {
				t.Errorf("parsed description is empty or default for %s", sk)
			}

			if parsed.Metadata.Scope == "" || parsed.Metadata.Scope == "No scope specified" {
				t.Errorf("parsed scope is empty or default for %s", sk)
			}
			
			// Canonical phrase check (Regression/no-stub contract)
			canonicalPhrases := map[string]string{
				"skill-creator": "Creates new AI agent skills following the Agent Skills spec",
				"skill-sync":    "Syncs skill metadata to AGENTS.md",
				"find-skills":   "Helps users discover and install agent skills",
			}
			if phrase, ok := canonicalPhrases[sk]; ok {
				if !strings.Contains(parsed.Metadata.Description, phrase) && !strings.Contains(parsed.RawBody, phrase) {
					t.Errorf("canonical phrase '%s' not found in %s", phrase, sk)
				}
			}
		})
	}
}
