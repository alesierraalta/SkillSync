package install

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// AutoskillsResult captures the outcome of an autoskills discovery run
type AutoskillsResult struct {
	Success bool
	Output  string
	Error   error
}

// AutoskillsPreflight verifies Node.js >= 22.6.0 and npx availability.
func AutoskillsPreflight() error {
	nodeOutput := func() (string, error) {
		out, err := exec.Command("node", "-v").Output()
		return string(out), err
	}
	npxLookup := func() error {
		_, err := exec.LookPath("npx")
		return err
	}
	return autoskillsPreflightWithMocks(nodeOutput, npxLookup)
}

func autoskillsPreflightWithMocks(nodeOutput func() (string, error), npxLookup func() error) error {
	// 1. Check Node.js version
	output, err := nodeOutput()
	if err != nil {
		return errors.New("Node.js not found")
	}

	versionStr := strings.TrimPrefix(strings.TrimSpace(output), "v")
	v, err := semver.NewVersion(versionStr)
	if err != nil {
		return fmt.Errorf("invalid Node.js version format: %s", versionStr)
	}

	constraint, _ := semver.NewConstraint(">= 22.6.0")
	if !constraint.Check(v) {
		return errors.New("Node.js >= 22.6.0 required")
	}

	// 2. Check npx
	if err := npxLookup(); err != nil {
		return errors.New("npx not found")
	}

	return nil
}

// AutoskillsInstall executes npx -y autoskills.
func AutoskillsInstall(ctx context.Context) AutoskillsResult {
	runner := func(ctx context.Context, name string, arg ...string) (string, error) {
		cmd := exec.CommandContext(ctx, name, arg...)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	return autoskillsInstallWithMocks(ctx, runner)
}

func autoskillsInstallWithMocks(ctx context.Context, runner func(context.Context, string, ...string) (string, error)) AutoskillsResult {
	output, err := runner(ctx, "npx", "-y", "autoskills")
	if err != nil {
		return AutoskillsResult{
			Success: false,
			Output:  output,
			Error:   err,
		}
	}

	return AutoskillsResult{
		Success: true,
		Output:  output,
		Error:   nil,
	}
}
