# Tasks: Fix Cursor and Metadata Bugs in Skillsync TUI

## Review Workload Forecast
```text
Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: stacked-to-main
400-line budget risk: Low
```

## Suggested Work Units
- **Work Unit 1**: Core Parser Enhancements (TDD tests + parsing implementation for scratch generation and nested metadata).
- **Work Unit 2**: TUI UI Alignment (TUI cursor condition update).

---

## Phase 1: Foundation (Tests / Preparation)

- **1.1: Write failing unit test for scratch frontmatter formatting**
  - [x] Add `TestFormatFromScratch` in `internal/parser/parser_test.go`.
  - [x] Instantiate a `types.Skill` without a `Path` (so `yamlStr == ""`).
  - [x] Set all metadata properties: `Name`, `Description`, `Scope`, `AutoInvoke`, and `LocalOnly`.
  - [x] Assert that `Format(skill)` returns a string containing the serialized frontmatter with all five properties.

- **1.2: Write failing unit test for nested metadata preservation**
  - [x] Add `TestFormatPreservesNestedMetadata` in `internal/parser/parser_test.go`.
  - [x] Set up an input string containing nested `metadata` (e.g. `metadata.scope`, `metadata.auto_invoke`) and custom fields (e.g. `metadata.author`, `metadata.version`, `license`).
  - [x] Parse the string, update `Scope` via `skill.Metadata.Scope`, and format it back.
  - [x] Assert that:
    - [x] `metadata.scope` is updated inside the nested `metadata` block.
    - [x] No duplicate `scope` key is added at the root level of the frontmatter.
    - [x] Custom fields (e.g. `metadata.author`, `metadata.version`, `license`) are preserved exactly.

---

## Phase 2: Core Implementation

- **2.1: Update `Format` scratch generation branch**
  - [x] In `internal/parser/parser.go`, rewrite the `yamlStr == ""` branch.
  - [x] Define a temporary struct with YAML tags matching the expected keys (`name`, `description`, `scope`, `local_only`, `auto_invoke`).
  - [x] Use `yaml.NewEncoder` with indentation set to 2 spaces to serialize the struct.
  - [x] Prefix with headers and append the `RawBody`.

- **2.2: Implement nested metadata support in `updateMappingNode`**
  - [x] Locate `updateMappingNode` in `internal/parser/parser.go`.
  - [x] Inspect the root MappingNode to see if it contains a key named `"metadata"`.
  - [x] If a `"metadata"` key is found and its value is a MappingNode, route the updates for `scope` and `auto_invoke` to that nested MappingNode instead of updating/appending them to the root MappingNode.
  - [x] Ensure root-level keys for `scope` and `auto_invoke` are deleted or not appended if updated in the nested block.

- **2.3: Fix TUI installer cursor duplicate check**
  - [x] In `internal/ui/installer_model.go` at the end of the `OptionsView` method, find the rendering logic for the `[ Execute Install ]` button's cursor.
  - [x] Modify the condition checking `m.Cursor == 9+storageOffset` to check `m.Cursor == 10+storageOffset`.

---

## Phase 3: Testing / Verification

- **3.1: Execute Parser unit tests**
  - [x] Run the parser tests: `go test -v ./internal/parser/...`
  - [x] Ensure both newly added TDD tests pass successfully along with all existing parser unit tests.

- **3.2: Execute UI unit tests**
  - [x] Run the UI tests: `go test -v ./internal/ui/...`
  - [x] Ensure all TUI tests build and pass cleanly.

---

## Phase 4: Cleanup / Documentation

- **4.1: Perform code quality check and formatting**
  - [x] Run `go fmt ./...` and `go vet ./...` to guarantee code standard compliance.
  
- **4.2: Final verification checklist**
  - [x] Verify no comments or custom YAML attributes are deleted during the formatting round-trips.

---

## Phase 5: Judgment Day Round 2 Fixes (TDD)
- [x] Remove unreachable key checks inside space/enter block in `internal/ui/installer_model.go`.
- [x] Prevent Windows os.Rename failure in `Save()` inside `internal/parser/parser.go` by calling `os.Remove` on target first, and logging clean up errors.
- [x] Handle empty `Scope` cleanly in `updateNestedMetadataNode` inside `internal/parser/parser.go`.
- [x] Avoid unconditionally appending empty/false fields (scope, local_only, description) to root mapping in `internal/parser/parser.go` if they are not originally present.
- [x] Ensure all calls to `Parse` and `os.ReadFile` in `internal/parser/parser_test.go` do not discard errors.

---

## Phase 6: Judgment Day Round 3 Fixes (TDD)
- [x] Implement atomic write backup strategy in `internal/parser/parser.go` `Save()`
- [x] Populate computed fallback name when frontmatter name is empty in `ParseContentWithName`/`Parse`
- [x] Ensure optional empty scope/auto_invoke fields are not appended to formatting mapping node on save when not originally present
- [x] Define dynamic card style based on `m.Height` and use it in UI rendering
- [x] Skip/disable cursor navigation in `ScreenLicenseDisclosure`

---

## Phase 7: Judgment Day Round 4 Fixes (TDD)
- [x] Guarantee temp file deletion in `Save()` using a deferred call at the start.
- [x] Reset `Screen` to `ScreenInstaller` in `Update()` when user confirms license with "y"/"Y".
- [x] Allow "space" and " " keys (alongside "enter") to trigger the "Execute Install" action at `10+storageOffset`.


