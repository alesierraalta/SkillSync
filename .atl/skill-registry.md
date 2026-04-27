# Skill Registry

**Delegator use only.** Any agent that launches sub-agents reads this registry to resolve compact rules, then injects them directly into sub-agent prompts. Sub-agents do NOT read this registry or individual SKILL.md files.

See `_shared/skill-resolver.md` for the full resolution protocol.

## User Skills

| Trigger | Skill | Path |
|---------|-------|------|
| user asks "how do I do X", "find a skill for X", "is there a skill that can...", or express interest in extending capabilities | find-skills | .agents/skills/find-skills/SKILL.md |
| editing Go files in installer/internal/tui/, working on TUI screens, or adding new UI features | gentleman-bubbletea | .agents/skills/gentleman-bubbletea/SKILL.md |
| user asks to create a new skill, add agent instructions, or document patterns for AI | skill-creator | .agents/skills/skill-creator/SKILL.md |
| updating skill metadata (metadata.scope/metadata.auto_invoke), regenerating Auto-invoke tables, or running sync.sh | skill-sync | .agents/skills/skill-sync/SKILL.md |
| writing Go tests, using teatest, or adding test coverage | go-testing | C:\Users\ismar\.gemini\antigravity\skills\go-testing\SKILL.md |
| user says "judgment day", "judgment-day", "review adversarial", "dual review", "doble review", "juzgar", "que lo juzguen" | judgment-day | C:\Users\ismar\.gemini\antigravity\skills\judgment-day\SKILL.md |

## Compact Rules

Pre-digested rules per skill. Delegators copy matching blocks into sub-agent prompts as `## Project Standards (auto-resolved)`.

### find-skills
- Use `npx skills find [query]` to discover skills interactively or by keyword.
- Install with `npx skills add <package> -g -y` for global, non-interactive setup.
- Verify quality: prefer >1K installs, known sources (vercel-labs, anthropics, microsoft), and >100 stars.
- Check https://skills.sh/ leaderboard first for popular domain-specific skills.

### gentleman-bubbletea
- Screen constants MUST be in `model.go` (iota-based).
- `Model` struct in `model.go` holds ALL application state.
- `Update(msg tea.Msg)` uses type switch for message handling.
- Separate `handle{Screen}Keys` function in `update.go` for each screen.
- Reset `m.Cursor = 0` on screen transitions; use `m.PrevScreen` for back navigation.

### skill-creator
- Create skill ONLY for repeated patterns, project-specific conventions, or complex workflows.
- Required: `.agent/skills/{skill-name}/SKILL.md` with Trigger in description.
- Use `assets/` for code templates/schemas, `references/` for LOCAL doc links.
- Naming: generic (`tech`), project (`proj-comp`), test (`proj-test-comp`), workflow (`act-target`).
- Register in `AGENTS.md` table after creation.

### skill-sync
- Run `./.agent/skills/skill-sync/assets/sync.sh` after any skill metadata change.
- Metadata MUST have `scope` (root, api, common, infra) and `auto_invoke` (string or list).
- Updates the `### Auto-invoke Skills` section in `AGENTS.md` files.

### go-testing
- Use table-driven tests (`t.Run`) for pure functions and state permutations.
- Test `Update` by simulating `tea.KeyMsg` and checking Model state.
- Use Charmbracelet's `teatest` for integration/interactive flow testing.
- Use Golden file testing for visual TUI output verification.
- organization: `*_test.go` next to source; use `testdata/` for golden files.

### judgment-day
- Launch TWO blind judges in parallel via `delegate` (async).
- Orchestrator synthesizes verdict: Confirmed, Suspect, Contradiction.
- Classify warnings: Real (production bug) vs Theoretical (contrived/controllable edge case).
- 2 fix iterations before user escalation; fix ONLY confirmed issues surgically.
- JUDGMENT: APPROVED requires 0 confirmed criticals and 0 confirmed real warnings.

## Project Conventions

| File | Path | Notes |
|------|------|-------|
| AGENTS.md | AGENTS.md | Index — lists available skills and auto-invoke triggers |
| CLAUDE.md | CLAUDE.md | Project instructions and RTK command registry |
| SPEC.md | SPEC.md | Source of truth for requirements and scenarios |
| DESIGN.md | DESIGN.md | Technical approach and architecture decisions |
| PROPOSAL.md | PROPOSAL.md | Change proposal with intent and scope |
| .atl/skill-registry.md | .atl/skill-registry.md | This file — pre-digested rules for delegators |

Read the convention files listed above for project-specific patterns and rules.
