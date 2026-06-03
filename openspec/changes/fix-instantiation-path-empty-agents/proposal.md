# Proposal: Fix Instantiation Path and Empty AGENTS.md/Assistant Config Files

## Goal

Ensure `AGENTS.md` and related assistant configuration/instruction files (`CLAUDE.md`, `GEMINI.md`, `OPENCODE.md`) are created in the project root path rather than the process current working directory (which can be a subdirectory like `internal/ui` during tests or nested runs), and ensure they are populated with the installed skills during instantiation rather than being left empty.

## Scope

- Modify `EnsureAgentsMD` to accept a `root string` directory parameter and resolve the path dynamically using `filepath.Join(root, "AGENTS.md")`.
- Update TUI installer step `0.8` (Sync Configs) and `instantiateEcosystemCmd` to discover skills and update `AGENTS.md` directly using `syncengine` utilities before copying it to `CLAUDE.md`, `GEMINI.md`, and `OPENCODE.md`.
- Ensure all assistant config files are written in the correct `rootPath` instead of current working directory.
- Update mocks and tests to comply with the updated interface of `EnsureAgentsMD`.

## Non-goals

- Automating git operations or commits.
