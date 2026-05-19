# Agent Skills

This document lists the AI skills available in the TUI project.

## Available Skills

| Skill | Description | Location |
| ----- | ----------- | -------- |
| `context7` | Disciplined use of Context7 docs lookup: resolve library ID first, query docs with precise questi... | [SKILL.md](.agents/skills/context7/SKILL.md) |
| `find-skills` | Helps users discover and install agent skills when they ask questions like "how do I do X", "find... | [SKILL.md](.agents/skills/find-skills/SKILL.md) |
| `gentleman-bubbletea` | Bubbletea TUI patterns for Gentleman.Dots installer. Trigger: When editing Go files in installer/... | [SKILL.md](.agents/skills/gentleman-bubbletea/SKILL.md) |
| `git-github-branches` | Guidelines for working with Git branches, commits, pushes, GitHub PRs, or branch-based developmen... | [SKILL.md](.agents/skills/git-github-branches/SKILL.md) |
| `sequential-thinking` | Disciplined use of the sequential-thinking tool for complex planning, analysis, revisions, branch... | [SKILL.md](.agents/skills/sequential-thinking/SKILL.md) |
| `serena-mcp-tools` | Serena MCP and tool-selection discipline for high-efficiency codebase navigation, symbolic editin... | [SKILL.md](.agents/skills/serena-mcp-tools/SKILL.md) |
| `skill-creator` | Creates new AI agent skills following the Agent Skills spec. Trigger: When user asks to create a ... | [SKILL.md](.agents/skills/skill-creator/SKILL.md) |
| `skill-sync` | Syncs skill metadata to AGENTS.md Auto-invoke sections. Trigger: When updating skill metadata (me... | [SKILL.md](.agents/skills/skill-sync/SKILL.md) |
### Auto-invoke Skills

When performing these actions, ALWAYS invoke the corresponding skill FIRST:

| Action                              | Skill      |
| ----------------------------------- | ---------- |
| context7                            | `context7` |
| gentleman-bubbletea                 | `gentleman-bubbletea` |
| sequential-thinking                 | `sequential-thinking` |
| serena-mcp-tools                    | `serena-mcp-tools` |
| skill-creator                       | `skill-creator` |