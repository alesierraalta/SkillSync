# TUI Go Skills Specification (sdd/tui-go-skills/spec)

## Purpose
Interactive tool for local skill frontmatter management. Replaces manual `SKILL.md` editing + CLI `sync.sh`. Built with Go/Bubble Tea.

---

## Domain: Discovery & Parser

### Requirement: Skill Discovery
System MUST scan current directory and subfolders for `SKILL.md` files.
- **Scenario: Found Skills**
  - GIVEN folders `.agent/skills/a` and `.agent/skills/b` with `SKILL.md`.
  - WHEN TUI starts.
  - THEN List shows `a` and `b`.
- **Scenario: Missing Skills**
  - GIVEN no `SKILL.md` in subtree.
  - WHEN TUI starts.
  - THEN Show empty state message.

### Requirement: YAML Frontmatter Parsing
System MUST read/write YAML between `---` delimiters in `SKILL.md`.
- **Scenario: Read metadata**
  - GIVEN `SKILL.md` with `metadata: { name: "test", scope: ["api"] }`.
  - WHEN TUI loads skill.
  - THEN name and scope fields populate in form.
- **Scenario: Preserve Formatting**
  - GIVEN `SKILL.md` with comments in YAML.
  - WHEN TUI saves edits.
  - THEN Comments SHALL be preserved (use `yaml.v3` Node API).

### Requirement: Metadata Structure
`types.Skill` MUST include `Prefix string`.

### Requirement: Body Content Cleaning
The parser MUST ensure the `RawBody` is clean of leading/trailing whitespace introduced during split.
- **Scenario: Clean Parsing**
  - GIVEN a skill file with leading newlines after the frontmatter
  - WHEN `parser.Parse` is called
  - THEN the returned `Skill.RawBody` MUST have leading and trailing whitespace trimmed.

### Requirement: Prefix Preservation
The parser MUST extract and store all text preceding the first `---` delimiter as `Prefix`.
- **Scenario: Parse skill with header**
  - GIVEN file starts with "# My Skill\n---\n"
  - WHEN `Parse` called
  - THEN `Skill.Prefix` contains "# My Skill\n"
  - AND `Skill.Metadata` populated from YAML block

### Requirement: Idempotent Save with Prefix
`Save` MUST prepend `Prefix` to reconstructed file content.
- **Scenario: Save keeps header**
  - GIVEN `Skill.Prefix` is "# Header"
  - WHEN `Save` executed
  - THEN file starts with "# Header" followed by `---`

---

## Domain: UI & Interaction

### Requirement: Field Editing
UI MUST allow editing: `description`, `scope` (list), `auto_invoke` (list), `local_only` (bool).
- **Scenario: Edit Description**
  - GIVEN selected skill.
  - WHEN User enters Edit Mode → changes description.
  - THEN `SKILL.md` content reflects new text after Save.

### Requirement: Local-Only Skills
UI MUST allow marking skill as `local_only: true`.
- **Scenario: Mark Local**
  - GIVEN skill `X`.
  - WHEN Toggle `local_only` → Save.
  - THEN `metadata.local_only: true` added to `SKILL.md`.

### Requirement: Skill Content Viewing
System MUST allow viewing skill content as an inline preview (default) OR full-screen view.
- **Scenario: Maximize Preview**
  - GIVEN user on `ScreenList`.
  - WHEN user presses 'v'.
  - THEN system MUST transition to `ScreenContentView` (100% width, no list).
  - AND 'esc' SHALL return to `ScreenList` (split view).

### Requirement: ScreenList Split View
`ScreenList` MUST display two horizontal panels: List (40% width) and Preview (60% width).
- **Scenario: Layout Calculation**
  - GIVEN user on `ScreenList`.
  - WHEN `WindowSizeMsg` received with width $W$.
  - THEN List width SHALL be $0.4 * W$ AND Preview width SHALL be $0.6 * W$.
- **Scenario: Instant Live Preview**
  - GIVEN user navigates list in `ScreenList`.
  - WHEN selected item changes.
  - THEN Preview viewport MUST update immediately with rendered Markdown of selected skill.
- **Scenario: Independent Preview Scroll**
  - GIVEN user in `ScreenList`.
  - WHEN user uses scroll keys (PageUp/PageDown).
  - THEN Preview viewport scrolls AND list selection SHALL NOT change.
- **Scenario: Split-View Resizing**
  - GIVEN terminal width changes.
  - WHEN `WindowSizeMsg` processed.
  - THEN `glamour` renderer MUST be re-initialized with new Preview panel width.
  - AND content re-rendered to fit new dimensions.

### Requirement: Dynamic WordWrap
TUI MUST re-render Markdown content when window size changes.
- **Scenario: Window Resize**
  - GIVEN user is on `ScreenContentView`
  - WHEN `WindowSizeMsg` received
  - THEN `glamour` renderer re-initialized with new width
  - AND viewport content updated with new wrapped text

---

## Domain: Orchestration

### Requirement: Sync Action
TUI MUST provide key/button to trigger `sync.sh`.
- **Scenario: Trigger Sync**
  - GIVEN TUI running.
  - WHEN User presses `S` (sync).
  - THEN Execute `./.agent/skills/skill-sync/assets/sync.sh`.
  - AND Show stdout/stderr in TUI status bar.

---

## Testing Phase: Consistency & Risks

### Requirement Validation
1. **List skills**: Covered (Discovery).
2. **Parse Frontmatter**: Covered (Parser).
3. **Interactive UI**: Covered (Field Editing).
4. **Sync Button**: Covered (Sync Action).
5. **local_only**: Covered (UI/Metadata).

### Risks & Mitigations
- **Risk: sync.sh path failure**
  - Validation: TUI MUST check if `sync.sh` exists relative to current dir.
- **Risk: sync.sh ignoring local_only**
  - Validation: If `local_only: true`, TUI SHOULD warn user that `sync.sh` needs update to exclude it, or TUI MUST patch sync process.
- **Risk: Data corruption on Save**
  - Validation: System MUST perform atomic write (write temp -> rename).

### Acceptance Criteria
- [ ] List all local skills.
- [ ] Edit/Save metadata fields correctly.
- [ ] `SKILL.md` body (markdown) remains untouched after YAML edit.
- [ ] Sync script executes and output visible.
- [ ] `local_only` flag persists in YAML.
