package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRemove_DeletesCanonical(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create minimal project
	_ = os.MkdirAll(".agents/skills/test-skill", 0755)
	_ = os.WriteFile(".agents/skills/test-skill/SKILL.md", []byte("name: test-skill\n"), 0644)
	_ = os.WriteFile("AGENTS.md", []byte("# Agent Skills\n"), 0644)

	err := runRemove("test-skill", false, true)
	if err != nil {
		t.Fatalf("runRemove failed: %v", err)
	}

	// Verify skill directory is gone
	if _, err := os.Stat(".agents/skills/test-skill"); !os.IsNotExist(err) {
		t.Error("expected skill directory to be deleted")
	}
}

func TestRunRemove_CoreSkillRejected(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	_ = os.MkdirAll(".agents", 0755)
	_ = os.WriteFile("AGENTS.md", []byte("# Agent Skills\n"), 0644)

	err := runRemove("skill-creator", false, true)
	if err == nil {
		t.Fatal("expected error for core skill removal")
	}
	if !strings.Contains(err.Error(), "core") {
		t.Errorf("expected error containing 'core', got: %v", err)
	}
}

func TestRunRemove_MissingSkill(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	_ = os.MkdirAll(".agents", 0755)
	_ = os.WriteFile("AGENTS.md", []byte("# Agent Skills\n"), 0644)

	err := runRemove("nonexistent", false, true)
	if err == nil {
		t.Fatal("expected error for missing skill")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestRunRemove_ConfirmationRejected(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	skillDir := filepath.Join(".agents", "skills", "test-skill")
	_ = os.MkdirAll(skillDir, 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("name: test-skill\n"), 0644)
	_ = os.WriteFile("AGENTS.md", []byte("# Agent Skills\n"), 0644)

	// Mock stdin: answer "n" to the prompt
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	done := make(chan struct{})
	go func() {
		_, _ = w.Write([]byte("n\n"))
		_ = w.Close()
		close(done)
	}()

	err := runRemove("test-skill", false, false)

	os.Stdin = oldStdin
	<-done

	if err != nil {
		t.Fatalf("runRemove on cancel should not return error, got: %v", err)
	}

	// Verify skill directory still exists
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		t.Error("expected skill to still exist after cancel")
	}
}

func TestRunRemove_ConfirmationAccepted(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	skillDir := filepath.Join(".agents", "skills", "test-skill")
	_ = os.MkdirAll(skillDir, 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("name: test-skill\n"), 0644)
	_ = os.WriteFile("AGENTS.md", []byte("# Agent Skills\n"), 0644)

	// Mock stdin: answer "y" to the prompt
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	done := make(chan struct{})
	go func() {
		_, _ = w.Write([]byte("y\n"))
		_ = w.Close()
		close(done)
	}()

	err := runRemove("test-skill", false, false)

	os.Stdin = oldStdin
	<-done

	if err != nil {
		t.Fatalf("runRemove failed: %v", err)
	}

	// Verify skill directory is deleted
	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Error("expected skill directory to be deleted")
	}
}

func TestHandleRemove_ParsesFlags(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	_ = os.MkdirAll(".agents/skills/test-skill", 0755)
	_ = os.WriteFile(".agents/skills/test-skill/SKILL.md", []byte("name: test-skill\n"), 0644)
	_ = os.WriteFile("AGENTS.md", []byte("# Agent Skills\n"), 0644)

	// Test --force flag via handleRemove
	err := handleRemove([]string{"test-skill", "--force"})
	if err != nil {
		t.Fatalf("handleRemove with --force failed: %v", err)
	}

	if _, err := os.Stat(".agents/skills/test-skill"); !os.IsNotExist(err) {
		t.Error("expected skill directory to be deleted")
	}
}

func TestHandleRemove_RequiresName(t *testing.T) {
	err := handleRemove([]string{})
	if err == nil {
		t.Fatal("expected error when no skill name provided")
	}
	if !strings.Contains(err.Error(), "usage") {
		t.Errorf("expected usage message in error, got: %v", err)
	}
}

func TestRunRemove_RegeneratesAgentsMD(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create canonical skill
	_ = os.MkdirAll(".agents/skills/test-skill", 0755)
	_ = os.WriteFile(".agents/skills/test-skill/SKILL.md", []byte("---\ndescription: A test skill\nauto_invoke: true\nscope: root\n---\n"), 0644)

	// Create .opencode/ directory structure (needed for regen tools to write)
	_ = os.MkdirAll(".opencode", 0755)

	// Create AGENTS.md with proper headers and the skill entry
	agentsContent := `# Agent Skills

## Available Skills

| Skill | Description | Location |
| ----- | ----------- | -------- |
| ` + "`test-skill`" + ` | A test skill | [SKILL.md](.agents/skills/test-skill/SKILL.md) |

### Auto-invoke Skills

When performing these actions, ALWAYS invoke the corresponding skill FIRST:

| Action                              | Skill      |
| ----------------------------------- | ---------- |
| test-skill                          | ` + "`test-skill`" + ` |
`
	if err := os.WriteFile("AGENTS.md", []byte(agentsContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := runRemove("test-skill", false, true)
	if err != nil {
		t.Fatalf("runRemove failed: %v", err)
	}

	// Verify AGENTS.md no longer contains test-skill
	agentsAfter, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(agentsAfter), "test-skill") {
		t.Error("AGENTS.md should not contain the deleted skill after regeneration")
	}
}
