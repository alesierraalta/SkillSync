package syncengine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skillsync/tui/internal/diff"
	"skillsync/tui/internal/parser"
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/types"
)

type SyncOptions struct {
	DryRun     bool
	ProgressCb func(stage string, done, total int)
}

func Sync(root string, opts SyncOptions) (*runner.SyncReport, error) {
	report := &runner.SyncReport{}

	skills, err := DiscoverSkills(root)
	if err != nil {
		return report, fmt.Errorf("discover skills: %w", err)
	}
	if opts.ProgressCb != nil {
		opts.ProgressCb("Discovering skills", 1, 8)
	}

	agentsPath := filepath.Join(root, "AGENTS.md")
	before := ""
	if content, err := os.ReadFile(agentsPath); err == nil {
		before = string(content)
	}

	if err := UpdateAgentsMarkdown(root, skills, opts.DryRun); err != nil {
		return report, fmt.Errorf("update AGENTS.md: %w", err)
	}

	if !opts.DryRun {
		after := ""
		if content, err := os.ReadFile(agentsPath); err == nil {
			after = string(content)
		}
		if before != after {
			diffStr, summary := diff.UnifiedDiff(before, after, 50)
			status := "modified"
			if before == "" {
				status = "created"
			}
			report.Changes = append(report.Changes, runner.FileChange{
				Path:    "AGENTS.md",
				Status:  status,
				Before:  before,
				After:   after,
				Diff:    diffStr,
				Summary: summary,
			})
		}
	}

	if opts.ProgressCb != nil {
		opts.ProgressCb("Updating AGENTS.md", 2, 8)
	}

	return report, nil
}

func DiscoverSkills(root string) ([]types.Skill, error) {
	skillsDir := filepath.Join(root, ".agents", "skills")
	var skills []types.Skill

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return skills, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		skillPath := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillPath); err == nil {
			skill, err := parser.Parse(skillPath)
			if err != nil {
				continue // skip invalid skills
			}
			skills = append(skills, *skill)
		}
	}
	return skills, nil
}

func AggregateMetadata(skills []types.Skill, targetScope string) []types.Skill {
	var matched []types.Skill
	for _, s := range skills {
		if s.Metadata.Scope == targetScope {
			matched = append(matched, s)
			break
		}
	}
	return matched
}

func UpdateAgentsMarkdown(root string, skills []types.Skill, dryRun bool) error {
	agentsPath := filepath.Join(root, "AGENTS.md")
	
	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		return nil
	}
	
	content, err := os.ReadFile(agentsPath)
	if err != nil {
		return err
	}
	
	lines := strings.Split(string(content), "\n")
	var out []string
	inSection := false
	sectionReplaced := false
	
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "### Auto-invoke Skills") {
			inSection = true
			out = append(out, line)
			out = append(out, "")
			out = append(out, "When performing these actions, ALWAYS invoke the corresponding skill FIRST:")
			out = append(out, "")
			out = append(out, "| Action                              | Skill      |")
			out = append(out, "| ----------------------------------- | ---------- |")
			
			rootSkills := AggregateMetadata(skills, "root")
			for _, s := range rootSkills {
				if s.Metadata.AutoInvoke {
					action := s.Name
					paddedAction := fmt.Sprintf("%-35s", action)
					out = append(out, fmt.Sprintf("| %s | `%s` |", paddedAction, s.Name))
				}
			}
			
			sectionReplaced = true
			continue
		}
		
		if inSection {
			if strings.HasPrefix(line, "### ") {
				inSection = false
				out = append(out, "")
				out = append(out, line)
			}
			continue
		}
		
		out = append(out, line)
	}
	
	if sectionReplaced && !dryRun {
		return os.WriteFile(agentsPath, []byte(strings.Join(out, "\n")), 0644)
	}
	
	return nil
}
