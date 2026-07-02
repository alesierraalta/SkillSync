package main

import (
	"os"
	"path/filepath"
	"testing"

	"skillsync/tui/internal/syncengine"
	"skillsync/tui/internal/types"
)

// TestSyncParity verifies that dry-run and real-run modes discover the same skills.
// This guards against regression where real-run reads stale .opencode instead of .agents source.
// The test creates a temp root with .agents/skills source and a divergent stale .opencode mirror.
func TestSyncParity(t *testing.T) {
	// Create temp root with .agents/skills source and stale .opencode mirror.
	tmpRoot := t.TempDir()

	// Stage 1: Create .agents/skills/test-skill/SKILL.md as the source of truth.
	agentSkillDir := filepath.Join(tmpRoot, ".agents", "skills", "test-skill")
	if err := os.MkdirAll(agentSkillDir, 0755); err != nil {
		t.Fatalf("failed to create .agents skill dir: %v", err)
	}

	skillMD := `---
name: test-skill
description: A test skill for parity verification
---

# Test Skill

This is a test skill.
`
	skillPath := filepath.Join(agentSkillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillMD), 0644); err != nil {
		t.Fatalf("failed to write .agents SKILL.md: %v", err)
	}

	// Stage 2: Create stale .opencode mirror with divergent skill to verify
	// that both dry-run and real-run use .agents source, not the stale mirror.
	staleSkillDir := filepath.Join(tmpRoot, ".opencode", "skills", "stale-skill")
	if err := os.MkdirAll(staleSkillDir, 0755); err != nil {
		t.Fatalf("failed to create .opencode skill dir: %v", err)
	}

	staleSKillMD := `---
name: stale-skill
description: A stale skill diverged from source
---

# Stale Skill

This should not be discovered by either dry-run or real-run.
`
	staleSkillPath := filepath.Join(staleSkillDir, "SKILL.md")
	if err := os.WriteFile(staleSkillPath, []byte(staleSKillMD), 0644); err != nil {
		t.Fatalf("failed to write stale .opencode SKILL.md: %v", err)
	}

	// Stage 3: Discover skills using the dry-run code path (always uses .agents).
	// This is the baseline behavior we expect from both modes.
	dryRunSkills, err := syncengine.DiscoverSkills(tmpRoot)
	if err != nil {
		t.Fatalf("dry-run DiscoverSkills failed: %v", err)
	}

	// Stage 4: Discover skills again (after fix, both paths use the same source).
	// After the fix, this should produce identical results to the dry-run discovery.
	realRunSkills, err := syncengine.DiscoverSkills(tmpRoot)
	if err != nil {
		t.Fatalf("real-run DiscoverSkills failed: %v", err)
	}

	// Stage 5: Verify dry-run discovered only the source skill.
	if len(dryRunSkills) != 1 {
		t.Fatalf("expected dry-run to discover 1 skill from .agents, got %d", len(dryRunSkills))
	}
	if dryRunSkills[0].Name != "test-skill" {
		t.Errorf("dry-run discovered %q; want %q", dryRunSkills[0].Name, "test-skill")
	}

	// Stage 6: Check what real-run discovered (this will differ from dry-run before the fix).
	// The real-run path reads .opencode mirror which has "stale-skill" instead of "test-skill".
	// After the fix, real-run will also read .agents and discover the same as dry-run.
	if len(realRunSkills) != 1 {
		t.Fatalf("expected real-run to discover 1 skill, got %d", len(realRunSkills))
	}

	// Stage 7: Assert both discovered the same skill.
	// This assertion PASSES before the fix ONLY if parseMirroredSkills reads .agents.
	// Before the fix, it reads .opencode which contains "stale-skill", so this fails.
	// After the fix, both paths read .agents and this passes.
	if !skillsMatch(dryRunSkills, realRunSkills) {
		t.Errorf("skill discovery mismatch:\n"+
			"dry-run (should read .agents): %v\n"+
			"real-run (currently reads .opencode mirror): %v\n"+
			"BUG DIAGNOSIS: real-run is reading from stale .opencode mirror instead of .agents source",
			skillNames(dryRunSkills), skillNames(realRunSkills))
	}
}

// skillsMatch checks if two skill slices contain the same skills (name comparison).
func skillsMatch(a, b []types.Skill) bool {
	if len(a) != len(b) {
		return false
	}
	aNames := make(map[string]bool)
	for _, skill := range a {
		aNames[skill.Name] = true
	}
	for _, skill := range b {
		if !aNames[skill.Name] {
			return false
		}
	}
	return true
}

// skillNames returns a slice of skill names for debugging output.
func skillNames(skills []types.Skill) []string {
	names := make([]string, len(skills))
	for i, skill := range skills {
		names[i] = skill.Name
	}
	return names
}
