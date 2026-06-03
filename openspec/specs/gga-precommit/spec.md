# gga-precommit Specification

## Purpose

GGA-based pre-commit hook that auto-reviews staged Go code against project standards using the opencode AI provider. Blocks commits on violations, passes clean code through.

## Requirements

### Configuration

| ID | Requirement | Strength |
|----|-------------|----------|
| CFG-01 | `.gga` MUST set `PROVIDER=opencode`, `MODEL=opencode/deepseek-v4-flash-free` | MUST |
| CFG-02 | `.gga` MUST set `RULES_FILE=.gga-rules.md` for standalone rules (not AGENTS.md) | MUST |
| CFG-03 | `.gga` SHOULD set `FILE_PATTERNS=*.go,*.md,*.mod,*.sum` | SHOULD |
| CFG-04 | `.gga` SHOULD set `STRICT_MODE=true` | SHOULD |
| CFG-05 | `.gga` MAY set `EXCLUDE_PATTERNS=*_test.go` | MAY |

### Rules File

| ID | Requirement | Strength |
|----|-------------|----------|
| RULES-01 | `.gga-rules.md` MUST exist with coding standards consumable by the AI reviewer | MUST |
| RULES-02 | MUST cover Go conventions: `gofumpt` formatting, CamelCase exports / camelCase privates, `if err != nil` handling, table-driven tests with `t.Parallel()` | MUST |
| RULES-03 | MUST cover Bubbletea MVU: model/update/view separation, `tea.Cmd` for side effects, `tea.Msg` as event protocol | MUST |
| RULES-04 | MUST cover project conventions: import grouping (std / third-party / internal), no `log.Fatal` in lib packages, clean architecture layer isolation | MUST |

### Hook Behavior

| ID | Requirement | Strength |
|----|-------------|----------|
| HOOK-01 | Pre-commit hook MUST invoke `gga run` on every `git commit` | MUST |
| HOOK-02 | On review FAILED, commit MUST be blocked with feedback displayed | MUST |
| HOOK-03 | Hook MUST NOT modify any project files (AGENTS.md, source, config) | MUST |

### Non-Functional

| ID | Requirement | Strength |
|----|-------------|----------|
| NFR-01 | Review timeout SHOULD be 300 seconds | SHOULD |
| NFR-02 | Hook MUST add ≤200ms overhead when no staged files | MUST |
| NFR-03 | `.gga-rules.md` SHOULD be under 200 lines | SHOULD |
| NFR-04 | Hook MUST work on Windows Git Bash | MUST |

## Scenarios

### S-01: Clean commit passes

- GIVEN a staged Go file following all project standards
- WHEN `git commit` is executed
- THEN GGA runs review and reports PASSED
- AND the commit proceeds

### S-02: Naming violation blocks commit

- GIVEN a staged Go file with a function name violating conventions
- WHEN `git commit` is executed
- THEN GGA runs review and reports FAILED
- AND the commit is blocked with violation feedback

### S-03: No staged files skips review

- GIVEN no staged files
- WHEN `git commit` is executed
- THEN GGA skips review (nothing to review)
- AND the commit proceeds

### S-04: Provider unavailable bypass

- GIVEN the opencode provider is unavailable
- WHEN developer commits with `git commit --no-verify`
- THEN the hook is bypassed and commit proceeds

### S-05: First commit on empty repo

- GIVEN no previous commits
- WHEN initial files are staged and `git commit` is executed
- THEN GGA reviews staged files
- AND commit proceeds if review passes

## File Specifications

### `.gga` (config)

Bash-sourced `KEY="VALUE"` file:
- `PROVIDER="opencode"`, `MODEL="opencode/deepseek-v4-flash-free"`
- `RULES_FILE=".gga-rules.md"`
- `FILE_PATTERNS="*.go,*.md,*.mod,*.sum"`
- `EXCLUDE_PATTERNS="*_test.go"`
- `STRICT_MODE="true"`, `TIMEOUT="300"`

### `.gga-rules.md` (rules)

Markdown with sections: Go conventions (gofumpt, naming, error handling, table-driven tests), Bubbletea MVU patterns (model/update/view, tea.Cmd, tea.Msg), project conventions (import groups, clean architecture layers, no log.Fatal in lib packages, commit message format).
