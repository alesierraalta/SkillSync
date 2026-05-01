package ui

import (
	"strings"
	"testing"

	"skillsync/tui/internal/runner"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSyncFailureExplicitState(t *testing.T) {
	m := NewModel()
	m.Screen = ScreenSyncing
	m.PrevScreen = ScreenList

	// 1. Simulate a sync failure with the exact output shape reported by the user
	// Exit: 1, Output: "", Err: "" -> lacks "Error:" prefix
	msg := runner.SyncResult{
		ExitCode: 1,
		Stdout:   "",
		Stderr:   "",
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Verify explicit failure state is set
	if !m.SyncFailed {
		t.Error("expected SyncFailed to be true after non-zero exit code")
	}

	// 2. Verify view rendering
	view := m.syncingView()
	
	// Should show "esc: back"
	if !strings.Contains(view, "esc: back") {
		t.Error("expected view to contain 'esc: back' in failure state")
	}
	
	// 3. Verify esc navigation
	res, _ := m.handleSyncingKeys(tea.KeyMsg{Type: tea.KeyEsc})
	m = res.(Model)
	if m.Screen != ScreenHome {
		t.Errorf("expected ScreenHome after esc in failure state, got %v", m.Screen)
	}
}

func TestSyncSuccessResetsState(t *testing.T) {
	m := NewModel()
	m.SyncFailed = true // start with failed state
	m.Screen = ScreenList

	// Trigger sync again
	newModel, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m = newModel.(Model)

	if m.SyncFailed {
		t.Error("expected SyncFailed to be reset to false when starting new sync")
	}
}
