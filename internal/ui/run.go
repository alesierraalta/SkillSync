package ui

import (
	"fmt"
	"skillsync/tui/internal/storage"

	tea "github.com/charmbracelet/bubbletea"
)

// Run launches the TUI program
func Run() error {
	storageService := storage.NewService("")
	backend := NewBackend(storageService)
	p := tea.NewProgram(NewModel(backend), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("alas, there's been an error: %v", err)
	}
	return nil
}
