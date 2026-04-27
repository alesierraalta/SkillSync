# Agent Skills

This document lists the AI skills available in the TUI project.

## Available Skills

| Skill | Description | Location |
| ----- | ----------- | -------- |
| `find-skills` | Helps find and install skills | [.agents/skills/find-skills/SKILL.md](.agents/skills/find-skills/SKILL.md) |
| `gentleman-bubbletea` | Bubbletea TUI patterns | [.agents/skills/gentleman-bubbletea/SKILL.md](.agents/skills/gentleman-bubbletea/SKILL.md) |
| `skill-creator` | Creates new agent skills | [.agents/skills/skill-creator/SKILL.md](.agents/skills/skill-creator/SKILL.md) |
| `skill-sync` | Syncs skill metadata | [.agents/skills/skill-sync/SKILL.md](.agents/skills/skill-sync/SKILL.md) |

### Auto-invoke Skills

When performing these actions, ALWAYS invoke the corresponding skill FIRST:

| Action | Skill |
| ------ | ----- |
| After creating/modifying a skill | `skill-sync` |
| Creating new skills | `skill-creator` |
| Regenerate AGENTS.md Auto-invoke tables (sync.sh) | `skill-sync` |
| Searching for or installing new agent skills | `find-skills` |
| Troubleshoot why a skill is missing from AGENTS.md auto-invoke | `skill-sync` |
| Working on TUI screens or adding new UI features in Go | `gentleman-bubbletea` |

