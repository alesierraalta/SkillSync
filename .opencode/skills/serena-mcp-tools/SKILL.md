---
name: serena-mcp-tools
description: >
  Serena MCP and tool-selection discipline for high-efficiency codebase navigation,
  symbolic editing, and complex task delegation.
metadata:
  author: a.sierra
  version: "1.1"
  scope: [root]
  auto_invoke: "Using Serena MCP or tool-heavy work"
allowed-tools: Read, Edit, Write, Glob, Grep, Bash, Task, WebFetch, Serena MCP, Engram, Context7, Proxima, delegate
---

# Serena MCP Tools & Delegation

## When to Use

Use this skill when:
- Navigating large codebases where symbols and references are more efficient than grep.
- Performing safe, targeted edits using symbolic tools (replace body, insert before/after).
- Selecting the optimal tool (cost vs precision) for search and read operations.
- Coordinating multi-stage tasks involving memory (Engram), documentation (Context7), and research (Proxima).
- Delegating autonomous units of work to sub-agents (especially for implementation or deep research).

## Critical Patterns

### 1. Serena-first Code Navigation
- **Bootstrap**: Call `serena_initial_instructions` once. Use `serena_activate_project` and `serena_check_onboarding_performed`.
- **Structural Discovery**: Use `serena_get_symbols_overview` to understand a file's "map" without reading every line.
- **Targeted Reading**: Use `serena_find_symbol` to jump to definitions. Only read the body if you need implementation details.
- **Impact Analysis**: Use `serena_find_referencing_symbols` before making changes to public APIs or common utilities.

### 2. Tool Selection Discipline
| Goal | Strategy | Tool |
| ---- | -------- | ---- |
| Find file by name | Precise path/pattern | `glob` or `serena_find_file` |
| Text search | Quick, flat search | `grep` (plain text) |
| Structural search | Search for symbols/kinds | `serena_search_for_pattern` |
| Read code | Full file (small) | `read` |
| Read code | Symbols/Methods | `serena_read_file` with ranges or symbolic tools |
| CLI operations | Native shell | `bash` (preferred for standard CLI) |
| Safe shell | Managed env | `serena_execute_shell_command` |

### 3. Delegation & Sub-agents
For complex tasks, leverage sub-agents as requested by the user:
- **Task Tool**: Use for multi-step, open-ended searches or complex implementations where you need to track state.
- **Delegate**: Use for autonomous research or implementation tasks that can run in the background.
- **Protocol**: Provide the sub-agent with a clear goal, the necessary tool context (e.g., "use Serena for navigation"), and the artifact persistence mode (usually Engram).

### 4. Memory & Persistence (Engram Mandatory)
Save discoveries IMMEDIATELY to `mem_save` after:
- Finding a non-obvious code path or logic quirk.
- Making an architectural decision or selecting a library.
- Fixing a bug (record root cause).
- Establishing a new pattern or convention.
- Learning a user preference or project constraint.

## Rules

- **Precision over Volume**: Never read 1000 lines when 50 lines of a symbol body suffice.
- **Safe Edits**: Use symbolic insertion (`serena_insert_after_symbol`) for new functions/types to avoid line-count drift.
- **Context Awareness**: Always check `serena_check_onboarding_performed` to ensure you aren't missing project-specific knowledge.
- **No Shortcuts**: Use Context7 for library docs instead of guessing or searching generic web content.
- **Mirroring**: Keep one source of truth: update references, registries, and mirrored skill files through the project sync flow.

## Commands

```bash
# Register and mirror all skills
go run ./cmd/synck sync
```
