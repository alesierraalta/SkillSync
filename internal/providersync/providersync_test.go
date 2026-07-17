package providersync

import (
	"os"
	"path/filepath"
	"testing"
)

func seedSkill(t *testing.T, root, name string, files map[string]string) {
	t.Helper()
	base := filepath.Join(root, ".agents", "skills", name)
	for rel, content := range files {
		p := filepath.Join(base, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMirrorCopiesSkillTree(t *testing.T) {
	root := t.TempDir()
	seedSkill(t, root, "my-skill", map[string]string{
		"SKILL.md":              "# My Skill",
		"references/palette.md": "palette",
		"assets/sync.sh":        "#!/bin/sh",
	})

	n, err := Mirror(root, ".claude")
	if err != nil {
		t.Fatalf("Mirror failed: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 skill mirrored, got %d", n)
	}

	for _, rel := range []string{"SKILL.md", "references/palette.md", "assets/sync.sh"} {
		p := filepath.Join(root, ".claude", "skills", "my-skill", filepath.FromSlash(rel))
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s mirrored to .claude, missing: %v", rel, err)
		}
	}
}

func TestMirrorSkipsNonSkillDirs(t *testing.T) {
	root := t.TempDir()
	seedSkill(t, root, "real-skill", map[string]string{"SKILL.md": "# ok"})
	// A dir without SKILL.md must be ignored.
	if err := os.MkdirAll(filepath.Join(root, ".agents", "skills", "not-a-skill"), 0755); err != nil {
		t.Fatal(err)
	}

	n, err := Mirror(root, ".gemini")
	if err != nil {
		t.Fatalf("Mirror failed: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 skill, got %d", n)
	}
	if _, err := os.Stat(filepath.Join(root, ".gemini", "skills", "not-a-skill")); !os.IsNotExist(err) {
		t.Errorf("non-skill dir should not be mirrored")
	}
}

func TestMirrorMissingSourceIsNoop(t *testing.T) {
	root := t.TempDir()
	n, err := Mirror(root, ".claude")
	if err != nil {
		t.Fatalf("expected no error for missing source, got %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}
