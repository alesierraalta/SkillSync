package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"fmt"

)

// Run launches the TUI program
func Run() error {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("alas, there's been an error: %v", err)
	}
	return nil
}
