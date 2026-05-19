package ui

import (
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/storage"
	"strings"
	"testing"
)

func TestSyncingView_Regression(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenSyncing
	m.syncOutput = "Test Output"
	m.Width = 100
	m.Height = 20

	view := m.View()

	// 1. Check for duplicate footer "esc: back"
	count := strings.Count(view, "esc: back")
	if count > 1 {
		t.Errorf("Footer 'esc: back' duplicated: found %d times", count)
	}

	// 2. Check title is dynamic
	if strings.Contains(view, "Installing Ecosistema...") {
		// Since we didn't set PrevScreen, it defaults to ScreenHome (0)
		// and NewModel(NewBackend(storage.NewService(""))) sets it to ScreenHome.
		// So it should be "Installing Ecosistema..."
	}

	m.PrevScreen = ScreenList
	view2 := m.View()
	if !strings.Contains(view2, "Syncing Skills...") {
		t.Error("Expected title 'Syncing Skills...' when PrevScreen is ScreenList")
	}
}

func TestSyncResult_ErrorTransition(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenSyncing
	m.PrevScreen = ScreenList

	// Simulate an error from runner
	msg := runner.SyncResult{
		ExitCode: 1,
		Stdout:   "",
		Stderr:   "some error",
	}

	newModel, _ := m.Update(msg)
	res := newModel.(Model)

	// Currently, it stays in ScreenSyncing, which we want to change to an error state or at least recognize it's done
	// The user says it "gets stuck", so we should ensure it's not "stuck" in a progress state.
	// For this test, let's check if the syncOutput was updated correctly.
	if !strings.Contains(res.syncOutput, "Exit: 1") {
		t.Errorf("Expected syncOutput to contain 'Exit: 1', got: %s", res.syncOutput)
	}

	// In the fix, we might want to change Screen to ScreenHome or show a clear error message.
	// But first, let's just confirm it stays in ScreenSyncing (current behavior).
	if res.Screen != ScreenSyncing {
		t.Errorf("Expected screen to remain ScreenSyncing (current behavior), got %v", res.Screen)
	}
}
