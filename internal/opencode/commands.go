package opencode

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skillsync/tui/internal/diff"
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/types"
)

const ManagedMarker = "managed_by: skillsync"

// RegenerateCommands generates OpenCode slash command markdown files
func RegenerateCommands(root string, skills []types.Skill, dryRun bool) (*runner.SyncReport, error) {
	report := &runner.SyncReport{}
	cmdDir := filepath.Join(root, ".opencode", "commands")
	if !dryRun {
		if err := os.MkdirAll(cmdDir, 0755); err != nil {
			return report, fmt.Errorf("mkdir commands: %w", err)
		}
	}

	managedFiles := make(map[string]bool)

	// Base commands
	baseCommands := []struct {
		name string
		desc string
	}{
		{"skill", "Skill management"},
		{"find", "Find skills in the registry"},
		{"create", "Create a new skill from a find result"},
		{"sync", "Synchronize all skills with OpenCode"},
		{"fullskills", "Full skill lifecycle flow (find -> create -> sync)"},
	}

	for _, bc := range baseCommands {
		filename := bc.name + ".md"
		content := fmt.Sprintf("---\nmanaged_by: skillsync\ndescription: %s\n---\n\n/synck %s \"$!1\"\n", bc.desc, bc.name)
		managedFiles[filename] = true
		if change, err := writeIfManaged(filepath.Join(cmdDir, filename), content, dryRun, root); err != nil {
			return report, err
		} else if change != nil {
			report.Changes = append(report.Changes, *change)
		}
	}

	// Skill-specific commands
	for _, s := range skills {
		if len(s.Metadata.AutoInvoke) == 0 {
			continue
		}
		filename := s.Name + ".md"
		content := fmt.Sprintf("---\nmanaged_by: skillsync\ndescription: %s\n---\n\n/synck %s \"$!1\"\n", s.Metadata.Description, s.Name)
		managedFiles[filename] = true
		if change, err := writeIfManaged(filepath.Join(cmdDir, filename), content, dryRun, root); err != nil {
			return report, err
		} else if change != nil {
			report.Changes = append(report.Changes, *change)
		}
	}

	// Prune orphans
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		if os.IsNotExist(err) {
			return report, nil
		}
		return report, fmt.Errorf("read commands dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		if managedFiles[filename] {
			continue
		}

		path := filepath.Join(cmdDir, filename)
		isManaged, err := isManagedBySkillSync(path)
		if err != nil {
			continue // Skip files we can't read
		}

		if isManaged {
			relPath, _ := filepath.Rel(root, path)
			before := ""
			if content, err := os.ReadFile(path); err == nil {
				before = string(content)
			}
			if dryRun {
				report.Changes = append(report.Changes, runner.FileChange{
					Path:   relPath,
					Status: "deleted",
					Before: before,
					After:  "",
				})
				fmt.Printf("[dry-run] would prune orphan command: %s\n", filename)
			} else {
				if err := os.Remove(path); err != nil {
					return report, fmt.Errorf("remove orphan %s: %w", filename, err)
				}
				diffStr, summary := diff.UnifiedDiff(before, "", 50)
				report.Changes = append(report.Changes, runner.FileChange{
					Path:    relPath,
					Status:  "deleted",
					Before:  before,
					After:   "",
					Diff:    diffStr,
					Summary: summary,
				})
			}
		}
	}

	return report, nil
}

func writeIfManaged(path string, content string, dryRun bool, root string) (*runner.FileChange, error) {
	relPath, _ := filepath.Rel(root, path)

	if _, err := os.Stat(path); err == nil {
		managed, err := isManagedBySkillSync(path)
		if err != nil {
			return nil, err
		}
		if !managed {
			return nil, nil // Don't touch user files
		}

		// Read current content to avoid unnecessary writes
		current, _ := os.ReadFile(path)
		if string(current) == content {
			return nil, nil
		}

		before := string(current)
		if dryRun {
			diffStr, summary := diff.UnifiedDiff(before, content, 50)
			return &runner.FileChange{
				Path:    relPath,
				Status:  "modified",
				Before:  before,
				After:   content,
				Diff:    diffStr,
				Summary: summary,
			}, nil
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return nil, err
		}
		diffStr, summary := diff.UnifiedDiff(before, content, 50)
		return &runner.FileChange{
			Path:    relPath,
			Status:  "modified",
			Before:  before,
			After:   content,
			Diff:    diffStr,
			Summary: summary,
		}, nil
	}

	// File doesn't exist → created
	if dryRun {
		diffStr, summary := diff.UnifiedDiff("", content, 50)
		return &runner.FileChange{
			Path:    relPath,
			Status:  "created",
			Before:  "",
			After:   content,
			Diff:    diffStr,
			Summary: summary,
		}, nil
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, err
	}
	diffStr, summary := diff.UnifiedDiff("", content, 50)
	return &runner.FileChange{
		Path:    relPath,
		Status:  "created",
		Before:  "",
		After:   content,
		Diff:    diffStr,
		Summary: summary,
	}, nil
}

func isManagedBySkillSync(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(content), ManagedMarker), nil
}
