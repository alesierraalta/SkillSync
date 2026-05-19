package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"skillsync/tui/internal/types"
)

func TestStorage(t *testing.T) {
	tmpDir := t.TempDir()
	
	s := &Service{
		RootPath: tmpDir,
	}

	skill := &types.Skill{
		Name:    "test-skill",
		RawBody: "full skill content with --- and metadata",
	}
	
	metadata := StoredMetadata{
		SkillName:     skill.Name,
		OriginProject: "/abs/path/project",
		OriginPath:    ".agents/skills/test-skill",
		SavedAt:       time.Now(),
	}

	t.Run("Save", func(t *testing.T) {
		err := s.Save(skill, metadata)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Verify files exist
		skillPath := filepath.Join(tmpDir, "test-skill", "SKILL.md")
		metaPath := filepath.Join(tmpDir, "test-skill", "METADATA.json")

		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			t.Errorf("SKILL.md not found")
		}
		if _, err := os.Stat(metaPath); os.IsNotExist(err) {
			t.Errorf("METADATA.json not found")
		}

		content, _ := os.ReadFile(skillPath)
		if !strings.Contains(string(content), skill.RawBody) {
			t.Errorf("Expected body %q to be in %q", skill.RawBody, string(content))
		}
	})

	t.Run("List", func(t *testing.T) {
		skills, err := s.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(skills) != 1 {
			t.Errorf("Expected 1 skill, got %d", len(skills))
		}
		if skills[0].Metadata.SkillName != "test-skill" {
			t.Errorf("Expected skill name test-skill, got %s", skills[0].Metadata.SkillName)
		}
	})

	t.Run("OverwriteAndBytePreservation", func(t *testing.T) {
		// New content with specific bytes
		originalContent := "line 1\nline 2\r\nline 3\tTabbed"
		skill := &types.Skill{
			Name:    "test-skill",
			RawBody: originalContent,
		}
		
		metadata := StoredMetadata{
			SkillName: "test-skill",
			SavedAt:   time.Now(),
		}

		// Save again (overwrite)
		err := s.Save(skill, metadata)
		if err != nil {
			t.Fatalf("Overwrite Save failed: %v", err)
		}

		skillPath := filepath.Join(tmpDir, "test-skill", "SKILL.md")
		content, _ := os.ReadFile(skillPath)
		
		if !strings.Contains(string(content), originalContent) {
			t.Errorf("Byte-for-byte failure. Expected body %q to be in %q", originalContent, string(content))
		}

		// Verify List still shows only 1 entry (overwritten, not duplicated)
		skills, _ := s.List()
		if len(skills) != 1 {
			t.Errorf("Expected 1 skill after overwrite, got %d", len(skills))
		}
	})
}

func TestSavePersistsFullFormattedContent(t *testing.T) {
	tmpDir := t.TempDir()
	s := &Service{RootPath: tmpDir}

	skill := &types.Skill{
		Name: "formatted-skill",
		Metadata: types.Metadata{
			Scope: "global",
		},
		RawBody: "# Body",
	}

	metadata := StoredMetadata{
		SkillName: skill.Name,
		SavedAt:   time.Now(),
	}

	err := s.Save(skill, metadata)
	if err != nil {
		t.Fatal(err)
	}

	skillPath := filepath.Join(tmpDir, "formatted-skill", "SKILL.md")
	raw, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(raw)
	if !strings.Contains(content, "---") {
		t.Error("Frontmatter delimiters '---' missing in saved file")
	}
	if !strings.Contains(content, "scope: global") {
		t.Error("Metadata 'scope: global' missing in saved file")
	}
	if !strings.Contains(content, "# Body") {
		t.Error("Body missing in saved file")
	}
}


func TestNewServiceIsolation(t *testing.T) {
	t.Run("prioritizes SKILLSYNC_HOME", func(t *testing.T) {
		tmpHome := t.TempDir()
		t.Setenv("SKILLSYNC_HOME", tmpHome)

		s := NewService("")
		expected := filepath.Join(tmpHome, "storage")
		if s.RootPath != expected {
			t.Errorf("expected RootPath %q, got %q", expected, s.RootPath)
		}
	})

	t.Run("fallbacks to user home if env not set", func(t *testing.T) {
		t.Setenv("SKILLSYNC_HOME", "")
		// We can't easily mock os.UserHomeDir without more intrusive changes,
		// but we can check it's not the empty root or something weird.
		s := NewService("")
		if s.RootPath == "" {
			t.Error("expected non-empty RootPath")
		}
		if !strings.Contains(s.RootPath, ".skillsync") {
			t.Errorf("expected RootPath to contain .skillsync, got %q", s.RootPath)
		}
	})
}
