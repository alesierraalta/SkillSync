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
		filepath.Join(tmp, "skill1", "SKILL.md"),
		filepath.Join(tmp, "nested", "skill2", "SKILL.md"),
	}
	
	for _, p := range paths {
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("# Skill"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	
	// Should skip this
	os.MkdirAll(filepath.Join(tmp, "no-skill"), 0755)
	os.WriteFile(filepath.Join(tmp, "no-skill", "file.txt"), []byte("not a skill"), 0644)

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
