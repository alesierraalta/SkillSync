package ui

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"skillsync/tui/internal/opencode"
	"skillsync/tui/internal/parser"
	"skillsync/tui/internal/syncengine"
	"skillsync/tui/internal/types"
)

// RegenerateAfterDelete runs all post-delete regeneration steps best-effort.
// Each step is independent; errors from one step do not block subsequent steps.
// Returns combined errors from all failed steps.
func RegenerateAfterDelete(root string) error {
	var errs []error

	// Step 1: Discover remaining skills from .agents/skills/
	skills, err := syncengine.DiscoverSkills(root)
	if err != nil {
		errs = append(errs, fmt.Errorf("step 1 discover skills: %w", err))
	}
	if skills == nil {
		skills = []types.Skill{}
	}

	// Step 2: Update AGENTS.md (non-fatal)
	if err := syncengine.UpdateAgentsMarkdown(root, skills, "", false); err != nil {
		errs = append(errs, fmt.Errorf("step 2 update AGENTS.md: %w", err))
	}

	// Step 3: SyncSkills with prune (non-fatal)
	if _, err := opencode.SyncSkills(root, opencode.Options{Prune: true}); err != nil {
		errs = append(errs, fmt.Errorf("step 3 sync skills: %w", err))
	}

	// Step 4: Parse mirrored skills from .opencode/skills/ (inline, no Backend dependency)
	mirroredSkills := parseMirroredSkillsForRegen(root)

	// Step 5: Add base command skills (inline slice)
	allSkills := make([]types.Skill, 0, len(mirroredSkills)+5)
	allSkills = append(allSkills, types.Skill{
		Name: "skill", Metadata: types.Metadata{Description: "Entry point for skill management", AutoInvoke: []string{"skill"}},
	})
	allSkills = append(allSkills, types.Skill{
		Name: "find", Metadata: types.Metadata{Description: "Search and list existing skills", AutoInvoke: []string{"find"}},
	})
	allSkills = append(allSkills, types.Skill{
		Name: "create", Metadata: types.Metadata{Description: "Create a new agent skill from a prompt", AutoInvoke: []string{"create"}},
	})
	allSkills = append(allSkills, types.Skill{
		Name: "sync", Metadata: types.Metadata{Description: "Synchronize skills and update AGENTS.md/OPENCODE.md", AutoInvoke: []string{"sync"}},
	})
	allSkills = append(allSkills, types.Skill{
		Name: "fullskills", Metadata: types.Metadata{Description: "Complete skill workflow", AutoInvoke: []string{"fullskills"}},
	})
	allSkills = append(allSkills, mirroredSkills...)

	// Step 6: Regenerate tools (non-fatal)
	if _, err := opencode.RegenerateTools(root, allSkills, false); err != nil {
		errs = append(errs, fmt.Errorf("step 6 regenerate tools: %w", err))
	}

	// Step 7: Regenerate agent (non-fatal)
	if _, err := opencode.RegenerateAgent(root, allSkills, false); err != nil {
		errs = append(errs, fmt.Errorf("step 7 regenerate agent: %w", err))
	}

	// Step 8: Copy AGENTS.md → OPENCODE.md (non-fatal)
	if _, err := opencode.CopyAgentsMD(root); err != nil {
		errs = append(errs, fmt.Errorf("step 8 copy AGENTS.md: %w", err))
	}

	// Step 9: Regenerate commands (non-fatal)
	if _, err := opencode.RegenerateCommands(root, allSkills, false); err != nil {
		errs = append(errs, fmt.Errorf("step 9 regenerate commands: %w", err))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// parseMirroredSkillsForRegen walks .opencode/skills/ and parses each SKILL.md.
// This is the inline equivalent of Backend.parseMirroredSkillsForEcosystem.
func parseMirroredSkillsForRegen(root string) []types.Skill {
	skillsPath := filepath.Join(root, ".opencode", "skills")
	var skillPaths []string
	err := filepath.WalkDir(skillsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil
		}
		if info.IsDir() {
			skillFile := filepath.Join(path, "SKILL.md")
			if _, err := os.Stat(skillFile); err == nil {
				skillPaths = append(skillPaths, skillFile)
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil
	}

	var skills []types.Skill
	for _, p := range skillPaths {
		skill, err := parser.Parse(p)
		if err != nil {
			continue
		}
		skills = append(skills, *skill)
	}
	return skills
}
