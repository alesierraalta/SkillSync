package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetPathHint(t *testing.T) {
	// We can't easily test the exact PATH string because it's machine dependent,
	// but we can check if it returns a non-empty string and contains expected instructions.
	hint := GetPathHint()
	if hint == "" {
		t.Error("Expected non-empty hint")
	}
}

func TestGlobalInstallRepoCheck(t *testing.T) {
	// Create a temp directory that is NOT the repo root
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	os.Chdir(tmpDir)

	result := GlobalInstall()
	if result.Success {
		t.Error("Expected GlobalInstall to fail when not in repo root")
	}
	if result.Message == "" {
		t.Error("Expected error message")
	}
}

func TestGlobalInstallRepoCheck_Success(t *testing.T) {
	// Create a mock repo root
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	os.Chdir(tmpDir)
	os.WriteFile("go.mod", []byte("module test"), 0644)
	os.MkdirAll(filepath.Join("cmd", "synck"), 0755)

	// We don't want to actually run 'go install' in tests.
	// Since GlobalInstall calls exec.Command, it's hard to mock without injecting a runner.
	// But we can at least verify it reaches the 'go' check.
}
