package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"skillsync/tui/internal/storage"
	"testing"
)

func TestNavigationToProjects(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenHome
	m.HomeCursor = 4 // Assuming "Proyectos" is the 5th entry (0-indexed)

	// Press Enter
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	if m.Screen != ScreenProjects {
		t.Errorf("expected ScreenProjects, got %v", m.Screen)
	}

	// Press Esc
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(Model)

	if m.Screen != ScreenHome {
		t.Errorf("expected ScreenHome after esc, got %v", m.Screen)
	}
}
