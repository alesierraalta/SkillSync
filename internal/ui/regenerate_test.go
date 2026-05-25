package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestSkill(t *testing.T, path, name, desc string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	content := "---\ndescription: " + desc + "\nauto_invoke: true\nscope: root\n---\n\nBody of " + name + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestRegenerateAfterDelete_FullIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agents/skills/test-skill/SKILL.md
	writeTestSkill(t, filepath.Join(tmpDir, ".agents", "skills", "test-skill", "SKILL.md"), "test-skill", "A test skill")
	// Create .agents/skills/other-skill/SKILL.md
	writeTestSkill(t, filepath.Join(tmpDir, ".agents", "skills", "other-skill", "SKILL.md"), "other-skill", "Another test skill")

	// Create AGENTS.md with proper headers and both skills
	agentsContent := `# Agent Skills

This document lists the AI skills available in the TUI project.

## Available Skills

| Skill | Description | Location |
| ----- | ----------- | -------- |
| ` + "`test-skill`" + ` | A test skill | [SKILL.md](.agents/skills/test-skill/SKILL.md) |
| ` + "`other-skill`" + ` | Another test skill | [SKILL.md](.agents/skills/other-skill/SKILL.md) |

### Auto-invoke Skills

When performing these actions, ALWAYS invoke the corresponding skill FIRST:

| Action                              | Skill      |
| ----------------------------------- | ---------- |
| test-skill                          | ` + "`test-skill`" + ` |
| other-skill                         | ` + "`other-skill`" + ` |
`
	if err := os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte(agentsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create mirrors in .opencode/skills/
	writeTestSkill(t, filepath.Join(tmpDir, ".opencode", "skills", "test-skill", "SKILL.md"), "test-skill", "A test skill")
	writeTestSkill(t, filepath.Join(tmpDir, ".opencode", "skills", "other-skill", "SKILL.md"), "other-skill", "Another test skill")

	// Create .opencode/package.json
	pkgDir := filepath.Join(tmpDir, ".opencode")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}
	pkgContent := `{
  "opencode": {
    "tools": []
  }
}`
	if err := os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte(pkgContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .opencode/agents/skill-manager.md
	agentsDir := filepath.Join(tmpDir, ".opencode", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentsDir, "skill-manager.md"), []byte("placeholder"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .opencode/commands/ with managed_by markers
	// Filenames must match the skill name (RegenerateCommands uses {skill.Name}.md)
	cmdsDir := filepath.Join(tmpDir, ".opencode", "commands")
	if err := os.MkdirAll(cmdsDir, 0755); err != nil {
		t.Fatal(err)
	}
	cmdTestContent := "---\nmanaged_by: skillsync\ndescription: Test skill command\n---\n\n/test-skill \"$!1\"\n"
	if err := os.WriteFile(filepath.Join(cmdsDir, "test-skill.md"), []byte(cmdTestContent), 0644); err != nil {
		t.Fatal(err)
	}
	cmdOtherContent := "---\nmanaged_by: skillsync\ndescription: Other skill command\n---\n\n/other-skill \"$!1\"\n"
	if err := os.WriteFile(filepath.Join(cmdsDir, "other-skill.md"), []byte(cmdOtherContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Simulate RemoveByID: delete canonical test-skill only
	if err := os.RemoveAll(filepath.Join(tmpDir, ".agents", "skills", "test-skill")); err != nil {
		t.Fatal(err)
	}

	// Call RegenerateAfterDelete
	err := RegenerateAfterDelete(tmpDir)
	if err != nil {
		t.Fatalf("RegenerateAfterDelete failed: %v", err)
	}

	// Verify AGENTS.md no longer contains test-skill
	agentsAfter, err := os.ReadFile(filepath.Join(tmpDir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(agentsAfter), "test-skill") {
		t.Error("AGENTS.md should not contain test-skill")
	}
	if !strings.Contains(string(agentsAfter), "other-skill") {
		t.Error("AGENTS.md should contain other-skill")
	}

	// Verify .opencode/skills/test-skill/ is pruned (orphan mirror)
	if _, err := os.Stat(filepath.Join(tmpDir, ".opencode", "skills", "test-skill")); !os.IsNotExist(err) {
		t.Error(".opencode/skills/test-skill should have been pruned")
	}

	// Verify .opencode/skills/other-skill/ still exists
	if _, err := os.Stat(filepath.Join(tmpDir, ".opencode", "skills", "other-skill")); os.IsNotExist(err) {
		t.Error(".opencode/skills/other-skill should still exist")
	}

	// Verify orphan command is pruned (filename = skill name)
	if _, err := os.Stat(filepath.Join(tmpDir, ".opencode", "commands", "test-skill.md")); !os.IsNotExist(err) {
		t.Error("test-skill.md command should have been pruned")
	}

	// Verify other-skill command still exists
	if _, err := os.Stat(filepath.Join(tmpDir, ".opencode", "commands", "other-skill.md")); os.IsNotExist(err) {
		t.Error("other-skill.md command should still exist")
	}

	// Verify OPENCODE.md exists and matches AGENTS.md content
	opencodeAfter, err := os.ReadFile(filepath.Join(tmpDir, "OPENCODE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(opencodeAfter) != string(agentsAfter) {
		t.Error("OPENCODE.md should match AGENTS.md content")
	}
}

func TestRegenerateAfterDelete_EmptySkillsDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty .agents/skills/
	if err := os.MkdirAll(filepath.Join(tmpDir, ".agents", "skills"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create .opencode/ directory structure so regen tools can write
	if err := os.MkdirAll(filepath.Join(tmpDir, ".opencode"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create AGENTS.md with proper headers (empty tables)
	agentsContent := `# Agent Skills

## Available Skills

| Skill | Description | Location |
| ----- | ----------- | -------- |

### Auto-invoke Skills

When performing these actions, ALWAYS invoke the corresponding skill FIRST:

| Action                              | Skill      |
| ----------------------------------- | ---------- |
`
	if err := os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte(agentsContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := RegenerateAfterDelete(tmpDir)
	if err != nil {
		t.Fatalf("RegenerateAfterDelete failed with empty skills dir: %v", err)
	}

	// AGENTS.md should still have the headers
	agentsAfter, err := os.ReadFile(filepath.Join(tmpDir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(agentsAfter), "## Available Skills") {
		t.Error("AGENTS.md should retain headers after regen with empty skills")
	}
	if !strings.Contains(string(agentsAfter), "### Auto-invoke Skills") {
		t.Error("AGENTS.md should retain auto-invoke headers after regen with empty skills")
	}
}

func TestRegenerateAfterDelete_MissingHeaders(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .agents/skills/test-skill
	writeTestSkill(t, filepath.Join(tmpDir, ".agents", "skills", "test-skill", "SKILL.md"), "test-skill", "A test skill")

	// Create mirror in .opencode
	writeTestSkill(t, filepath.Join(tmpDir, ".opencode", "skills", "test-skill", "SKILL.md"), "test-skill", "A test skill")

	// Create .opencode/package.json (needed for tools regen)
	pkgDir := filepath.Join(tmpDir, ".opencode")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}
	pkgContent := `{
  "opencode": {
    "tools": []
  }
}`
	if err := os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte(pkgContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create AGENTS.md WITHOUT proper headers (missing ## Available Skills and ### Auto-invoke Skills)
	agentsContent := "# Agent Skills\n\nSome content without the required sections.\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte(agentsContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := RegenerateAfterDelete(tmpDir)
	// Should have an error from UpdateAgentsMarkdown but not crash
	if err == nil {
		t.Error("expected non-nil error due to missing AGENTS.md headers")
	} else if !strings.Contains(err.Error(), "headers missing") {
		t.Errorf("expected error about missing headers, got: %v", err)
	}

	// Verify other steps still ran - test-skill mirror should still exist (no deletion happened)
	if _, err := os.Stat(filepath.Join(tmpDir, ".opencode", "skills", "test-skill")); os.IsNotExist(err) {
		t.Error(".opencode/skills/test-skill should still exist")
	}

	// Verify package.json was at least created/updated (OpenCode tools regen ran)
	if _, err := os.Stat(filepath.Join(tmpDir, ".opencode", "package.json")); os.IsNotExist(err) {
		t.Error(".opencode/package.json should exist")
	}
}
