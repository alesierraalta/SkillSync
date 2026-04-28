package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Options holds installation preferences
type Options struct {
	Global bool
}

// Result captures the outcome of an installation step
type Result struct {
	Success bool
	Message string
	Error   error
}

// GlobalInstall runs 'go install ./cmd/synck' from the repo root
func GlobalInstall() Result {
	// 1. Verify we are in the repo root (look for go.mod and cmd/synck)
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return Result{Success: false, Message: "Not in repo root (go.mod missing)", Error: err}
	}
	if _, err := os.Stat(filepath.Join("cmd", "synck")); os.IsNotExist(err) {
		return Result{Success: false, Message: "cmd/synck missing - run from repo root", Error: err}
	}

	// 2. Check for 'go' command
	if _, err := exec.LookPath("go"); err != nil {
		return Result{Success: false, Message: "'go' command not found in PATH", Error: err}
	}

	// 3. Execute go install
	cmd := exec.Command("go", "install", "./cmd/synck")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return Result{Success: false, Message: "Failed to execute 'go install ./cmd/synck'", Error: err}
	}

	// 4. Provide PATH hints
	hint := GetPathHint()
	msg := fmt.Sprintf("Successfully installed 'synck' globally!\n\n%s", hint)

	return Result{Success: true, Message: msg}
}

// GetPathHint returns instructions on how to add Go bin to PATH if needed
func GetPathHint() string {
	goBin := ""
	if goBinEnv := os.Getenv("GOBIN"); goBinEnv != "" {
		goBin = goBinEnv
	} else {
		userHome, _ := os.UserHomeDir()
		if runtime.GOOS == "windows" {
			goBin = filepath.Join(userHome, "go", "bin")
		} else {
			goBin = filepath.Join(userHome, "go", "bin")
		}
	}

	pathEnv := os.Getenv("PATH")
	if !strings.Contains(pathEnv, goBin) {
		if runtime.GOOS == "windows" {
			return fmt.Sprintf("HINT: Ensure %s is in your PATH.\nTo add it, run in PowerShell:\n[System.Environment]::SetEnvironmentVariable(\"PATH\", $env:PATH + \";%s\", \"User\")", goBin, goBin)
		}
		return fmt.Sprintf("HINT: Ensure %s is in your PATH.\nAdd this to your shell profile:\nexport PATH=$PATH:%s", goBin, goBin)
	}

	return "You can now run 'synck' from any project."
}
