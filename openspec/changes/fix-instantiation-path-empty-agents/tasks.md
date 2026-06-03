# Tasks

- [x] Modify interface and implementation of `EnsureAgentsMD` in `internal/ui/backend.go` to accept `root string`.
- [x] Update mock implementations in `ui_tdd_test.go` and `ecosystem_tdd_test.go` to accept `root string`.
- [x] Modify ecosystem.go case 0.8 and `instantiateEcosystemCmd` to use `rootPath` and execute skill discovery and markdown updates before copying `AGENTS.md` to assistant markdown instruction files.
- [x] Update test calls in `ecosystem_test.go` and `update.go`.
- [x] VERIFY: run `go test ./internal/ui/...` to ensure all tests are green.
