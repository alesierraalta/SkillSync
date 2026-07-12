package runner

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestExecuteLegacyScriptSync_Success(t *testing.T) {
	tmpDir := t.TempDir()
	var mockScript string

	if runtime.GOOS == "windows" {
		mockScript = filepath.Join(tmpDir, "mock.ps1")
		os.WriteFile(mockScript, []byte("Write-Output 'hello'; Write-Error 'world'"), 0644)
	} else {
		mockScript = filepath.Join(tmpDir, "mock.sh")
		os.WriteFile(mockScript, []byte("#!/bin/sh\necho 'hello'\necho 'world' >&2"), 0755)
	}

	runner := NewLegacyScriptRunner(mockScript)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch := runner.ExecuteLegacyScriptSync(ctx, nil)
	result := <-ch

	if strings.TrimSpace(result.Stdout) != "hello" {
		t.Errorf("Expected stdout 'hello', got %q", result.Stdout)
	}
	// PowerShell Write-Error goes to stderr but might be noisier.
	// Basic check.
	if !strings.Contains(result.Stderr, "world") {
		t.Errorf("Expected stderr contains 'world', got %q", result.Stderr)
	}
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
}

func TestExecuteLegacyScriptSync_Timeout(t *testing.T) {
	tmpDir := t.TempDir()
	var mockScript string
	if runtime.GOOS == "windows" {
		mockScript = filepath.Join(tmpDir, "sleep.ps1")
		os.WriteFile(mockScript, []byte("Start-Sleep -Seconds 5"), 0644)
	} else {
		mockScript = filepath.Join(tmpDir, "sleep.sh")
		os.WriteFile(mockScript, []byte("#!/bin/sh\nsleep 5"), 0755)
	}

	runner := NewLegacyScriptRunner(mockScript)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ch := runner.ExecuteLegacyScriptSync(ctx, nil)
	result := <-ch

	if result.Error == nil && result.ExitCode == 0 {
		t.Error("Expected error or non-zero exit code due to timeout")
	}
}
