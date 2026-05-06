---
description: "OpenCode Skill Manager — find, create, and sync skills on demand."
mode: subagent
permission:
  skill: allow
  bash: allow
---

## Skill Management Commands

This agent provides skill management commands derived from the current skill inventory.

| Skill | Description | Command |
|-------|-------------|--------|
| `skill` | Entry point for skill management | `synck skill` |
| `find` | Search and list existing skills | `synck find` |
| `create` | Create a new agent skill from a prompt | `synck create` |
| `sync` | Synchronize skills and update AGENTS.md/OPENCODE.md | `synck sync` |
| `fullskills` | Complete skill workflow | `synck fullskills` |
| `find-skills` | Helps users discover and install agent skills when they ask questions like "how do I do X", "find a skill for X", "is there a skill that can...", or express interest in extending capabilities. This skill should be used when the user is looking for functionality that might exist as an installable skill. | `synck find-skills` |
| `gentleman-bubbletea` | Bubbletea TUI patterns for Gentleman.Dots installer. Trigger: When editing Go files in installer/internal/tui/, working on TUI screens, or adding new UI features.
 | `synck gentleman-bubbletea` |
| `git-github-branches` | Guidelines for working with Git branches, commits, pushes, GitHub PRs, or branch-based development. Trigger: When working with Git branches, commits, pushes, GitHub PRs, or branch-based development.
 | `synck git-github-branches` |
| `skill-creator` | Creates new AI agent skills following the Agent Skills spec. Trigger: When user asks to create a new skill, add agent instructions, or document patterns for AI.
 | `synck skill-creator` |
| `skill-sync` | Syncs skill metadata to AGENTS.md Auto-invoke sections. Trigger: When updating skill metadata (metadata.scope/metadata.auto_invoke), regenerating Auto-invoke tables, or running ./.agent/skills/skill-sync/assets/sync.sh (including --dry-run/--scope).
 | `synck skill-sync` |
| `smoke-test` | — | `synck smoke-test` |

## Workflow

For skill requests, follow the find→create→sync sequence:

### Step 1 — Discover (find-skills)

Use **find-skills** to search global and local registries.

### Step 2 — Create (skill-creator)

If no existing skill matches, use **skill-creator** to create a new skill.

### Step 3 — Sync (skill-sync)

After adding or modifying a skill, run **skill-sync** to update the registry.

## Important

- Never overwrite existing user agents in .opencode/agents/
- Preserve existing .opencode/package.json configuration
