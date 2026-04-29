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
| `s` | **Save to Storage**: Save the highlighted skill to the global skill storage (`~/.skillsync/storage`). |
| `S` | **Sync**: Manually trigger the `skill-sync` workflow (updates `AGENTS.md`). |
| `?` | Toggle the Help overlay. |
| `esc` | Go back or cancel the current operation. |
| `q` / `ctrl+c` | Exit the application. |

## CLI: Skill Sandbox

The `skill-sandbox` tool allows you to test skills in isolation without launching the full TUI.

### Usage

```bash
./bin/skill-sandbox [flags]
```

## TUI Installer

When running `synck` in a project that hasn't been initialized, the TUI Installer screen will appear.

### Workflow

1. **Select Providers**: Use `space` or `enter` to toggle the AI providers you want to support (Claude, Gemini, Codex, Copilot, OpenCode).
2. **Execute Install**: Navigate to the "Install" or "Finish" action and press `enter`.
3. **Verification**: The installer will generate/sync the following files:
   - `AGENTS.md`: The main skill registry.
   - `CLAUDE.md`: Claude-specific instructions.
   - `codex.md`: Codex/Copilot configuration.
   - `.github/copilot-instructions.md`: GitHub Copilot specific docs.
   - `OPENCODE.md`: OpenCode configuration.bash
./bin/skill-sandbox [flags]

```

### Flags
- `--skill <name>`: Specify the skill to execute.
- `--input <path>`: Provide an input file or context for the skill.
- `--verbose`: Enable detailed logging of the execution and path validation.

## Core Workflows

### Manual Sync
If you've manually edited a `SKILL.md` file, you can run `./scripts/sync.sh` directly to update the project registry.

