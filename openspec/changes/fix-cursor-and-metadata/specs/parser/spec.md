# Spec: Metadata Preservation during Format

## Issue Description
In `internal/parser/parser.go`, when generating a YAML frontmatter block from scratch (i.e., when `yamlStr == ""` inside `Format`), the serializer only outputs the `scope` and `auto_invoke` fields, discarding other important fields like `name`, `description`, and `local_only`.
Furthermore, if a skill originally contains a nested `metadata` MappingNode structure, updates to `scope` and `auto_invoke` must be preserved within that nested structure rather than duplicating them at the root level, and all other metadata fields (e.g. `author`, `version`, `license`, `origin`) must be preserved exactly as they are.

## Requirements
- When calling `Format` on a skill structure with no existing frontmatter (`yamlStr == ""`):
  - The generated frontmatter MUST include the `name` field if `skill.Name` is not empty.
  - The generated frontmatter MUST include the `description` field if `skill.Metadata.Description` is not empty.
  - The generated frontmatter MUST include the `scope` field if `skill.Metadata.Scope` is not empty.
  - The generated frontmatter MUST include the `auto_invoke` field if `skill.Metadata.AutoInvoke` is not empty.
  - The generated frontmatter MUST include the `local_only` field if `skill.Metadata.LocalOnly` is true.
- When formatting a skill with an existing frontmatter that contains a nested `metadata` MappingNode (such as `metadata.scope` and `metadata.auto_invoke`):
  - The `Format` function MUST update `scope` and `auto_invoke` inside the nested `metadata` block if they are defined there.
  - The `Format` function MUST NOT create duplicate `scope` or `auto_invoke` keys at the root level if they are updated in the nested block.
  - All other fields in the frontmatter, including custom metadata fields (like `author`, `version`, `license`, `origin`), MUST be preserved exactly.

## Scenarios

### Scenario 1: Formatting a new skill from scratch
- **Given** a new `Skill` struct with the following values:
  - `Name` = `"my-new-skill"`
  - `Metadata.Description` = `"This is a description."`
  - `Metadata.Scope` = `"project"`
  - `Metadata.AutoInvoke` = `["trigger-1", "trigger-2"]`
  - `Metadata.LocalOnly` = `true`
- **When** `Format(skill)` is called and the skill does not have an existing path or `yamlStr` is empty
- **Then** the returned string MUST contain the following YAML fields in its frontmatter section:
  ```yaml
  name: my-new-skill
  description: This is a description.
  scope: project
  auto_invoke:
    - trigger-1
    - trigger-2
  local_only: true
  ```

### Scenario 2: Preserving nested metadata and custom fields
- **Given** a skill with an existing frontmatter containing nested metadata and custom attributes:
  ```yaml
  name: gentleman-bubbletea
  description: Bubbletea patterns
  license: Apache-2.0
  metadata:
    author: gentleman-programming
    version: "1.0"
    scope: [root]
    auto_invoke: "Working on TUI screens"
  ```
- **When** `skill.Metadata.Scope` is updated to `"ui, root"` and `Format(skill)` is called
- **Then** the updated frontmatter MUST preserve `name`, `description`, `license`, `metadata.author`, and `metadata.version` exactly.
- **And** the nested `metadata.scope` MUST be updated to `ui, root`.
- **And** no duplicate `scope` key MUST be added at the root level of the frontmatter.
