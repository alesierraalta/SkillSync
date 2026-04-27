# Skillsync TUI

A powerful terminal user interface and management system for AI agent skills.

## Overview

Skillsync TUI is the central hub for discovering, managing, and testing skills across your development projects. It provides a robust, sandboxed environment for AI agents to interact with your codebase safely and effectively.

## 🚀 Getting Started

1. **Prerequisites**: [Go](https://golang.org/dl/) (1.21+)
2. **Launch**: `go run ./cmd/tui`
3. **Explore**: Use `j/k` to navigate and `enter` to select a skill.

## 📚 Documentation Index

Learn more about the project through our detailed documentation:

- **[Project Intro](docs/intro.md)**: Vision and high-level workflows.
- **[Architecture](docs/architecture.md)**: 3-tier structure and internal design.
- **[Commands Reference](docs/commands.md)**: CLI flags and TUI keybindings.
- **[The Skills System](docs/skills-system.md)**: Discovery, `SKILL.md` format, and registries.
- **[Standard Workflows](docs/workflows.md)**: SDD cycle, `rtk`, and sync processes.
- **[Sandbox Security](docs/sandbox.md)**: Path validation and isolation model.
- **[Testing and Validation](docs/testing.md)**: `go test` and TUI testing patterns.
- **[Troubleshooting](docs/troubleshooting.md)**: Common issues and solutions.
- **[Contributor Notes](docs/contributing.md)**: Guidelines for developers.

## 🛠 Project Structure

- `cmd/`: Application entrypoints.
- `internal/`: Core domain logic (UI, Discovery, Sandbox, Ecosystem).
- `docs/`: Comprehensive project documentation.
- `scripts/`: Utility scripts (sync, setup).

## 📄 Registry

See **[AGENTS.md](AGENTS.md)** for a full list of currently available skills and their auto-invoke rules.
