# Changelog

## 2026-06-25 — tui-state-audit

First audit pass. Six area reports under `docs/audits/2026-06-25/`
and surgical quick wins across the observability, test, and build
layers.

**Quick wins**

- Structured logging in `cmd/tui/main.go` (stdlib `log/slog` with
  `slog.NewJSONHandler`).
- Panic recovery in `internal/ui/run.go`; recovered panics surface
  as returned errors so the process exits 1, not 134/139.
- `%v` → `%w` fix on `internal/ui/run.go` (worst error-wrapper
  offender).
- `internal/parser/parser_bench_test.go` — `BenchmarkParseContent`,
  `BenchmarkFormat`.
- `internal/sandbox/sandbox_bench_test.go` — `BenchmarkSimulateInstall`.
- `internal/parser/parser_fuzz_test.go` — `FuzzParseContent` (4-seed
  corpus; `-fuzztime=1s`).
- `t.Parallel()` added to 4 stateless test files
  (`internal/diff`, `internal/agentdetect/parsemcp`,
  `internal/syncengine/sync`, `internal/ui/agent_eco_status`).
- `cmd/synck/e2e_test.go` — in-process CLI e2e for the `version`
  subcommand.
- `openspec/config.yaml` `test_command` broadened from
  `go test ./cmd/synck -count=1` to `go test ./... -count=1`.
- `docs/testing.md` — new `## Benchmarks` section.
- `Makefile` (new) — `test` / `bench` / `lint` / `build` / `verify`
  targets; `lint` degrades to a no-op note when golangci-lint is
  absent.
- `.golangci.yml` (new) — v2 schema; `run.timeout: 5m`; the five
  required linters enabled.
