package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/types"
	"testing"
)

func TestSetupInputs_SmallHeight(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Width = 80
	m.Height = 10 // Very small height

	m.selected = &types.Skill{
		ID:   "test",
		Name: "Test",
	}

	m.setupInputs()

	if len(m.inputs) != 3 {
		t.Fatalf("expected 3 inputs, got %d", len(m.inputs))
	}

	// Overhead was 17. 10 - 17 = -7.
	// contentHeight should have been clamped to 4 in setupInputs.
	if m.inputs[2].Height() != 4 {
		t.Errorf("expected content textarea height to be 4, got %d", m.inputs[2].Height())
	}
}

func TestStorageList_Height(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenStorage

	height := 24
	updatedModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: height})
	m = updatedModel.(Model)

	expectedHeight := height - 8
	actualHeight := m.storageList.Height()

	if actualHeight != expectedHeight {
		t.Errorf("expected storageList height %d, got %d", expectedHeight, actualHeight)
	}
}
