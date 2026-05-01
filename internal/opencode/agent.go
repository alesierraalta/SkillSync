package opencode

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skillsync/tui/internal/types"
)

// RegenerateAgent writes .opencode/agents/skill-manager.md with a table
// listing all provided skills and their commands.
func RegenerateAgent(root string, skills []types.Skill, dryRun bool) error {
	agentsDir := filepath.Join(root, ".opencode", "agents")
	agentPath := filepath.Join(agentsDir, "skill-manager.md")

	if dryRun {
		fmt.Printf("[dry-run] would regenerate skill-manager.md with %d skills\n", len(skills))
		return nil
	}

	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", agentsDir, err)
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("description: \"OpenCode Skill Manager — find, create, and sync skills on demand.\"\n")
	b.WriteString("mode: subagent\n")
	b.WriteString("permission:\n")
	b.WriteString("  skill: allow\n")
	b.WriteString("  bash: allow\n")
	b.WriteString("---\n")
	b.WriteString("\n")
	b.WriteString("## Skill Management Commands\n")
	b.WriteString("\n")
	b.WriteString("This agent provides skill management commands derived from the current skill inventory.\n")
	b.WriteString("\n")

	// Build command table
	if len(skills) > 0 {
		b.WriteString("| Skill | Description | Command |\n")
		b.WriteString("|-------|-------------|--------|\n")
		for _, s := range skills {
			cmd := "synck " + s.Name
			desc := s.Metadata.Description
			if desc == "" {
				desc = "—"
			}
			b.WriteString(fmt.Sprintf("| `%s` | %s | `%s` |\n", s.Name, desc, cmd))
		}
		b.WriteString("\n")
	}

	// Static workflow guidance
	b.WriteString("## Workflow\n")
	b.WriteString("\n")
	b.WriteString("For skill requests, follow the find→create→sync sequence:\n")
	b.WriteString("\n")
	b.WriteString("### Step 1 — Discover (find-skills)\n")
	b.WriteString("\n")
	b.WriteString("Use **find-skills** to search global and local registries.\n")
	b.WriteString("\n")
	b.WriteString("### Step 2 — Create (skill-creator)\n")
	b.WriteString("\n")
	b.WriteString("If no existing skill matches, use **skill-creator** to create a new skill.\n")
	b.WriteString("\n")
	b.WriteString("### Step 3 — Sync (skill-sync)\n")
	b.WriteString("\n")
	b.WriteString("After adding or modifying a skill, run **skill-sync** to update the registry.\n")
	b.WriteString("\n")
	b.WriteString("## Important\n")
	b.WriteString("\n")
	b.WriteString("- Never overwrite existing user agents in .opencode/agents/\n")
	b.WriteString("- Preserve existing .opencode/package.json configuration\n")

	return os.WriteFile(agentPath, []byte(b.String()), 0644)
}

// CopyAgentsMD copies AGENTS.md to OPENCODE.md.
func CopyAgentsMD(root string) error {
	src := filepath.Join(root, "AGENTS.md")
	dst := filepath.Join(root, "OPENCODE.md")

	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read AGENTS.md: %w", err)
	}

	return os.WriteFile(dst, content, 0644)
}
