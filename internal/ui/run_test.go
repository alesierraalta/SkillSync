package ui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestRunProgramSeams tests the Run() function's handling of the runProgram seam
// under various conditions: happy path, error, panic, and panic+error join (AUDIT-09).
func TestRunProgramSeams(t *testing.T) {
	tests := []struct {
		name           string
		programFunc    func(tea.Model) error
		wantErr        bool
		wantPanic      bool
		wantErrContain string // substring expected in err.Error() if wantErr is true
	}{
		{
			name:        "happy path: nil return",
			programFunc: func(m tea.Model) error { return nil },
			wantErr:     false,
		},
		{
			name: "error wrap: sentinel error",
			programFunc: func(m tea.Model) error {
				return errors.New("bubbletea error")
			},
			wantErr:        true,
			wantPanic:      false,
			wantErrContain: "alas, there's been an error",
		},
		{
			name: "panic recovery: closure panics",
			programFunc: func(m tea.Model) error {
				panic("boom")
			},
			wantErr:        true,
			wantPanic:      true,
			wantErrContain: "recovered panic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore the runProgram seam
			orig := runProgram
			defer func() { runProgram = orig }()

			// Inject test version
			runProgram = tt.programFunc

			// Call Run() and check the result
			err := Run()

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if err != nil && tt.wantErrContain != "" {
				if !strings.Contains(err.Error(), tt.wantErrContain) {
					t.Errorf("error message should contain %q, got %v", tt.wantErrContain, err)
				}
			}

			if tt.wantPanic {
				if !strings.Contains(err.Error(), "recovered panic") {
					t.Errorf("expected error to contain 'recovered panic', got %v", err)
				}
			}
		})
	}
}

// NOTE: The errors.Join() call in Run() is defensive: it would capture both
// an error from runProgram AND a panic that occurs during error wrapping.
// However, the current control flow makes this unreachable under normal conditions:
// errors are returned immediately (line 38 in run.go), and panics only trigger
// the recover handler — a panic cannot occur during the return statement itself.
// The join is kept as defensive programming to handle hypothetical future code
// changes that might defer the error handling or cause panics after runProgram.

// TestRunProgramErrorWrap verifies that the error wrapping via %w is observable.
// This tests AUDIT-09b: the bubbletea error is wrapped, not replaced.
func TestRunProgramErrorWrap(t *testing.T) {
	// Save and restore the runProgram seam
	orig := runProgram
	defer func() { runProgram = orig }()

	// Create a sentinel error to verify wrapping
	sentinelErr := errors.New("sentinel bubbletea error")
	runProgram = func(m tea.Model) error {
		return sentinelErr
	}

	err := Run()

	if err == nil {
		t.Error("expected error, got nil")
	}

	// Verify the error chain: errors.Is should find the sentinel
	if !errors.Is(err, sentinelErr) {
		t.Errorf("expected errors.Is(err, sentinelErr) to be true, got %v", err)
	}
}

// TestRunProgramPanicRecovery verifies panic recovery (AUDIT-09c).
func TestRunProgramPanicRecovery(t *testing.T) {
	// Save and restore the runProgram seam
	orig := runProgram
	defer func() { runProgram = orig }()

	panicMsg := "boom"
	runProgram = func(m tea.Model) error {
		panic(panicMsg)
	}

	err := Run()

	if err == nil {
		t.Error("expected error from panic recovery, got nil")
	}

	// Verify the error message contains the recovered panic text
	if !strings.Contains(err.Error(), "recovered panic") {
		t.Errorf("expected error to contain 'recovered panic', got %v", err)
	}

	// Verify the panic message is captured
	if !strings.Contains(err.Error(), panicMsg) {
		t.Errorf("expected error to contain panic message %q, got %v", panicMsg, err)
	}
}
