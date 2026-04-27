# Commands and Keybindings

Skillsync TUI provides both a graphical terminal interface and specialized CLI tools for developers.

## TUI Keybindings

When running the main `tui` application, use the following shortcuts:

| Key | Action |
| --- | --- |
| `up` / `down` | Navigate through the list of discovered skills. |
| `j` / `k` | Alternative navigation (Vim-style). |
| `enter` | Select a skill to view details or trigger its primary action. |
| `tab` | Switch focus between the Skill List and the Preview/Log panel. |
| `s` | **Sync**: Manually trigger the `skill-sync` workflow (updates `AGENTS.md`). |
| `?` | Toggle the Help overlay. |
| `esc` | Go back or cancel the current operation. |
| `q` / `ctrl+c` | Exit the application. |

## CLI: Skill Sandbox

The `skill-sandbox` tool allows you to test skills in isolation without launching the full TUI.

### Usage
```bash
./bin/skill-sandbox [flags]
```

### Flags
- `--skill <name>`: Specify the skill to execute.
- `--input <path>`: Provide an input file or context for the skill.
- `--verbose`: Enable detailed logging of the execution and path validation.

## Core Workflows

### Manual Sync
If you've manually edited a `SKILL.md` file, you can run `./scripts/sync.sh` directly to update the project registry.

### RTK (Rust Token Killer)
- `rtk init`: Initialize the token saving layer.
- `rtk save`: Manually trigger context compaction if supported.
