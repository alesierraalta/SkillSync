# Agent Skills

This document lists the AI skills available in the TUI project.

## Available Skills

| Skill | Description | Location |
| ----- | ----------- | -------- |
| `bash-defensive-patterns` | Master defensive Bash programming techniques for production-grade scripts. Use when writing robus... | [SKILL.md](.agents/skills/bash-defensive-patterns/SKILL.md) |
| `context7` | Disciplined use of Context7 docs lookup: resolve library ID first, query docs with precise questi... | [SKILL.md](.agents/skills/context7/SKILL.md) |
| `find-skills` | Helps users discover and install agent skills when they ask questions like "how do I do X", "find... | [SKILL.md](.agents/skills/find-skills/SKILL.md) |
| `gentleman-bubbletea` | Bubbletea TUI patterns for Gentleman.Dots installer. Trigger: When editing Go files in installer/... | [SKILL.md](.agents/skills/gentleman-bubbletea/SKILL.md) |
| `git-github-branches` | Guidelines for working with Git branches, commits, pushes, GitHub PRs, or branch-based developmen... | [SKILL.md](.agents/skills/git-github-branches/SKILL.md) |
| `golang-patterns` | Idiomatic Go patterns, best practices, and conventions for building robust, efficient, and mainta... | [SKILL.md](.agents/skills/golang-patterns/SKILL.md) |
| `golang-testing` | Go testing patterns including table-driven tests, subtests, benchmarks, fuzzing, and test coverag... | [SKILL.md](.agents/skills/golang-testing/SKILL.md) |
| `sequential-thinking` | Disciplined use of the sequential-thinking tool for complex planning, analysis, revisions, branch... | [SKILL.md](.agents/skills/sequential-thinking/SKILL.md) |
| `serena-mcp-tools` | Serena MCP para navegacion semantica, edicion por simbolos y refactors con bajo coste de tokens. ... | [SKILL.md](.agents/skills/serena-mcp-tools/SKILL.md) |
| `skill-creator` | Creates new AI agent skills following the Agent Skills spec. Trigger: When user asks to create a ... | [SKILL.md](.agents/skills/skill-creator/SKILL.md) |
| `skill-sync` | Syncs skill metadata to AGENTS.md Auto-invoke sections. Trigger: When updating skill metadata (me... | [SKILL.md](.agents/skills/skill-sync/SKILL.md) |
| `test-prompt-scope` | test prompt scope. Trigger: When the user asks for test prompt scope. | [SKILL.md](.agents/skills/test-prompt-scope/SKILL.md) |
### Auto-invoke Skills

When performing these actions, ALWAYS invoke the corresponding skill FIRST:

| Action                              | Skill      |
| ----------------------------------- | ---------- |
| After creating/modifying a skill    | `skill-sync` |
| Branch-based development            | `git-github-branches` |
| Complex planning or multi-step analysis | `sequential-thinking` |
| Conventional commits                | `git-github-branches` |
| Creating new skills                 | `skill-creator` |
| Creating or updating test-prompt-scope workflows | `test-prompt-scope` |
| GitHub PR preparation               | `git-github-branches` |
| Regenerate AGENTS.md Auto-invoke tables (synck sync) | `skill-sync` |
| Searching external documentation    | `context7` |
| Searching for or installing new agent skills | `find-skills` |
| Troubleshoot why a skill is missing from AGENTS.md auto-invoke | `skill-sync` |
| Using Serena MCP, semantic code navigation, symbol editing, or refactors | `serena-mcp-tools` |
| Working on TUI screens or adding new UI features in Go | `gentleman-bubbletea` |
| Working with Git branches           | `git-github-branches` |