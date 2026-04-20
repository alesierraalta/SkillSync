**What**: Modified `detailView` in `internal/ui/view.go` to use compact layout and added `> ` indicator for focused inputs.
**Why**: Improve layout density and input visibility in TUI.
**Where**: `internal/ui/view.go`
**Learned**: Golden files must be updated after UI changes. Using `UPDATE_GOLDEN=true` environment variable via `powershell -Command` is necessary in this environment.