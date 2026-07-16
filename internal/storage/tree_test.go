package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"skillsync/tui/internal/types"
)

// writeTree creates a skill source dir with SKILL.md plus reference files.
func writeTree(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"SKILL.md":                 "---\nname: tree-skill\ndescription: test\n---\n# Tree Skill",
		"references/palette.md":    "palette",
		"assets/sync.sh":           "#!/bin/sh",
		"references/deep/notes.md": "notes",
	}
	for rel, content := range files {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestSaveCopiesSkillTree(t *testing.T) {
	src := writeTree(t)
	s := NewService(filepath.Join(t.TempDir(), "storage"))

	skill := &types.Skill{
		Name: "tree-skill",
		Path: filepath.Join(src, "SKILL.md"),
	}
	meta := StoredMetadata{SkillName: "tree-skill", SavedAt: time.Now()}

	if err := s.Save(skill, meta); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	for _, rel := range []string{
		"SKILL.md",
		"METADATA.json",
		"references/palette.md",
		"assets/sync.sh",
		"references/deep/notes.md",
	} {
		p := filepath.Join(s.RootPath, "tree-skill", filepath.FromSlash(rel))
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s in storage, missing: %v", rel, err)
		}
	}
}

func TestSaveWithoutPathStillWorks(t *testing.T) {
	s := NewService(filepath.Join(t.TempDir(), "storage"))
	skill := &types.Skill{Name: "no-path"}
	if err := s.Save(skill, StoredMetadata{SkillName: "no-path"}); err != nil {
		t.Fatalf("Save without Path failed: %v", err)
	}
}

func TestCopyExtras(t *testing.T) {
	src := writeTree(t)
	s := NewService(filepath.Join(t.TempDir(), "storage"))
	skill := &types.Skill{Name: "tree-skill", Path: filepath.Join(src, "SKILL.md")}
	if err := s.Save(skill, StoredMetadata{SkillName: "tree-skill"}); err != nil {
		t.Fatal(err)
	}

	dst := t.TempDir()
	if err := s.CopyExtras("tree-skill", dst); err != nil {
		t.Fatalf("CopyExtras failed: %v", err)
	}

	// Extras copied
	for _, rel := range []string{"references/palette.md", "assets/sync.sh", "references/deep/notes.md"} {
		if _, err := os.Stat(filepath.Join(dst, filepath.FromSlash(rel))); err != nil {
			t.Errorf("expected extra %s, missing: %v", rel, err)
		}
	}
	// SKILL.md and METADATA.json are NOT extras
	for _, rel := range []string{"SKILL.md", "METADATA.json"} {
		if _, err := os.Stat(filepath.Join(dst, rel)); err == nil {
			t.Errorf("%s should not be copied by CopyExtras", rel)
		}
	}
}
