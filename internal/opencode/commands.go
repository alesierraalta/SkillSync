package opencode

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"skillsync/tui/internal/types"
)

const ManagedMarker = "managed_by: skillsync"

// RegenerateCommands generates OpenCode slash command markdown files
func RegenerateCommands(root string, skills []types.Skill, dryRun bool) error {
	cmdDir := filepath.Join(root, ".opencode", "commands")
	if !dryRun {
		if err := os.MkdirAll(cmdDir, 0755); err != nil {
			return fmt.Errorf("mkdir commands: %w", err)
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
		if err := writeIfManaged(filepath.Join(cmdDir, filename), content, dryRun); err != nil {
			return err
		}
	}

	// Skill-specific commands
	for _, s := range skills {
		if !s.Metadata.AutoInvoke {
			continue
		}
		filename := s.Name + ".md"
		content := fmt.Sprintf("---\nmanaged_by: skillsync\ndescription: %s\n---\n\n/synck %s \"$!1\"\n", s.Metadata.Description, s.Name)
		managedFiles[filename] = true
		if err := writeIfManaged(filepath.Join(cmdDir, filename), content, dryRun); err != nil {
			return err
		}
	}

	// Prune orphans
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read commands dir: %w", err)
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
			if dryRun {
				fmt.Printf("[dry-run] would prune orphan command: %s\n", filename)
			} else {
				if err := os.Remove(path); err != nil {
					return fmt.Errorf("remove orphan %s: %w", filename, err)
				}
			}
		}
	}

	return nil
}

func writeIfManaged(path string, content string, dryRun bool) error {
	if _, err := os.Stat(path); err == nil {
		managed, err := isManagedBySkillSync(path)
		if err != nil {
			return err
		}
		if !managed {
			return nil // Don't touch user files
		}

		// Read current content to avoid unnecessary writes
		current, _ := os.ReadFile(path)
		if string(current) == content {
			return nil
		}
	}

	if dryRun {
		fmt.Printf("[dry-run] would write command: %s\n", path)
		return nil
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func isManagedBySkillSync(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(content), ManagedMarker), nil
}
