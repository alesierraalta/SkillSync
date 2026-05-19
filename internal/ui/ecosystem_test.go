package ui

import (
	"bytes"
	"os"
	"path/filepath"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/types"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInstallCoreSkill(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(tmpDir)

	skills := []string{"skill-creator", "skill-sync", "find-skills"}

	for _, sk := range skills {
		backend := NewBackend(nil)
		err := backend.InstallCoreSkill(sk)
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
	backend := NewBackend(nil)
	if err := backend.InstallCoreSkill(sk); err != nil {
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

	storageRoot := t.TempDir()
	sService := storage.NewService(storageRoot)

	// Run instantiateEcosystemCmd
	backend := NewBackend(sService)
	cmd := instantiateEcosystemCmd(backend, "")
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

func TestInstantiateEcosystemCmd_RegistersProject(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	storageRoot := filepath.Join(tmpDir, "storage")
	sService := storage.NewService(storageRoot)

	// Run instantiateEcosystemCmd with service and path
	backend := NewBackend(sService)
	cmd := instantiateEcosystemCmd(backend, tmpDir)
	msg := cmd()

	// Should not error
	if em, ok := msg.(ecosystemMsg); ok && em.err != nil {
		t.Fatalf("instantiateEcosystemCmd failed: %v", em.err)
	}

	// Verify project was registered
	projects, err := sService.GetProjects()
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, p := range projects {
		if p.Path == tmpDir {
			found = true
			if !p.LastSynced.IsZero() {
				t.Errorf("expected LastSynced to be zero, got %v", p.LastSynced)
			}
			break
		}
	}

	if !found {
		t.Errorf("expected project %s to be registered", tmpDir)
	}
}

func TestInstallFromStorage(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(tmpDir)

	// Setup fake global storage
	storageRootStored := t.TempDir()
	sService := &storage.Service{RootPath: storageRootStored}

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
	_ = t.TempDir()
	backendSvc := NewBackend(sService) // Use the one with the skill!
	m := NewModel(backendSvc)
	m.storedSkills, _ = sService.List()
	m.Installer.AllStored = m.storedSkills
	m.Installer.StoredSkills = make([]bool, len(m.storedSkills))
	m.Installer.StoredSkills[0] = true // Select the one skill

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
	if !strings.Contains(string(content), skillContent) {
		t.Errorf("content mismatch. expected body %q to be in %q", skillContent, string(content))
	}
}

func TestNextInstallerStep_RegistersProject(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	storageRoot := filepath.Join(tmpDir, "storage")
	sService := storage.NewService(storageRoot)

	m := NewModel(NewBackend(sService))
	m.rootPath = tmpDir

	// Run step 0.6
	cmd := nextInstallerStep(m, 0.6)
	msg := cmd()

	progress, ok := msg.(installerProgressMsg)
	if !ok {
		t.Fatalf("expected installerProgressMsg, got %T", msg)
	}
	if progress.percent != 1.0 {
		t.Errorf("expected progress 1.0, got %f", progress.percent)
	}

	// Verify project was registered
	projects, err := sService.GetProjects()
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, p := range projects {
		absTmp, _ := filepath.Abs(tmpDir)
		if p.Path == absTmp {
			found = true
			if !p.LastSynced.IsZero() {
				t.Errorf("expected LastSynced to be zero, got %v", p.LastSynced)
			}
			break
		}
	}

	if !found {
		t.Errorf("expected project %s to be registered", tmpDir)
	}
}

func TestEcosystemInstantiation_EndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	storageRoot := filepath.Join(tmpDir, "storage")
	sService := storage.NewService(storageRoot)
	backend := NewBackend(sService)

	m := NewModel(backend)
	m.rootPath = tmpDir

	// Initialize sizes
	newModelSize, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	m = newModelSize.(Model)

	// 1. Initial state: Projects screen is empty
	m.Screen = ScreenProjects
	msg := m.loadProjectsCmd()()
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if len(m.projectList.Items()) != 0 {
		t.Errorf("expected 0 projects initially, got %d", len(m.projectList.Items()))
	}

	// 2. Trigger Init (Instantiate Ecosystem)
	_ = m.Init()
	// Init returns tea.Batch(loadSkills, instantiateEcosystemCmd)
	// We only care about the ecosystem one for this test.
	// In Bubble Tea testing, we'd normally use a simulator, but here we can just execute.

	// Simulate the ecosystemMsg arriving
	ecoMsg := instantiateEcosystemCmd(m.backend, m.rootPath)()
	newModel, _ = m.Update(ecoMsg)
	m = newModel.(Model)

	if m.StatusMsg != "Ecosistema instanciado" {
		t.Errorf("expected status 'Ecosistema instanciado', got %q", m.StatusMsg)
	}

	// 3. Navigate back to Projects and load
	msg = m.loadProjectsCmd()()
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	if len(m.projectList.Items()) != 1 {
		t.Errorf("expected 1 project after instantiation, got %d", len(m.projectList.Items()))
	}

	item := m.projectList.Items()[0].(projectItem)
	absTmp, _ := filepath.Abs(tmpDir)
	if item.project.Path != absTmp {
		t.Errorf("expected project path %s, got %s", absTmp, item.project.Path)
	}
	if !item.project.LastSynced.IsZero() {
		t.Errorf("expected LastSynced to be zero")
	}

	view := m.projectsView()
	if !strings.Contains(view, "Último sync: Nunca") {
		t.Errorf("expected view to show 'Nunca', got:\n%s", view)
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

	backend := NewBackend(nil)
	err := backend.RegisterOpenCodeTools()
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

	backend := NewBackend(nil)
	err := backend.RegisterSkillManagerAgent()
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

	backend := NewBackend(nil)
	// First call
	err := backend.RegisterSkillManagerAgent()
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Second call should not fail
	err = backend.RegisterSkillManagerAgent()
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
	backend := NewBackend(nil)
	_ = backend.RegisterOpenCodeTools()
	_ = backend.RegisterSkillManagerAgent()

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
	backend := NewBackend(nil)
	err := backend.RegisterOpenCodeTools()
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

func TestInstallCoreSkill_LFNormalization(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(tmpDir)

	// Setup: create a provider dir
	_ = os.MkdirAll(".agents/skills", 0755)

	// Run installation for a skill that has .sh files (skill-sync)
	backend := NewBackend(nil)
	err := backend.InstallCoreSkill("skill-sync")
	if err != nil {
		t.Fatalf("InstallCoreSkill failed: %v", err)
	}

	// Verify sync.sh has LF line endings
	syncSh := filepath.Join(".agents", "skills", "skill-sync", "assets", "sync.sh")
	content, err := os.ReadFile(syncSh)
	if err != nil {
		t.Fatalf("failed to read sync.sh: %v", err)
	}

	if bytes.Contains(content, []byte("\r\n")) {
		t.Errorf("sync.sh contains CRLF, expected only LF")
	}
}

func TestInstallCoreSharedLib_RepairCRLF(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(tmpDir)

	libDir := filepath.Join(".agents", "skills", "lib")
	_ = os.MkdirAll(libDir, 0755)
	utilsPath := filepath.Join(libDir, "utils.sh")

	// 1. Create a file with CRLF and some user edit
	userContent := "# User edit\r\necho 'hello'\r\n"
	err := os.WriteFile(utilsPath, []byte(userContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 2. Run repair
	backend := NewBackend(nil)
	err = backend.installCoreSharedLib()
	if err != nil {
		t.Fatalf("installCoreSharedLib failed: %v", err)
	}

	// 3. Verify it was repaired (LF) but preserved user content
	content, err := os.ReadFile(utilsPath)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Contains(content, []byte("\r\n")) {
		t.Errorf("utils.sh still contains CRLF")
	}

	expected := "# User edit\necho 'hello'\n"
	if string(content) != expected {
		t.Errorf("expected content %q, got %q", expected, string(content))
	}
}
