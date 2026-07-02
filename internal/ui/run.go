package ui

import (
	"fmt"
	"skillsync/tui/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
)

// runProgram is a seam for testing tea.Program.Run() calls.
// It can be swapped out in tests to inject errors or panic behavior.
var runProgram = func(m tea.Model) error {
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

// Run launches the TUI program
func Run() error {
	storageService := storage.NewService("")
	backend := NewBackend(storageService)

	// runWithPanicRecovery wraps the actual TUI run with panic recovery.
	err := func() (returnErr error) {
		defer func() {
			if r := recover(); r != nil {
				// Panic was caught and converted to error return
				// This ensures panics inside runProgram become errors, not crashes
				returnErr = fmt.Errorf("recovered panic: %v", r)
			}
		}()
		return runProgram(NewModel(backend))
	}()

	if err != nil {
		return fmt.Errorf("alas, there's been an error: %w", err)
	}
	return nil
}
