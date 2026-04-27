# TUI Architecture

The Skillsync TUI follows a clean, 3-tier architecture designed for maintainability, testability, and clear separation of concerns.

## 3-Tier Structure

1. **Entrypoint Layer (`cmd/`)**
   - `cmd/tui/main.go`: The main interactive application entrypoint. Initializes the Bubbletea model and starts the TUI loop.
   - `cmd/skill-sandbox/main.go`: A specialized CLI entrypoint for running skills in isolation.

2. **UI Layer (`internal/ui/`)**
   - Implements the **Model-View-Update (MVU)** pattern using the [Bubbletea](https://github.com/charmbracelet/bubbletea) framework.
   - `model.go`: Contains the application state, including the list of discovered skills and UI status.
   - `view.go`: Defines the visual layout and rendering logic using [Lipgloss](https://github.com/charmbracelet/lipgloss).
   - `update.go`: Handles all user interactions, keybindings, and asynchronous commands.

3. **Domain Logic Layer (`internal/`)**
   - `internal/discovery`: The engine that scans the filesystem for `SKILL.md` files across various hidden directories (`.agents`, `.claude`, etc.).
   - `internal/sandbox`: Manages process isolation and path validation to ensure skills run safely without escaping their boundaries.
   - `internal/ecosystem`: Coordinates skill installation, core bootstrapping, and metadata synchronization.

## Key Flows

### Skill Discovery
1. The `DiscoveryService` scans configured paths.
2. It parses YAML frontmatter from `SKILL.md` files.
3. The resulting `Skill` objects are loaded into the UI `Model`.

### Event Loop
1. User presses a key (e.g., `s` for sync).
2. `Update()` receives the message and triggers a `tea.Cmd`.
3. The command executes the underlying domain logic (e.g., calling the `sync.sh` script).
4. The result is returned to `Update()` as a message, updating the `Model` state.

## Persistence Model

- **Filesystem**: `SKILL.md` files are the primary source of truth for skill definitions.
- **Engram**: Used for long-term memory of decisions, bug fixes, and project-specific AI context.
- **AGENTS.md**: A human-readable registry generated and updated by the `skill-sync` workflow.
