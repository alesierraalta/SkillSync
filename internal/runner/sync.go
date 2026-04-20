package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// SyncResult captures the output of the sync process
type SyncResult struct {
	Stdout   string
	Stderr   string
	Error    error
	ExitCode int
}

// Runner handles execution of shell scripts
const DefaultSyncPath = ".agents/skills/skill-sync/assets/sync.sh"

// Runner handles execution of shell scripts
type Runner struct {
	DefaultScriptPath string
}

// NewRunner creates a new instance of Runner
func NewRunner(defaultPath string) *Runner {
	return &Runner{DefaultScriptPath: defaultPath}
}

// ExecuteSync runs the sync script asynchronously
func (r *Runner) ExecuteSync(ctx context.Context, args []string) <-chan SyncResult {
	res := make(chan SyncResult, 1)

	go func() {
		if _, err := os.Stat(r.DefaultScriptPath); os.IsNotExist(err) {
			res <- SyncResult{
				Error:    fmt.Errorf("script not found at %s", r.DefaultScriptPath),
				ExitCode: 1,
			}
			return
		}

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			if strings.HasSuffix(r.DefaultScriptPath, ".ps1") {
				fullArgs := append([]string{"-ExecutionPolicy", "Bypass", "-File", r.DefaultScriptPath}, args...)
				cmd = exec.CommandContext(ctx, "powershell", fullArgs...)
			} else {
				// For .sh scripts on Windows, attempt to use sh if available
				// Otherwise try direct execution (which might fail if not associated)
				if _, err := exec.LookPath("sh"); err == nil {
					cmd = exec.CommandContext(ctx, "sh", append([]string{r.DefaultScriptPath}, args...)...)
				} else {
					cmd = exec.CommandContext(ctx, r.DefaultScriptPath, args...)
				}
			}
		} else {
			cmd = exec.CommandContext(ctx, r.DefaultScriptPath, args...)
		}

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		exitCode := 0
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			} else if ctx.Err() != nil {
				// Context cancellation/timeout
				exitCode = 124 // Common exit code for timeout
			} else {
				exitCode = 1
			}
		}

		res <- SyncResult{
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			Error:    err,
			ExitCode: exitCode,
		}
	}()

	return res
}
