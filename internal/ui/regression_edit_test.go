package ui

import (
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/types"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestEditSkillContentShortcutBug(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenContentView
	m.Width = 100
	m.Height = 50

	skill := types.Skill{
		ID:   "test-skill",
		Name: "Test Skill",
		Path: "test/SKILL.md",
		Metadata: types.Metadata{
			Description: "Test Description",
			Scope:       "Test Scope",
		},
		RawBody: "Test Content",
	}
	m.selected = &skill

	// Manual update instead of updatePreview which depends on list state
	metadata := "*e Edit content*\n"
	m.List.viewport.SetContent(metadata + skill.RawBody)

	// Verify initial affordance text
	t.Logf("Viewport content (raw): %q", m.List.viewport.View())
	if !contains(metadata+skill.RawBody, "e Edit content") {
		t.Errorf("expected affordance 'e Edit content' in viewport, not found")
	}

	// Simulate pressing 'e' in ScreenContentView
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m = newModel.(Model)

	if m.Screen != ScreenDetail {
		t.Errorf("expected Screen to be ScreenDetail after pressing 'e', got %v", m.Screen)
	}

	// CHECK IF CONTENT IS EDITABLE
	if len(m.inputs) < 3 {
		t.Fatalf("expected at least 3 inputs (Description, Scope, Content), got %d", len(m.inputs))
	} else if m.inputs[2].Value() != "Test Content" {
		t.Errorf("expected third input value to be 'Test Content', got %q", m.inputs[2].Value())
	}

	// Tab to focus content textarea (from input[0] -> input[1] -> input[2])
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyTab})
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyTab})

	if !m.inputs[2].Focused() {
		t.Errorf("expected third input to be focused after 2 tabs")
	}

	// Press Enter in content field
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.inputs[2].Value() != "Test Content\n" {
		t.Errorf("expected newline to be accepted in third input, got %q", m.inputs[2].Value())
	}

	// Simulate Ctrl+S
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.selected.RawBody != "Test Content\n" {
		t.Errorf("expected selected.RawBody to be updated to %q, got %q", "Test Content\n", m.selected.RawBody)
	}
}

func updateModel(m Model, msg tea.Msg) (Model, tea.Cmd) {
	newModel, cmd := m.Update(msg)
	return newModel.(Model), cmd
}
