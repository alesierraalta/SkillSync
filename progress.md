# Validation Protocol: Sync + AGENTS.md Integrity After create-skill Changes

## Audit Summary

Scout findings were audited against the codebase. All core claims are correct, with two behavioral bugs identified and one test gap.

## Verified Findings

### Correct: Sync Engine Flow

- `synck sync` → `handleSync()` → `syncengine.Sync(root, opts)` → `DiscoverSkills()` + `UpdateAgentsMarkdown()` — confirmed in `cmd/synck/main.go:handleSync()` and `internal/syncengine/sync.go`.
- Skills are discovered from `.agents/skills/<name>/SKILL.md` — confirmed in `DiscoverSkills()` at `sync.go:74`.
- `UpdateAgentsMarkdown` matches headers via `strings.HasPrefix` for `## Available Skills` and `### Auto-invoke Skills` — confirmed at `sync.go:108` and `sync.go:127`.
- All 15 test packages pass: `go test ./...` — confirmed.

### Correct: Precondition Requirements

- `AGENTS.md` must exist — `UpdateAgentsMarkdown` returns `nil` silently if file is missing (`sync.go:99-101`). No file created.
- Skills need `SKILL.md` with parseable YAML frontmatter in `.agents/skills/<name>/` — confirmed in `DiscoverSkills()` (invalid skills are silently skipped via `continue` at `sync.go:91`).
- For auto-invoke table: `scope` must contain `"root"` (checked via `strings.Contains` in `AggregateMetadata` at `sync.go:99`) and `auto_invoke` must be truthy (checked via `isTrue` in `parser.go:72`).

### Correct: Failure Modes

1. **Header typo** — case-sensitive `strings.HasPrefix` match. Wrong header = section skipped, no error.
2. **Scope missing** — skill won't appear in auto-invoke table (`AggregateMetadata` filters out non-root).
3. **YAML parse error** — skill silently skipped (`parser.Parse` returns error, `DiscoverSkills` continues).

## Bugs Identified

### Bug 1: create-skill auto_invoke loses descriptive value on round-trip

**Files:** `cmd/synck/main.go:renderSkillMarkdown()`, `internal/parser/parser.go:isTrue()`, `internal/parser/parser.go:updateMappingNode()`

`renderSkillMarkdown` creates skills with:

```
auto_invoke: 'Creating or updating %s workflows'
```

This is a non-empty string, so `isTrue()` returns `true`. The `Metadata.AutoInvoke` field is typed as `bool`. When `Format()`/`Save()` writes back via `updateMappingNode()`, it writes `auto_invoke: true` (boolean), **losing the original descriptive action text**.

**Impact:** Any newly created skill loses its intended auto-invoke action label on the next parse+save cycle. The auto-invoke table will show the skill **name** as the action (from `sync.go:141: action := s.Name`), not the descriptive text.

### Bug 2: create-skill does not auto-sync

**File:** `cmd/synck/main.go:handleCreateSkill()`

After creating the skill scaffold, the function prints `"Next: run 'synck sync' to register the new skill."` but does not invoke sync automatically. The user must manually run a second command.

**Impact:** After `synck create <prompt>`, the new skill exists on disk but is NOT registered in `AGENTS.md` until the user runs `synck sync` separately.

### Bug 3: DiscoverSkills silently swallows parse errors

**File:** `internal/syncengine/sync.go:91`

```go
skill, err := parser.Parse(skillPath)
if err != nil {
    continue // skip invalid skills
}
```

No warning is emitted. A skill with broken YAML frontmatter disappears from sync without any indication.

## Test Gap

**Missing:** No integration test covers the `create-skill → sync → verify AGENTS.md` flow.

`main_test.go:TestCreateSkill_GeneratesPack` only verifies scaffold files exist. It does not run `synck sync` afterward or verify AGENTS.md was updated. The closest test is `TestHandleSync_ProgressAndReport` which uses a pre-made skill, not a newly created one.

## Validation Protocol

### Test 1: Create Skill + Sync Integration

**Goal:** Verify `synck create <prompt>` followed by `synck sync` produces a correct AGENTS.md entry.

**Setup:**

```bash
cd /tmp/test-skill-sync
mkdir -p .agents
echo '# Agent Skills

## Available Skills

| Skill | Description | Location |
| ----- | ----------- | -------- |

### Auto-invoke Skills

When performing these actions, ALWAYS invoke the corresponding skill FIRST:

| Action                              | Skill      |
| ----------------------------------- | ---------- |
' > AGENTS.md
```

**Steps:**

1. `synck create "My new test skill for validation"`
2. `synck sync --dry-run --verbose`
3. `synck sync --verbose`

**Expected Output (step 2 — dry-run):**

```
🔄 Running synchronization...
→ [1/8] Discovering skills...
→ [2/8] Updating AGENTS.md...
...
Changed files:

  AGENTS.md (modified) +N -M
  ────────────────────────────────────────
  @@ ... @@
  +| `my-new-test-skill-for-validation` | My new test skill for validation. Trigger: When the user asks for My new test skill for validation. ... | [SKILL.md](.agents/skills/my-new-test-skill-for-validation/SKILL.md) |
  ...
```

**Expected Output (step 3 — apply):**

```
🔄 Running synchronization...
→ [1/8] Discovering skills...
→ [2/8] Updating AGENTS.md...
→ [3/8] Mirroring skills...
→ [4/8] Parsing mirrored skills...
→ [5/8] Regenerating tools...
→ [6/8] Regenerating agent...
→ [7/8] Copying AGENTS.md...
→ [8/8] Regenerating commands...
✅ Synchronization complete.

Changed files:

  AGENTS.md (modified) +N -M
  ...
```

**Verification after step 3:**

```bash
cat AGENTS.md
```

The new skill should appear in both tables:

- In "Available Skills" table with its name, truncated description, and link.
- In "Auto-invoke Skills" table because `create-skill` sets `scope: [root]`.

### Test 2: Idempotency

**Goal:** Running `synck sync` twice produces no changes on second run.

**Steps:**

1. Run `synck sync` (creates initial state)
2. Run `synck sync --verbose` again

**Expected:** Output contains `No files changed.`

### Test 3: Edge Case — Skill Without Root Scope

**Setup:** Create a skill with `scope: [api]` (not root).
**Expected:** Skill appears in "Available Skills" table but NOT in "Auto-invoke Skills" table.

### Test 4: Edge Case — Skill With auto_invoke: false

**Setup:** Create a skill with `scope: [root]` but `auto_invoke: false`.
**Expected:** Skill appears in "Available Skills" table but NOT in "Auto-invoke Skills" table.

### Test 5: Edge Case — Missing AGENTS.md

**Setup:** Remove `AGENTS.md` from project root.
**Steps:** `synck sync`
**Expected:** No error. Sync completes. AGENTS.md is NOT created (UpdateAgentsMarkdown returns nil on missing file).

### Test 6: Edge Case — Malformed SKILL.md

**Setup:** Create `.agents/skills/broken/SKILL.md` with invalid YAML (e.g., `---\nbad: yaml: : :\n---`).
**Steps:** `synck sync --verbose`
**Expected:** No error. Sync completes. Broken skill is silently skipped (no entry in AGENTS.md).

### Test 7: Edge Case — Header Variations

**Setup:** AGENTS.md with `## available skills` (lowercase) instead of `## Available Skills`.
**Steps:** `synck sync`
**Expected:** Section is NOT replaced (case-sensitive prefix match). This is the documented failure mode.

### Test 8: Edge Case — Existing Skill With Same Name

**Steps:** `synck create "Go test quality guardrails"` (same as existing test)
**Expected:** Error: `strict validation failed: skill already exists at ...`

### Test 9: Edge Case — Weak Prompt

**Steps:** `synck create "auth"` (1 word, < 3 words)
**Expected:** Error: `strict validation failed: prompt must include at least 3 words...`

### Test 10: OPENCODE.md Integrity

**Goal:** Verify AGENTS.md content propagates to OPENCODE.md.
**Steps:** Run `synck sync`, then check `OPENCODE.md`.
**Expected:** OPENCODE.md contains the same auto-invoke table as AGENTS.md (via `opencode.CopyAgentsMD`).

## Recommended Test Additions (Code)

Add to `cmd/synck/main_test.go`:

```go
func TestCreateSkillThenSync(t *testing.T) {
    tmpDir := t.TempDir()
    origDir, _ := os.Getwd()
    _ = os.Chdir(tmpDir)
    defer os.Chdir(origDir)

    _ = os.MkdirAll(".agents", 0755)
    agentsContent := `# Agent Skills

## Available Skills

| Skill | Description | Location |
| ----- | ----------- | -------- |

### Auto-invoke Skills

When performing these actions, ALWAYS invoke the corresponding skill FIRST:

| Action                              | Skill      |
| ----------------------------------- | ---------- |
`
    _ = os.WriteFile("AGENTS.md", []byte(agentsContent), 0644)

    // Create skill
    err := run([]string{"synck", "create", "Test integration skill for validation"})
    if err != nil {
        t.Fatalf("create failed: %v", err)
    }

    // Run sync
    err = handleSync(nil)
    if err != nil {
        t.Fatalf("sync failed: %v", err)
    }

    // Verify AGENTS.md was updated
    content, _ := os.ReadFile("AGENTS.md")
    s := string(content)
    if !strings.Contains(s, "test-integration-skill-for-validation") {
        t.Errorf("AGENTS.md missing new skill name, got:\n%s", s)
    }
    if !strings.Contains(s, "Test integration skill for validation") {
        t.Errorf("AGENTS.md missing new skill description, got:\n%s", s)
    }
    // Verify auto-invoke entry (name used as action label)
    if !strings.Contains(s, "test-integration-skill-for-validation") {
        t.Errorf("AGENTS.md auto-invoke table missing new skill")
    }
}
```
