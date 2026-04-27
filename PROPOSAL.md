# Proposal: TUI Go Skills Management

## Intent

Current skill management manual. Editing `SKILL.md` frontmatter + running `sync.sh` prone to error. Need interactive tool for local browsing, editing, and syncing.

## Scope

### In Scope
- List skills from `.agents/skills`.
- Read/Write YAML frontmatter in `SKILL.md`.
- Trigger `sync.sh` to update `agents.md`.
- Bubble Tea TUI (MVU pattern).

### Out of Scope
- Remote git operations.
- Global skill registry management.
- Creation of new skill templates (handled by `skill-creator`).

## Capabilities

### New Capabilities
- `skill-manager-tui`: core TUI app.
- `frontmatter-parser`: Go logic to handle YAML metadata.
- `sync-orchestrator`: bridge between TUI and shell sync scripts.

### Modified Capabilities
- None.

## Approach

- Stack: Go 1.21+, Bubble Tea, Lip Gloss.
- Standard "gentleman-bubbletea" structure:
  - `Model`: tracks skill list, selection, edit state.
  - `Update`: handles keys (arrows, enter, esc) and file IO.
  - `View`: renders lists and forms.
- Use `gopkg.in/yaml.v3` for safe metadata round-tripping.
- Call existing `sync.sh` via `os/exec` to ensure logic parity.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `/TUI` | New | Main Go project home. |
| `TUI/main.go` | New | Entry point. |
| `TUI/internal/ui` | New | TUI components (list, form). |
| `TUI/internal/skill` | New | Skill IO and metadata logic. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| YAML data loss | Low | Use structured parsing; backup file before write. |
| Sync failure | Med | Validate `sync.sh` output; report errors in TUI. |
| `agents.md` corruption | Low | Sync script is source of truth; TUI only triggers it. |

## Rollback Plan
Delete `TUI` folder. System reverts to manual CLI/Editor workflow. No changes to core `.agents/skills` files until explicit Save in TUI.

## Dependencies
- Go compiler.
- Existing `sync.sh` and folder structure.

## Success Criteria
- [ ] TUI lists all local skills.
- [ ] Changing name/description in TUI updates `SKILL.md`.
- [ ] Sync action correctly updates `agents.md`.
- [ ] No regression in manual script usage.
