# Contributing to Skillsync TUI

Thank you for your interest in improving Skillsync TUI!

## Getting Started

1. **Clone the Repo**: Ensure you have Go 1.21+ installed.
2. **Install Dependencies**: `go mod download`.
3. **Run the TUI**: `go run ./cmd/synck`.
4. **Explore the Docs**: Read `docs/architecture.md` to understand the layout.

## Development Install Workflow

If you have made local changes and want to test the full installation and skill ecosystem:

1. **Launch the TUI**: Run `go run ./cmd/synck`.
2. **Install Globally**: Run `go install ./cmd/synck` to update your global `synck` command with your local changes.
2. **Initialize Project**: If the project isn't initialized, the **Installer** screen will launch.
3. **Select Providers**: Toggle your desired AI providers (e.g., Claude, Gemini).
4. **Execute Install**: Trigger the installation action.
5. **Verify Output**: Confirm that `AGENTS.md`, `CLAUDE.md`, and other selected provider files are generated in the project root.

## Development Rules

- **Source Evidence**: When modifying logic, always provide source evidence (file and line) in your pull request.
- **SDD First**: For any new feature, start with a proposal and specs.
- **Test Everything**: Add unit tests for logic and `teatest` scenarios for UI changes.
  - Run the full suite: `go test ./...`.
  - Target UI/Storage tests: `go test ./internal/ui/...` and `go test ./internal/storage/...`.
  - Verify installer side effects by checking for generated files in temporary test directories.
- **No Direct Edits**: Do not modify generated files like `AGENTS.md` directly. Update the source `SKILL.md` or templates.

## Code Style

- We follow standard Go conventions.
- Use `rtk` to keep your commits clean of unnecessary AI context.
- Maintain the MVU separation in the UI layer.

## Feedback and Issues

If you find a bug or have a suggestion, please open an issue with detailed reproduction steps.
