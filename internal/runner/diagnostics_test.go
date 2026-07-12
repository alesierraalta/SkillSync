package runner

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestExecuteLegacyScriptSync_FileNotFound(t *testing.T) {
	// Use a path that definitely does not exist
	runner := NewLegacyScriptRunner("non_existent_script_12345.sh")
	ctx := context.Background()

	ch := runner.ExecuteLegacyScriptSync(ctx, nil)
	result := <-ch

	if result.ExitCode == 0 {
		t.Errorf("Expected non-zero exit code for missing file, got %d", result.ExitCode)
	}

	if result.Error == nil {
		t.Error("Expected error for missing file, but got nil")
	}

	// The current implementation returns a formatted error if os.Stat fails
	if !strings.Contains(result.Error.Error(), "script not found") {
		t.Errorf("Expected 'script not found' error, got: %v", result.Error)
	}
}

func TestExecuteLegacyScriptSync_CommandFailure(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a file that exists but will fail to execute as a shell script
	mockFile := tmpDir + "/test_fail.sh"
	err := os.WriteFile(mockFile, []byte("exit 1"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	runner := NewLegacyScriptRunner(mockFile)
	ctx := context.Background()

	ch := runner.ExecuteLegacyScriptSync(ctx, nil)
	result := <-ch

	// If it fails with Exit 1 and empty output, we want to see if we can improve diagnostics
	t.Logf("ExitCode: %d", result.ExitCode)
	t.Logf("Stdout: %q", result.Stdout)
	t.Logf("Stderr: %q", result.Stderr)
	t.Logf("Error: %v", result.Error)

	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code")
	}

	if result.Stderr == "" {
		t.Error("Expected non-empty Stderr for failure diagnostics")
	}
}

func TestExecuteLegacyScriptSync_MockCommandNotFound(t *testing.T) {
	// Simulate an executable that doesn't exist
	// (Skipping implementation as it's hard to trigger consistently)
}
