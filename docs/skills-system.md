# The Skills System

The Skills system is the core of this project. It defines how AI agents receive instructions, patterns, and tools.

## Discovery Mechanics

The TUI scans the project root and subdirectories for any folder containing a `SKILL.md` file. It specifically looks into:
- `.agents/skills/`
- `.claude/skills/`
- `.gemini/skills/`
- `.codex/skills/`

## `SKILL.md` Format

Every skill MUST follow this structure:

```markdown
---
name: skill-name
description: What this skill does.
scope: project | personal
auto_invoke: [trigger actions]
---

# Skill Name

## Instructions
Detailed instructions for the AI agent...

## Patterns
Code patterns to follow...
```

### Metadata (YAML)
- **name**: Unique identifier for the skill.
- **description**: Short summary displayed in the TUI.
- **scope**: Whether it's specific to this project or a personal preference.
- **auto_invoke**: A list of actions that should automatically trigger this skill (e.g., "editing Go files").

## Skill Registry (`AGENTS.md`)

The `AGENTS.md` file at the root of the project is the **public registry**. It is automatically managed by the `skill-sync` skill. 

> [!IMPORTANT]
> Never edit the tables in `AGENTS.md` manually. They will be overwritten during the next sync. Edit the source `SKILL.md` files instead.

## Mirror Management

Skill mirrors exist in multiple locations across the project. The following table documents each mirror's source of truth and management strategy:

| Mirror Location | Source of Truth | Writer / Guard | Notes |
|---|---|---|---|
| `.opencode/` | `.agents/skills/` | `syncengine.SyncSkills` (real-run only) | Automatically generated mirror; regenerated during synchronization. Read-only during dry-run. |
| `.claude/skills/` | Manual | None | Manually maintained mirror for Claude-specific configurations. No automated writer; updates are manual. |
| `.qwen/skills/` | Manual | None | Manually maintained mirror for Qwen-specific configurations. No automated writer; updates are manual. |
| `internal/coreskills/` | `.agents/skills/{find-skills,skill-creator,skill-sync}` | Manual + `TestCoreSkillDrift` guard | Embedded snapshot of core skills; updated manually at build time. Consistency enforced by `internal/coreskills/drift_test.go` `TestCoreSkillDrift`, which fails CI if the embed diverges from source. |

### Maintaining Mirrors

- **`.opencode/`**: Generated during real synchronization runs. Do not edit manually.
- **`.claude/skills/` and `.qwen/skills/`**: Update these mirrors manually when core skills change and you need customizations for those platforms.
- **`internal/coreskills/`**: Update the embedded snapshot manually (via `go:embed`), then verify with `TestCoreSkillDrift`. The test will fail if the embed becomes out of sync with source.
