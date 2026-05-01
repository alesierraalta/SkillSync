package opencode

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"skillsync/tui/internal/types"
)

func TestRegenerateCommands_CreatesBaseFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cmdDir := filepath.Join(tmpDir, ".opencode", "commands")
	err := os.MkdirAll(cmdDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	skills := []types.Skill{}
	err = RegenerateCommands(tmpDir, skills, false)
	if err != nil {
		t.Fatalf("RegenerateCommands failed: %v", err)
	}

	baseFiles := []string{"skill.md", "find.md", "create.md", "sync.md", "fullskills.md"}
	for _, f := range baseFiles {
		path := filepath.Join(cmdDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
			continue
		}

		content, _ := os.ReadFile(path)
		if !strings.Contains(string(content), "managed_by: skillsync") {
			t.Errorf("file %s missing managed_by marker", f)
		}
	}
}

func TestRegenerateCommands_PreservesUserFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cmdDir := filepath.Join(tmpDir, ".opencode", "commands")
	err := os.MkdirAll(cmdDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	userCmdPath := filepath.Join(cmdDir, "my-custom-cmd.md")
	userContent := "--- title: My Cmd ---\nCustom logic"
	os.WriteFile(userCmdPath, []byte(userContent), 0644)

	skills := []types.Skill{}
	err = RegenerateCommands(tmpDir, skills, false)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(userCmdPath)
	if err != nil {
		t.Errorf("user file deleted")
	} else if string(content) != userContent {
		t.Errorf("user file modified")
	}
}

func TestRegenerateCommands_PrunesOrphans(t *testing.T) {
	tmpDir := t.TempDir()
	cmdDir := filepath.Join(tmpDir, ".opencode", "commands")
	err := os.MkdirAll(cmdDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	orphanPath := filepath.Join(cmdDir, "old-skill.md")
	os.WriteFile(orphanPath, []byte("---\nmanaged_by: skillsync\n---\n"), 0644)

	skills := []types.Skill{} // No skills, so old-skill.md should be pruned
	err = RegenerateCommands(tmpDir, skills, false)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(orphanPath); !os.IsNotExist(err) {
		t.Errorf("orphan file not pruned")
	}
}

func TestRegenerateCommands_SkillSpecific(t *testing.T) {
	tmpDir := t.TempDir()
	cmdDir := filepath.Join(tmpDir, ".opencode", "commands")
	err := os.MkdirAll(cmdDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	skills := []types.Skill{
		{
			Name: "deploy",
			Metadata: types.Metadata{
				Description: "Deploy to production",
				AutoInvoke:  true,
			},
		},
	}
	err = RegenerateCommands(tmpDir, skills, false)
	if err != nil {
		t.Fatal(err)
	}

	deployPath := filepath.Join(cmdDir, "deploy.md")
	if _, err := os.Stat(deployPath); os.IsNotExist(err) {
		t.Errorf("skill command not generated")
	} else {
		content, _ := os.ReadFile(deployPath)
		if !strings.Contains(string(content), "managed_by: skillsync") {
			t.Errorf("skill command missing marker")
		}
		if !strings.Contains(string(content), "synck deploy") {
			t.Errorf("skill command missing execution logic")
		}
	}
}

func TestRegenerateCommands_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	cmdDir := filepath.Join(tmpDir, ".opencode", "commands")
	os.MkdirAll(cmdDir, 0755)

	skills := []types.Skill{{Name: "test", Metadata: types.Metadata{AutoInvoke: true}}}
	
	// First run
	err := RegenerateCommands(tmpDir, skills, false)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(cmdDir, "test.md")
	stat1, _ := os.Stat(path)
	
	// Second run
	err = RegenerateCommands(tmpDir, skills, false)
	if err != nil {
		t.Fatal(err)
	}

	stat2, _ := os.Stat(path)
	if stat1.ModTime() != stat2.ModTime() {
		t.Errorf("expected idempotent sync to skip writing if content is same")
	}
}
