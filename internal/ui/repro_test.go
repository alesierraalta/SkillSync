package ui

import (
	"path/filepath"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/types"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDuplicateNameSelectionBug(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenList
	m.Width = 100
	m.Height = 50

	// Two skills with the same name but different paths and content
	s1 := types.Skill{ID: filepath.ToSlash(".agents/skills/skill-creator/SKILL.md"), Name: "skill-creator", Path: ".agents/skills/skill-creator/SKILL.md", RawBody: "Contenido base."}
	s2 := types.Skill{ID: filepath.ToSlash(".agents/skills/skill-creator/skill-creator/SKILL.md"), Name: "skill-creator", Path: ".agents/skills/skill-creator/skill-creator/SKILL.md", RawBody: "Real Content"}

	m.List.list.SetItems([]list.Item{item{skill: s1}, item{skill: s2}})
	m.List.viewport.Height = 20
	m.List.viewport.Width = 30

	// Select first skill
	m.List.updatePreview()
	if m.List.lastSelectedID != filepath.ToSlash(s1.Path) {
		t.Errorf("expected lastSelectedID to be '%s', got '%s'", filepath.ToSlash(s1.Path), m.List.lastSelectedID)
	}
	if m.List.viewport.View() == "" || !contains(m.List.viewport.View(), "Contenido base.") {
		t.Errorf("expected preview to show 'Contenido base.', got '%s'", m.List.viewport.View())
	}

	// Move to second skill
	m.List.list.Select(1)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = newModel.(Model)

	// lastSelectedID is now "filepath.ToSlash(s2.Path)", updatePreview WAS called
	if contains(m.List.viewport.View(), "Real Content") {
		t.Log("Preview updated correctly")
	} else if contains(m.List.viewport.View(), "Contenido base.") {
		t.Error("BUG STILL PRESENT: Preview did not update when moving between skills with same name")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
