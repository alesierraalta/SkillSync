# Proposal: Fix Cursor and Metadata Bugs in Skillsync TUI

## Intent
Resolve two distinct bugs:
1. A duplicate cursor rendering issue in `installer_model.go` under the installers view.
2. Metadata loss during raw frontmatter formatting from scratch in `parser.Format`.

## Scope
### In Scope
- Modify `internal/ui/installer_model.go` to fix the index check for rendering the cursor on the `[ Execute Install ]` button.
- Modify `internal/parser/parser.go` in the `Format` function to correctly serialize `name`, `description`, and `local_only` fields when creating the frontmatter from scratch (i.e. when `yamlStr == ""`).
- Add tests to ensure metadata fields are properly serialized on new/scratch frontmatter creation.

### Out of Scope
- Redesigning the TUI installer model or rendering logic.
- Editing or modifying other parser behaviors unrelated to YAML frontmatter serialization or loading.

## Capabilities
### New Capabilities
- None.

### Modified Capabilities
- **TUI Cursor Rendering**: The navigation cursor correctly highlights the active option ("Add shell aliases to profile" vs "[ Execute Install ]") without showing duplicate indicators.
- **Skill parser serialization**: When saving/formatting a newly created skill structure, the resulting frontmatter correctly contains all metadata attributes (`name`, `description`, `scope`, `auto_invoke`, and `local_only`).

## Approach
1. **Cursor Bug Fix**:
   - In `internal/ui/installer_model.go` (line 266), update the condition `if m.Cursor == 9+storageOffset` to check against `10+storageOffset`. This ensures that the navigation pointer correctly targets index `10+storageOffset` for the `[ Execute Install ]` action button, rather than duplicate-triggering on index `9+storageOffset`.

2. **Metadata Serialization Bug Fix**:
   - In `internal/parser/parser.go`'s `Format` function (around line 153):
     ```go
     if yamlStr == "" {
         // No existing file or no frontmatter, create minimal YAML
         // Include name, description, and local_only alongside scope and auto_invoke
     }
     ```
   - We will write a proper YAML serialization when generating the frontmatter from scratch. This can be accomplished by structuring a temporary map or struct, marshaling it with `yaml.Marshal`, and prepending it, or by generating it cleanly line-by-line:
     ```go
     var lines []string
     if skill.Name != "" {
         lines = append(lines, fmt.Sprintf("name: %s", skill.Name))
     }
     if skill.Metadata.Description != "" {
         lines = append(lines, fmt.Sprintf("description: %s", skill.Metadata.Description))
     }
     if skill.Metadata.Scope != "" {
         lines = append(lines, fmt.Sprintf("scope: %s", skill.Metadata.Scope))
     }
     if len(skill.Metadata.AutoInvoke) > 0 {
         var autoInvokeLines []string
         for _, trigger := range skill.Metadata.AutoInvoke {
             autoInvokeLines = append(autoInvokeLines, fmt.Sprintf("  - %s", trigger))
         }
         lines = append(lines, "auto_invoke:\n" + strings.Join(autoInvokeLines, "\n"))
     }
     if skill.Metadata.LocalOnly {
         lines = append(lines, "local_only: true")
     }
     
     yamlSection := "---\n" + strings.Join(lines, "\n") + "\n---\n"
     return skill.Prefix + yamlSection + skill.RawBody, nil
     ```

## Affected Areas
- `internal/ui/installer_model.go`: TUI cursor index mapping logic.
- `internal/parser/parser.go`: Frontmatter creation fallback block inside `Format()`.
- `internal/parser/parser_test.go`: Added test cases for formatting new/scratch frontmatter.

## Risks
- **Low Risk**: The cursor index alignment change is isolated to rendering in `installer_model.go` and has no impact on business logic.
- **Low Risk**: Frontmatter serialization additions are only triggered when formatting a skill with no existing frontmatter. Existing frontmatter editing relies on `yaml.Node` traversal which is untouched.

## Rollback Plan
- Revert code edits in `internal/ui/installer_model.go` and `internal/parser/parser.go` via Git (`git checkout`).

## Dependencies
- Standard library dependencies (`gopkg.in/yaml.v3` for parsing).

## Success Criteria
- No duplicate cursor shown on the install screen when navigating between the profile aliases option and the execute install button.
- Saving/formatting a skill metadata struct from scratch creates a YAML frontmatter containing `name`, `description`, `scope`, `auto_invoke`, and `local_only` (when set).
- `go test ./internal/ui/...` and `go test ./internal/parser/...` pass successfully.
