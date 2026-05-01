package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

		absPath, err := filepath.Abs(r.DefaultScriptPath)
		if err != nil {
			res <- SyncResult{
				Error:    fmt.Errorf("failed to get absolute path for %s: %w", r.DefaultScriptPath, err),
				ExitCode: 1,
			}
			return
		}
		scriptDir := filepath.Dir(absPath)
		scriptBase := filepath.Base(absPath)

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			if strings.HasSuffix(absPath, ".ps1") {
				fullArgs := append([]string{"-ExecutionPolicy", "Bypass", "-File", absPath}, args...)
				cmd = exec.CommandContext(ctx, "powershell", fullArgs...)
			} else {
				// For .sh scripts on Windows, attempt to use sh if available, then bash
				// Using ./scriptBase with cmd.Dir is most compatible with shells that don't like C:/ paths
				if _, err := exec.LookPath("sh"); err == nil {
					cmd = exec.CommandContext(ctx, "sh", append([]string{"./" + scriptBase}, args...)...)
				} else if _, err := exec.LookPath("bash"); err == nil {
					cmd = exec.CommandContext(ctx, "bash", append([]string{"./" + scriptBase}, args...)...)
				} else {
					cmd = exec.CommandContext(ctx, absPath, args...)
				}
			}
		} else {
			cmd = exec.CommandContext(ctx, "./"+scriptBase, args...)
		}
		cmd.Dir = scriptDir

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		exitCode := 0
		stderrStr := stderr.String()

		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			} else if ctx.Err() != nil {
				// Context cancellation/timeout
				exitCode = 124 // Common exit code for timeout
			} else {
				exitCode = 1
			}

			// If we have an error but stderr is empty, it's likely a command-not-found 
			// or execution permission error. Populate stderr with the error message.
			if stderrStr == "" {
				stderrStr = err.Error()
			}
		}

		res <- SyncResult{
			Stdout:   stdout.String(),
			Stderr:   stderrStr,
			Error:    err,
			ExitCode: exitCode,
		}

	}()

	return res
}
