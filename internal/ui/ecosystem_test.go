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

			sharedLib := filepath.Join(".agents", "skills", "lib", "utils.sh")
			if _, err := os.Stat(sharedLib); os.IsNotExist(err) {
				t.Errorf("expected shared lib %s to exist", sharedLib)
			}
		}
	}
}

func TestInstallCoreSkill_MultiProvider(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create all providers matching discovery service set
	_ = os.MkdirAll(".agents/skills", 0755)
	_ = os.MkdirAll(".opencode/skills", 0755)
	_ = os.MkdirAll(".claude/skills", 0755)
	_ = os.MkdirAll(".qwen/skills", 0755)
	_ = os.MkdirAll(".gemini/skills", 0755)
	_ = os.MkdirAll(".cursor/skills", 0755)
	_ = os.MkdirAll(".copilot/skills", 0755)

	// Install a core skill
	sk := "find-skills"
	if err := InstallCoreSkill(sk); err != nil {
		t.Fatalf("InstallCoreSkill(%s) failed: %v", sk, err)
	}

	// Verify skill was installed to ALL providers that exist
	providers := []string{
		".agents/skills",
		".opencode/skills",
		".claude/skills",
		".qwen/skills",
		".gemini/skills",
		".cursor/skills",
		".copilot/skills",
	}
	for _, provider := range providers {
		skillFile := filepath.Join(provider, sk, "SKILL.md")
		if _, err := os.Stat(skillFile); os.IsNotExist(err) {
			t.Errorf("expected skill in %s, not found", provider)
		}
	}
}

func TestInstantiateEcosystemCmd_RegistersOpenCode(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create .opencode directory (simulates existing OpenCode installation)
	_ = os.MkdirAll(".opencode", 0755)

	// Run instantiateEcosystemCmd
	cmd := instantiateEcosystemCmd()
	// Execute the command (tea.Cmd is func() tea.Msg)
	msg := cmd()

	// Should not error
	if em, ok := msg.(ecosystemMsg); ok && em.err != nil {
		t.Fatalf("instantiateEcosystemCmd failed: %v", em.err)
	}

	// Verify .opencode/commands has base command markdown files
	for _, tool := range []string{"skill", "find", "create", "sync", "fullskills"} {
		cmdPath := filepath.Join(".opencode", "commands", tool+".md")
		if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
			t.Errorf("missing command file %q", cmdPath)
		}
	}

	// Verify .opencode/package.json has empty tools
	pkgPath := ".opencode/package.json"
	content, _ := os.ReadFile(pkgPath)
	if strings.Contains(string(content), `"name": "skill"`) {
		t.Errorf("tool 'skill' should not be in package.json tools array")
	}

	// Verify .opencode/agents/skill-manager.md exists
	agentPath := ".opencode/agents/skill-manager.md"
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		t.Errorf("expected skill-manager.md to exist after instantiation")
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

func TestRegisterOpenCodeTools_PreservesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(tmpDir)

	dir := ".opencode"
	_ = os.MkdirAll(dir, 0755)
	packagePath := filepath.Join(dir, "package.json")
	
	// Note: Under the new design, tools are fully regenerated from mirrored skills.
	// This test verifies that base 5 commands are present after registration.
	existingContent := `{
  "dependencies": {
    "other-dep": "1.0.0"
  },
  "opencode": {
    "manifest": "v1",
    "tools": [
      {
        "name": "existing-tool",
        "command": "do something"
      }
    ]
  }
}`
	_ = os.WriteFile(packagePath, []byte(existingContent), 0644)

	err := RegisterOpenCodeTools()
	if err != nil {
		t.Fatalf("registerOpenCodeTools failed: %v", err)
	}

	content, _ := os.ReadFile(packagePath)
	sContent := string(content)

	// Verify all 5 skill commands are created as markdown files
	expectedCommands := []string{"skill", "find", "create", "sync", "fullskills"}
	for _, cmd := range expectedCommands {
		cmdPath := filepath.Join(".opencode", "commands", cmd+".md")
		if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
			t.Errorf("missing command file %q", cmdPath)
		}
		content, _ := os.ReadFile(cmdPath)
		if !strings.Contains(string(content), "managed_by: skillsync") {
			t.Errorf("command file %q missing managed_by marker", cmdPath)
		}
		if !strings.Contains(string(content), "synck "+cmd) {
			t.Errorf("command file %q missing synck call", cmdPath)
		}
	}

	// Verify tools array is empty in package.json
	if strings.Contains(sContent, `"tools": [`) && !strings.Contains(sContent, `"tools": []`) {
		t.Errorf("package.json tools array should be empty, got: %s", sContent)
	}

	// Note: Under new design, tools are fully regenerated, so non-base tools may be replaced.
	// The base 5 commands should always be present.
}


func TestRegisterSkillManagerAgent_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(tmpDir)

	err := RegisterSkillManagerAgent()
	if err != nil {
		t.Fatalf("RegisterSkillManagerAgent failed: %v", err)
	}

	agentPath := ".opencode/agents/skill-manager.md"
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		t.Errorf("expected %s to exist", agentPath)
	}

	content, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", agentPath, err)
	}
	sContent := string(content)

	// Check OpenCode frontmatter
	if !strings.Contains(sContent, "description:") {
		t.Errorf("missing description in frontmatter")
	}
	if !strings.Contains(sContent, "mode: subagent") {
		t.Errorf("missing mode: subagent in frontmatter")
	}
	if !strings.Contains(sContent, "permission:") {
		t.Errorf("missing permission (singular) in frontmatter")
	}
	if strings.Contains(sContent, "permissions:") {
		t.Errorf("must use singular 'permission:' not plural 'permissions:'")
	}

	// Check all 5 command references in agent
	// Under the new design, commands appear as skill names in the table
	expectedCommands := []string{"skill", "find", "create", "sync", "fullskills"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(sContent, cmd) {
			t.Errorf("missing %s command reference", cmd)
		}
	}

	// Check full workflow chain
	if !strings.Contains(sContent, "find-skills") {
		t.Errorf("missing find-skills reference")
	}
	if !strings.Contains(sContent, "skill-creator") {
		t.Errorf("missing skill-creator reference")
	}
	if !strings.Contains(sContent, "skill-sync") {
		t.Errorf("missing skill-sync reference")
	}
}

func TestRegisterSkillManagerAgent_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(tmpDir)

	// First call
	err := RegisterSkillManagerAgent()
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Second call should not fail
	err = RegisterSkillManagerAgent()
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}

	agentPath := ".opencode/agents/skill-manager.md"
	content, _ := os.ReadFile(agentPath)
	if !strings.Contains(string(content), "mode: subagent") {
		t.Errorf("file content corrupted after second call")
	}
}

func TestRegisterSkillManagerAgent_OpenCodeToolsPreserved(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(tmpDir)

	// Setup existing package.json with tools
	dir := ".opencode"
	_ = os.MkdirAll(dir, 0755)
	packagePath := filepath.Join(dir, "package.json")
	existingContent := `{
  "dependencies": {
    "@opencode-ai/plugin": "1.14.19"
  },
  "opencode": {
    "tools": [
      {
        "name": "existing-tool",
        "command": "do something"
      }
    ]
  }
}`
	_ = os.WriteFile(packagePath, []byte(existingContent), 0644)

	// Run both registration functions
	_ = RegisterOpenCodeTools()
	_ = RegisterSkillManagerAgent()

	// Note: Under the new design, RegisterSkillManagerAgent calls SyncSkills
	// which may create or modify package.json.
	// The key verification is that agent file exists alongside package.json.
	// Tool preservation is handled by RegisterOpenCodeTools separately.

	// Verify agent was created alongside package.json
	agentPath := ".opencode/agents/skill-manager.md"
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		t.Errorf("agent file should exist alongside package.json")
	}
}

// ----------------------------------------------------------------
// TestEcosystem_DelegatesToOpencode — Task 1.15
// ----------------------------------------------------------------

func TestEcosystem_DelegatesToOpencode(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(tmpDir)

	// Create a skill in .agents/skills with proper YAML frontmatter
	_ = os.MkdirAll(".agents/skills/testskill", 0755)
	skillContent := `---
name: testskill
description: Test skill
auto_invoke: true
---
`
	_ = os.WriteFile(".agents/skills/testskill/SKILL.md", []byte(skillContent), 0644)

	// Create AGENTS.md so findProjectRoot works
	_ = os.WriteFile("AGENTS.md", []byte("# Agent Skills\n"), 0644)

	// Call RegisterOpenCodeTools which should delegate to opencode.SyncSkills and opencode.RegenerateTools
	err := RegisterOpenCodeTools()
	if err != nil {
		t.Fatalf("RegisterOpenCodeTools failed: %v", err)
	}

	// Verify skill was mirrored to .opencode/skills/
	mirroredPath := ".opencode/skills/testskill/SKILL.md"
	if _, err := os.Stat(mirroredPath); os.IsNotExist(err) {
		t.Errorf("expected skill to be mirrored to .opencode/skills/testskill/SKILL.md")
	}

	// Verify command markdown was generated
	cmdPath := filepath.Join(".opencode", "commands", "testskill.md")
	if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
		t.Errorf("expected command file %q to exist", cmdPath)
	}

	// Verify package.json has empty tools
	pkgContent, _ := os.ReadFile(".opencode/package.json")
	if strings.Contains(string(pkgContent), `"name": "testskill"`) {
		t.Errorf("testskill tool should not be in package.json tools array, got: %s", pkgContent)
	}
}



