# Contributing to Skillsync TUI

Thank you for your interest in improving Skillsync TUI!

## Getting Started

1. **Clone the Repo**: Ensure you have Go 1.21+ installed.
2. **Install Dependencies**: `go mod download`.
3. **Run the TUI**: `go run ./cmd/tui`.
4. **Explore the Docs**: Read `docs/architecture.md` to understand the layout.

## Development Rules

- **Source Evidence**: When modifying logic, always provide source evidence (file and line) in your pull request.
- **SDD First**: For any new feature, start with a proposal and specs.
- **Test Everything**: Add unit tests for logic and `teatest` scenarios for UI changes.
- **No Direct Edits**: Do not modify generated files like `AGENTS.md` directly. Update the source `SKILL.md` or templates.

## Code Style

- We follow standard Go conventions.
- Use `rtk` to keep your commits clean of unnecessary AI context.
- Maintain the MVU separation in the UI layer.

## Feedback and Issues

If you find a bug or have a suggestion, please open an issue with detailed reproduction steps.
