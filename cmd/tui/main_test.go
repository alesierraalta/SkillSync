package main

import (
	"errors"
	"testing"
)

// TestMainUiRunError verifies that main() calls osExit(1) when uiRun() returns an error.
// This tests AUDIT-08b: error handling flows to osExit.
func TestMainUiRunError(t *testing.T) {
	// Save and restore the uiRun seam
	origUIRun := uiRun
	defer func() { uiRun = origUIRun }()

	// Save and restore the osExit seam
	var exitCodes []int
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	// Inject test versions
	uiRun = func() error {
		return errors.New("test error")
	}

	osExit = func(code int) {
		exitCodes = append(exitCodes, code)
	}

	// Call main()
	main()

	// Verify osExit was called with code 1
	if len(exitCodes) == 0 {
		t.Error("expected osExit to be called, but it was not")
	} else if len(exitCodes) > 1 {
		t.Errorf("expected osExit to be called once, but was called %d times", len(exitCodes))
	} else if exitCodes[0] != 1 {
		t.Errorf("expected osExit(1), got osExit(%d)", exitCodes[0])
	}
}

// TestMainUiRunNil verifies that main() does not call osExit when uiRun() returns nil.
// This tests AUDIT-08a: happy path flows without exit.
func TestMainUiRunNil(t *testing.T) {
	// Save and restore the uiRun seam
	origUIRun := uiRun
	defer func() { uiRun = origUIRun }()

	// Save and restore the osExit seam
	var exitCodes []int
	origOsExit := osExit
	defer func() { osExit = origOsExit }()

	// Inject test versions
	uiRun = func() error {
		return nil
	}

	osExit = func(code int) {
		exitCodes = append(exitCodes, code)
	}

	// Call main()
	main()

	// Verify osExit was not called
	if len(exitCodes) != 0 {
		t.Errorf("expected osExit not to be called, but was called %d times with codes %v", len(exitCodes), exitCodes)
	}
}

// TestMainExitCodeCapture verifies that exit codes are captured correctly.
// This tests the seam mechanism itself (AUDIT-08a/b).
func TestMainExitCodeCapture(t *testing.T) {
	tests := []struct {
		name          string
		uiRunErr      error
		wantExitCodes []int
	}{
		{
			name:          "error returns exit 1",
			uiRunErr:      errors.New("ui error"),
			wantExitCodes: []int{1},
		},
		{
			name:          "nil returns no exit",
			uiRunErr:      nil,
			wantExitCodes: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore the uiRun seam
			origUIRun := uiRun
			defer func() { uiRun = origUIRun }()

			// Save and restore the osExit seam
			var exitCodes []int
			origOsExit := osExit
			defer func() { osExit = origOsExit }()

			// Inject test versions
			uiRun = func() error {
				return tt.uiRunErr
			}

			osExit = func(code int) {
				exitCodes = append(exitCodes, code)
			}

			// Call main()
			main()

			// Verify exit codes match expected
			if len(exitCodes) != len(tt.wantExitCodes) {
				t.Errorf("expected %d exit calls, got %d", len(tt.wantExitCodes), len(exitCodes))
			}

			for i, code := range exitCodes {
				if i < len(tt.wantExitCodes) && code != tt.wantExitCodes[i] {
					t.Errorf("exit code at index %d: got %d, want %d", i, code, tt.wantExitCodes[i])
				}
			}
		})
	}
}
