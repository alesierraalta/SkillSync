package syncengine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"skillsync/tui/internal/diff"
	"skillsync/tui/internal/parser"
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/types"
)

type SyncOptions struct {
	DryRun     bool
	ProgressCb func(stage string, done, total int)
	Storage    *storage.Service
	Scope      string
}

func Sync(root string, opts SyncOptions) (*runner.SyncReport, error) {
	report := &runner.SyncReport{}

	if info, err := os.Stat(root); err != nil {
		return report, err
	} else if !info.IsDir() {
		return report, fmt.Errorf("root is not a directory: %s", root)
	}

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

	after, updated, err := renderAgentsMarkdown(root, skills, opts.Scope)
	if err != nil {
		return report, fmt.Errorf("update AGENTS.md: %w", err)
	}
	if updated {
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
		if !opts.DryRun {
			if err := os.WriteFile(agentsPath, []byte(after), 0644); err != nil {
				return report, fmt.Errorf("write AGENTS.md: %w", err)
			}
		}
	}

	if opts.ProgressCb != nil {
		opts.ProgressCb("Updating AGENTS.md", 2, 8)
	}

	if opts.Storage != nil && !opts.DryRun {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			absRoot = root
		}
		if err := opts.Storage.RegisterProject(absRoot); err != nil {
			return report, fmt.Errorf("failed to register project: %w", err)
		}
	}

	cleanupChanges, err := cleanupLegacyHarnessArtifacts(root, opts.DryRun)
	if err != nil {
		return report, fmt.Errorf("cleanup legacy harness artifacts: %w", err)
	}
	if len(cleanupChanges) > 0 {
		report.Changes = append(report.Changes, cleanupChanges...)
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
		if scopeMatches(s.Metadata.Scope, targetScope) {
			matched = append(matched, s)
		}
	}
	return matched
}

func scopeMatches(scope, targetScope string) bool {
	if targetScope == "" {
		targetScope = "root"
	}

	for _, part := range strings.Split(scope, ",") {
		if strings.TrimSpace(part) == targetScope {
			return true
		}
	}
	return false
}

func sanitizeMarkdownCell(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	return strings.ReplaceAll(value, "|", `\|`)
}

func UpdateAgentsMarkdown(root string, skills []types.Skill, scope string, dryRun bool) error {
	content, updated, err := renderAgentsMarkdown(root, skills, scope)
	if err != nil || !updated || dryRun {
		return err
	}

	agentsPath := filepath.Join(root, "AGENTS.md")
	return os.WriteFile(agentsPath, []byte(content), 0644)
}

func renderAgentsMarkdown(root string, skills []types.Skill, scope string) (string, bool, error) {
	agentsPath := filepath.Join(root, "AGENTS.md")

	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		return "", false, nil
	}

	content, err := os.ReadFile(agentsPath)
	if err != nil {
		return "", false, err
	}

	lines := strings.Split(string(content), "\n")
	hasAvailable := false
	hasAutoInvoke := false
	for _, line := range lines {
		if strings.HasPrefix(line, "## Available Skills") {
			hasAvailable = true
		}
		if strings.HasPrefix(line, "### Auto-invoke Skills") {
			hasAutoInvoke = true
		}
	}
	if !hasAvailable || !hasAutoInvoke {
		return "", false, fmt.Errorf("required headers missing in AGENTS.md: need '## Available Skills' and '### Auto-invoke Skills'")
	}

	var out []string
	inSection := false
	inAvailableSection := false
	sectionReplaced := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "## Available Skills") {
			inAvailableSection = true
			out = append(out, line)
			out = append(out, "")
			out = append(out, "| Skill | Description | Location |")
			out = append(out, "| ----- | ----------- | -------- |")

			for _, s := range skills {
				rel, _ := filepath.Rel(root, s.Path)
				desc := sanitizeMarkdownCell(s.Metadata.Description)
				if desc == "" {
					desc = "—"
				}
				// Truncate description for table if too long
				if len(desc) > 100 {
					desc = desc[:97] + "..."
				}
				name := sanitizeMarkdownCell(s.Name)
				locBase := sanitizeMarkdownCell(filepath.Base(s.Path))
				locRel := sanitizeMarkdownCell(filepath.ToSlash(rel))
				out = append(out, fmt.Sprintf("| `%s` | %s | [%s](%s) |", name, desc, locBase, locRel))
			}
			sectionReplaced = true
			continue
		}

		if strings.HasPrefix(line, "### Auto-invoke Skills") {
			inAvailableSection = false
			inSection = true
			out = append(out, line)
			out = append(out, "")
			out = append(out, "When performing these actions, ALWAYS invoke the corresponding skill FIRST:")
			out = append(out, "")
			out = append(out, "| Action                              | Skill      |")
			out = append(out, "| ----------------------------------- | ---------- |")

			targetScope := scope
			if targetScope == "" {
				targetScope = "root"
			}
			rootSkills := AggregateMetadata(skills, targetScope)
			type autoInvokeRow struct {
				Action    string
				SkillName string
			}
			var rows []autoInvokeRow
			for _, s := range rootSkills {
				for _, action := range s.Metadata.AutoInvoke {
					rows = append(rows, autoInvokeRow{
						Action:    action,
						SkillName: s.Name,
					})
				}
			}

			sort.Slice(rows, func(i, j int) bool {
				aAct := strings.ToLower(rows[i].Action)
				bAct := strings.ToLower(rows[j].Action)
				if aAct != bAct {
					return aAct < bAct
				}
				return strings.ToLower(rows[i].SkillName) < strings.ToLower(rows[j].SkillName)
			})

			for _, r := range rows {
				action := sanitizeMarkdownCell(r.Action)
				skillName := sanitizeMarkdownCell(r.SkillName)
				paddedAction := fmt.Sprintf("%-35s", action)
				out = append(out, fmt.Sprintf("| %s | `%s` |", paddedAction, skillName))
			}

			sectionReplaced = true
			continue
		}

		if inSection || inAvailableSection {
			if strings.HasPrefix(line, "### ") || (inAvailableSection && strings.HasPrefix(line, "## ")) {
				inSection = false
				inAvailableSection = false
				out = append(out, "")
				out = append(out, line)
			}
			continue
		}

		out = append(out, line)
	}

	if !sectionReplaced {
		return string(content), false, nil
	}

	return strings.Join(out, "\n"), true, nil
}

func cleanupLegacyHarnessArtifacts(root string, dryRun bool) ([]runner.FileChange, error) {
	var changes []runner.FileChange
	skillsDir := filepath.Join(root, ".agents", "skills")

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	legacyFiles := []string{"sync.sh", "sync_test.sh", "sync_test.ps1"}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// skill-sync legitimately owns these filenames as embedded,
		// git-tracked assets (its SKILL.md depends on assets/sync.sh). They
		// are not legacy harness copies, so never prune them here.
		if entry.Name() == "skill-sync" {
			continue
		}
		assetsDir := filepath.Join(skillsDir, entry.Name(), "assets")
		for _, legacyName := range legacyFiles {
			path := filepath.Join(assetsDir, legacyName)
			if _, err := os.Stat(path); err == nil {
				relPath, err := filepath.Rel(root, path)
				if err != nil {
					relPath = path
				}

				var beforeContent string
				if raw, err := os.ReadFile(path); err == nil {
					beforeContent = string(raw)
				}

				change := runner.FileChange{
					Path:    filepath.ToSlash(relPath),
					Status:  "deleted",
					Before:  beforeContent,
					After:   "",
					Summary: "deleted legacy harness artifact",
				}
				changes = append(changes, change)

				if !dryRun {
					if err := os.Remove(path); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	return changes, nil
}
