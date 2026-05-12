package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"skillsync/tui/internal/storage"
	"strings"
	"testing"
	"time"
)

func TestProjectsView(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenProjects
	m.Width = 100
	m.Height = 50

	// Initialize list size
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	m = newModel.(Model)

	// Mock some projects
	p1 := storage.ProjectInfo{Path: "/path/to/p1", LastSynced: time.Now()}
	p2 := storage.ProjectInfo{Path: "/path/to/p2", LastSynced: time.Now().Add(-1 * time.Hour)}

	m.projectList.SetItems([]list.Item{
		projectItem{project: p1},
		projectItem{project: p2},
	})

	view := m.View()

	if !strings.Contains(view, "/path/to/p1") {
		t.Errorf("expected view to contain /path/to/p1, got:\n%s", view)
	}
	if !strings.Contains(view, "/path/to/p2") {
		t.Errorf("expected view to contain /path/to/p2, got:\n%s", view)
	}
	if !strings.Contains(view, "Proyectos Sincronizados") {
		t.Errorf("expected view to contain 'Proyectos Sincronizados', got:\n%s", view)
	}
}

func TestProjectsView_EmptyState(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenProjects
	m.Width = 100
	m.Height = 50

	// No projects
	m.projectList.SetItems([]list.Item{})

	view := m.View()

	expectedMsg := "No se encontraron proyectos. Presioná '4' para sincronizar este proyecto y registrarlo."
	if !strings.Contains(view, expectedMsg) {
		t.Errorf("expected view to show actionable help message:\n%q\nGot:\n%s", expectedMsg, view)
	}
}
