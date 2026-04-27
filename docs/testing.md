# Testing and Validation

We emphasize robust testing, especially for TUI components and core domain logic.

## Go Testing

Run all unit and integration tests using standard Go tools:
```bash
go test ./internal/...
```

## TUI Testing (Teatest)

Testing terminal user interfaces can be tricky. We use `teatest` to verify Bubbletea models.
- **Location**: Look for `_test.go` files in `internal/ui/`.
- **Approach**: We simulate key presses and verify that the model's state and view update correctly.

## Sandbox Validation

The sandbox logic is tested by attempting to break out of the directory boundaries.
- Tests in `internal/sandbox/sandbox_test.go` verify that path traversal attempts are caught and blocked.

## Continuous Verification

In the SDD workflow, the `verify` phase is mandatory. No change is considered "done" until it passes its corresponding specs and the existing test suite.
