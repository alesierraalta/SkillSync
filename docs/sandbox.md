# Sandbox Security Model

The **Skill Sandbox** is a security layer that prevents skills from performing unintended or destructive actions on the host system.

## Isolation Principles

- **Directory Boundary**: Skills are restricted to the workspace directory.
- **Path Validation**: Every file operation (read, write, delete) is intercepted. Paths containing `..` or pointing outside the root are rejected.
- **Symlink Protection**: The sandbox rejects following symlinks that lead outside the project scope.

## Sandbox Tooling

The `internal/sandbox` package provides:
- **`SafeFs`**: A wrapper around the OS filesystem that enforces boundary checks.
- **`Runner`**: Executes external commands (like `git` or `go`) within a restricted environment.

## Validation Process

When a skill is selected for execution:
1. The sandbox environment is prepared.
2. The skill's `auto_invoke` triggers are verified.
3. Execution starts, and all tool calls are logged for auditing.
