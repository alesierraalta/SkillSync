package ui

import (
	"testing"
	"skillsync/tui/internal/types"
)

func TestSetupInputs_SmallHeight(t *testing.T) {
	m := NewModel()
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
