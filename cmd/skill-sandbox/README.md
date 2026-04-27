# SkillSync Sandbox

A dedicated developer tool to safely test skill installation, discovery, and parsing logic in isolation.

## Purpose

The sandbox allows you to:
1. Verify how `DiscoverSkills` behaves with different directory structures (including nested and duplicate cases).
2. Confirm `parser.Parse` correctly populates `Skill.ID` with normalized paths (the core of the recent fix).
3. Simulate skill installation scenarios without affecting your local `.agents` or `.opencode` directories.
4. Ensure cross-platform path consistency (Windows vs Unix style slashes).

## Usage

Run the sandbox tool from the project root:

```bash
go run ./cmd/skill-sandbox [flags]
```

### Flags
- `--keep-temp`: Keep the temporary directory and fixtures after the run. This is useful for manual inspection of the directory structure.
- `--json`: Output a structured JSON document instead of human-readable text. This is preferred for automated regression checks.

## Examples

### Human-readable output (Default)
```bash
go run ./cmd/skill-sandbox
```

### JSON output for automation
```bash
go run ./cmd/skill-sandbox --json
```

Output format:
```json
{
  "temp_root": "/tmp/skill-sandbox-123",
  "keep_temp": false,
  "discovered_paths": [...],
  "skills": [
    {
      "id": "deploy",
      "name": "deploy",
      "path": "...",
      "raw_body": "...",
      "description": "..."
    }
  ],
  "install_simulation": {
    "success": true,
    "discovered_after_install": true
  },
  "bug_repro_result": "OK"
}
```

- The tool creates a temporary directory using `os.MkdirTemp`.
- It populates it with a predefined set of "fixtures" (mock skills).
- It runs the actual discovery and parser logic on these fixtures.
- It prints a table of the resulting `Skill` objects (ID, Name, Description).
- It performs a simple installation simulation to ensure the installer pattern still works with the discovery logic.
- It cleans up the temporary directory after completion.

## Test Scenarios Covered

| Case | Scenario | Expected Outcome |
|------|----------|------------------|
| **Normal** | Single skill in `.agents/skills` | Discovered and parsed correctly. |
| **Collision** | Skills with same folder name in different providers | Both discovered with unique `ID`s. |
| **Nested Duplicate** | `SKILL.md` inside a folder that already has one | Nested one is skipped by `Discovery`. |
| **Bug Repro** | Skill with "Contenido base." placeholder | Parsed as a skill with literal content. |
| **Deep Folder** | `SKILL.md` in a deep subdirectory | Discovered correctly. |
| **Unsupported** | `SKILL.md` in a directory not in the provider list | Ignored by `Discovery`. |
| **Path Normalization** | ID populated on Windows | `ID` uses forward slashes for consistency. |

## Limitations

- This tool does not launch the full Bubbletea UI. For UI testing, use `internal/ui/update_test.go` or manual verification.
- Virtual skills (like `AGENTS.md`) are handled in the UI layer and are not part of the discovery/parsing sandbox.
