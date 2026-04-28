package main

import (
	"testing"
	"skillsync/tui/internal/install"
)

func TestCLIFlags(t *testing.T) {
	// 1. Test -g flag triggers GlobalInstaller
	calledGlobal := false
	GlobalInstaller = func() install.Result {
		calledGlobal = true
		return install.Result{Success: true, Message: "Mocked install"}
	}
	// Mock TUI runner to avoid launching it
	TUIRunner = func() error {
		return nil
	}

	err := run([]string{"synck", "-g"})
	if err != nil {
		t.Fatalf("run(-g) failed: %v", err)
	}
	if !calledGlobal {
		t.Error("expected GlobalInstaller to be called when -g is present")
	}

	// 2. Test 'install -g' triggers GlobalInstaller
	calledGlobal = false
	err = run([]string{"synck", "install", "-g"})
	if err != nil {
		t.Fatalf("run(install -g) failed: %v", err)
	}
	if !calledGlobal {
		t.Error("expected GlobalInstaller to be called when 'install -g' is present")
	}

	// 3. Test no flags triggers TUIRunner
	calledTUI := false
	TUIRunner = func() error {
		calledTUI = true
		return nil
	}
	calledGlobal = false

	err = run([]string{"synck"})
	if err != nil {
		t.Fatalf("run() failed: %v", err)
	}
	if !calledTUI {
		t.Error("expected TUIRunner to be called when no flags are present")
	}
	if calledGlobal {
		t.Error("expected GlobalInstaller NOT to be called when no flags are present")
	}
}
