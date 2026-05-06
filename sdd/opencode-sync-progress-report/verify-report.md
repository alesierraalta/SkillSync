# VERIFY-REPORT: opencode-sync-progress-report (PR2 Core Engine Slice)

**Status:** 🟡 WARNING (Instrumentation Missing)
**Change:** opencode-sync-progress-report
**Version:** PR2 Core Engine Slice
**Mode:** Standard SDD

---

## Executive Summary
Verification of the PR2 Core Engine Slice reveals that while the foundation for progress reporting (plumbing in `Options`) and the core sync reporting (plumbing in `SyncReport`) are present, the actual instrumentation of the progress callback in `internal/opencode/sync.go` is missing. Additionally, two regeneration functions deviate from the reporting pattern.

---

## Requirements Checklist (PR2)

### 1. Progress Plumbing & Callback
- [x] **REQ-1.1: Options Update**: `ProgressCb` added to `opencode.Options`.
- [ ] **REQ-1.2: Loop Instrumentation**: `ProgressCb` is NOT invoked during the mirroring loop in `SyncSkills`.
- [ ] **REQ-1.3: Nil-Safety**: No explicit `if opts.ProgressCb != nil` guards (since no calls exist).

### 2. Core Reporting Plumbing
- [x] **REQ-2.1: Sync Report Return**: `SyncSkills` and `RegenerateTools` return `*runner.SyncReport`.
- [x] **REQ-2.2: Change Detection**: Content-addressed copy correctly populates `FileChange` records.
- [x] **REQ-2.3: Status Handling**: Symlinks are correctly identified and reported with `"symlinked"` status.

### 3. Design Deviations
- [ ] **DEV-3.1: RegenerateAgent**: Returns only `error`, should return `(*runner.SyncReport, error)`.
- [ ] **DEV-3.2: RegenerateCommands**: Returns only `error`, should return `(*runner.SyncReport, error)`.

---

## Technical Findings

### File Snapshot & Change Detection
- **Status:** ✅ PASSED
- `internal/opencode/sync.go` and `internal/opencode/tools.go` correctly implement content comparison before writing.
- `report.Changes` is correctly populated with `Before`, `After`, `Diff`, and `Summary` using `internal/diff`.

### Symlink / Mirrored Status Handling
- **Status:** ✅ PASSED
- `SyncSkills` preserves symlinks correctly and records them in the sync report.

### Call-site Adapters
- **Status:** ✅ PASSED
- `internal/ui/ecosystem.go` calls the new `SyncSkills` implementation.

---

## Critical Risks
- **Incomplete UI Feedback**: The TUI will not show progress during the sync operation because the hook is never triggered by the core logic.
- **Audit Gaps**: Modifications to agent markdown and command files are not included in the sync report, leading to an incomplete "What changed" view for the user.

---

## Next Steps
1. **Instrument `SyncSkills`**: Add `opts.ProgressCb` calls in `internal/opencode/sync.go`.
2. **Refactor Regenerators**: Update `RegenerateAgent` and `RegenerateCommands` to return `SyncReport`.
3. **Verify with Tests**: Add unit tests in `sync_test.go` to assert progress callback frequency and report completeness.
