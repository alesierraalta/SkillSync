# Skill Registry

**Delegator use only.** Any agent that launches sub-agents reads this registry to resolve compact rules, then injects them directly into sub-agent prompts. Sub-agents do NOT read this registry or individual SKILL.md files.

## User Skills

| Trigger | Skill | Path |
|---------|-------|------|
| find-skills | find-skills | .agents/skills/find-skills/SKILL.md |
| gentleman-bubbletea | gentleman-bubbletea | .agents/skills/gentleman-bubbletea/SKILL.md |
| skill-creator | skill-creator | .agents/skills/skill-creator/SKILL.md |
| skill-sync | skill-sync | .agents/skills/skill-sync/SKILL.md |

## Compact Rules

### find-skills
- Use to find skills if not listed in agent tools or project registry.
- Trigger only when user explicitly asks for new capabilities or "how to X".

### gentleman-bubbletea
- Bubbletea TUI patterns for Gentleman.Dots.
- Use when editing `internal/ui/` or adding UI.
- Follow `internal/ui/` conventions.

### skill-creator
- Create skills following Agent Skills spec.
- Trigger when adding new agent behavior/instructions.

### skill-sync
- Sync skill metadata and refresh registry.
- Run when modifying `SKILL.md` or `AGENTS.md` tables.

## Project Conventions

| File | Path | Notes |
|------|------|-------|
| AGENTS.md | AGENTS.md | Index — defines rules/personality |
| DESIGN.md | DESIGN.md | Architecture decisions |
| SPEC.md | SPEC.md | Source of truth for specs |
| PROPOSAL.md | PROPOSAL.md | Change proposals |
