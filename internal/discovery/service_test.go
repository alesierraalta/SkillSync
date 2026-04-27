package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverSkills(t *testing.T) {
	// Setup temp dir
	tmp := t.TempDir()
	
	// Create structure
	// tmp/
	//   skill1/SKILL.md
	//   nested/skill2/SKILL.md
	//   no-skill/file.txt
	
	paths := []string{
		filepath.Join(tmp, ".agents", "skills", "skill1", "SKILL.md"),
		filepath.Join(tmp, ".opencode", "nested", "skill2", "SKILL.md"),
	}
	
	for _, p := range paths {
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("# Skill"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	
	// Should skip this (not a known provider)
	ignorePath := filepath.Join(tmp, ".git", "hooks", "SKILL.md")
	os.MkdirAll(filepath.Dir(ignorePath), 0755)
	os.WriteFile(ignorePath, []byte("# Ignored Skill"), 0644)

	// Should skip this (global home directory simulation)
	globalTmp := t.TempDir()
	globalPath := filepath.Join(globalTmp, ".agents", "skills", "skill3", "SKILL.md")
	os.MkdirAll(filepath.Dir(globalPath), 0755)
	os.WriteFile(globalPath, []byte("# Global Skill"), 0644)

	found, err := DiscoverSkills(tmp)
	if err != nil {
		t.Fatalf("DiscoverSkills failed: %v", err)
	}

	if len(found) != 2 {
		t.Errorf("Expected 2 skills, found %d", len(found))
	}

	// Verify paths
	count := 0
	for _, f := range found {
		for _, p := range paths {
			if f == p {
				count++
			}
		}
	}
	
	if count != 2 {
		t.Errorf("Expected to find all created skill paths, matches: %d", count)
	}
}

func TestDiscoverSkillsNestedDuplicates(t *testing.T) {
	tmp := t.TempDir()

	// Create a legitimate skill
	skillPath := filepath.Join(tmp, ".agents", "skills", "my-skill", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(skillPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skillPath, []byte("# My Skill"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a nested duplicate SKILL.md that should be ignored
	nestedPath := filepath.Join(tmp, ".agents", "skills", "my-skill", "subfolder", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(nestedPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nestedPath, []byte("# Nested Skill"), 0644); err != nil {
		t.Fatal(err)
	}

	found, err := DiscoverSkills(tmp)
	if err != nil {
		t.Fatalf("DiscoverSkills failed: %v", err)
	}

	// We expect only 1 skill (the top-level one)
	if len(found) != 1 {
		t.Errorf("Expected 1 skill, found %d: %v", len(found), found)
	}

	if len(found) > 0 && found[0] != skillPath {
		t.Errorf("Expected skill path %s, got %s", skillPath, found[0])
	}
}

func TestDiscoverSkillsNestedOnly(t *testing.T) {
	tmp := t.TempDir()

	// Create a folder but NO SKILL.md at the first level
	skillDir := filepath.Join(tmp, ".agents", "skills", "my-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a nested SKILL.md
	nestedPath := filepath.Join(skillDir, "nested-folder", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(nestedPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nestedPath, []byte("# Nested Skill"), 0644); err != nil {
		t.Fatal(err)
	}

	found, err := DiscoverSkills(tmp)
	if err != nil {
		t.Fatalf("DiscoverSkills failed: %v", err)
	}

	// We expect 1 skill (the nested one, since top-level didn't have one)
	if len(found) != 1 {
		t.Errorf("Expected 1 skill, found %d: %v", len(found), found)
	}

	if len(found) > 0 && found[0] != nestedPath {
		t.Errorf("Expected skill path %s, got %s", nestedPath, found[0])
	}
}
