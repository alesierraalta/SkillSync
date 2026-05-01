package discovery

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestDiscoverSkillsMultiProvider(t *testing.T) {
	// Setup temp dir
	tmp := t.TempDir()

	// Create skills in .claude, .qwen, and .agents to prove all are discovered
	paths := []string{
		filepath.Join(tmp, ".claude", "skills", "claude-skill", "SKILL.md"),
		filepath.Join(tmp, ".qwen", "skills", "qwen-skill", "SKILL.md"),
		filepath.Join(tmp, ".agents", "skills", "agents-skill", "SKILL.md"),
		filepath.Join(tmp, ".opencode", "skills", "opencode-skill", "SKILL.md"),
		filepath.Join(tmp, ".gemini", "skills", "gemini-skill", "SKILL.md"),
	}

	for _, p := range paths {
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("# Skill"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	found, err := DiscoverSkills(tmp)
	if err != nil {
		t.Fatalf("DiscoverSkills failed: %v", err)
	}

	if len(found) != 5 {
		t.Errorf("Expected 5 skills (.claude, .qwen, .agents, .opencode, .gemini), found %d: %v", len(found), found)
	}

	// Verify each provider was discovered
	providerFound := make(map[string]bool)
	for _, f := range found {
		// Extract provider from path - normalize to forward slashes for cross-platform parsing
		rel, _ := filepath.Rel(tmp, f)
		rel = strings.ReplaceAll(rel, string(filepath.Separator), "/")
		parts := strings.Split(rel, "/")
		for i := 0; i < len(parts); i++ {
			if strings.HasPrefix(parts[i], ".") && parts[i] != "." && parts[i] != ".." {
				providerFound[parts[i]] = true
				break
			}
		}
	}

	expectedProviders := []string{".claude", ".qwen", ".agents", ".opencode", ".gemini"}
	for _, ep := range expectedProviders {
		if !providerFound[ep] {
			t.Errorf("Provider %s was not discovered", ep)
		}
	}
}

func TestDiscoverSkillsDoesNotRegressAgents(t *testing.T) {
	// Explicit regression test: .agents must still work as before
	tmp := t.TempDir()

	agentsPath := filepath.Join(tmp, ".agents", "skills", "test-skill", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(agentsPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agentsPath, []byte("# Test Skill"), 0644); err != nil {
		t.Fatal(err)
	}

	found, err := DiscoverSkills(tmp)
	if err != nil {
		t.Fatalf("DiscoverSkills failed: %v", err)
	}

	if len(found) != 1 {
		t.Errorf("Expected 1 skill from .agents, found %d", len(found))
	}

	if len(found) > 0 && found[0] != agentsPath {
		t.Errorf("Expected .agents skill path %s, got %s", agentsPath, found[0])
	}
}

func TestDiscoverSkills_WithSymlinks(t *testing.T) {
	// Windows junctions / Unix symlinks are not followed by filepath.WalkDir.
	// DiscoverSkills must handle them by checking stat info, not just DirEntry.IsDir().
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer os.Chdir(origDir)

	// Create a real skill directory
	realSkill := filepath.Join(tmp, ".agents", "skills", "real-skill", "SKILL.md")
	_ = os.MkdirAll(filepath.Dir(realSkill), 0755)
	_ = os.WriteFile(realSkill, []byte("name: real-skill\n"), 0644)

	// Create a junction/symlink pointing to the real skill directory.
	linkTarget := filepath.Join(tmp, "link-target", "skills", "linked-skill")
	linkPath := filepath.Join(tmp, ".claude", "skills", "linked-skill")
	_ = os.MkdirAll(linkTarget, 0755)
	_ = os.WriteFile(filepath.Join(linkTarget, "SKILL.md"), []byte("name: linked-skill\n"), 0644)

	// Create symlink on Unix, junction on Windows
	var linkErr error
	if runtime.GOOS == "windows" {
		linkErr = os.Symlink(linkTarget, linkPath)
	} else {
		linkErr = os.Symlink(linkTarget, linkPath)
	}
	if linkErr != nil {
		t.Skipf("symlink creation failed (may need admin on Windows): %v", linkErr)
	}

	// Also test a real directory alongside the symlink
	claudeReal := filepath.Join(tmp, ".claude", "skills", "real-claude-skill", "SKILL.md")
	_ = os.MkdirAll(filepath.Dir(claudeReal), 0755)
	_ = os.WriteFile(claudeReal, []byte("name: real-claude-skill\n"), 0644)

	found, err := DiscoverSkills(tmp)
	if err != nil {
		t.Fatalf("DiscoverSkills failed: %v", err)
	}

	// Collect found provider skill paths
	foundMap := make(map[string]bool)
	for _, f := range found {
		rel, _ := filepath.Rel(tmp, f)
		rel = strings.ReplaceAll(rel, string(filepath.Separator), "/")
		foundMap[rel] = true
	}

	// Real directories must be found
	if !foundMap[".agents/skills/real-skill/SKILL.md"] {
		t.Errorf("real-skill not found in .agents")
	}
	if !foundMap[".claude/skills/real-claude-skill/SKILL.md"] {
		t.Errorf("real-claude-skill not found in .claude")
	}

	// Symlink/junction target must be found via .claude
	if !foundMap[".claude/skills/linked-skill/SKILL.md"] {
		t.Errorf("linked-skill (symlink) not found — DiscoverSkills must follow symlinks/junctions")
	}
}
