package ui

import (
	"skillsync/tui/internal/storage"
	"strings"
	"testing"
)

func TestSyncingViewTitle(t *testing.T) {
	tests := []struct {
		name       string
		prevScreen Screen
		wantTitle  string
	}{
		{"from storage", ScreenStorage, "Installing Skill..."},
		{"from list", ScreenList, "Syncing Skills..."},
		{"from content", ScreenContentView, "Syncing Skills..."},
		{"from home", ScreenHome, "Installing Ecosistema..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(NewBackend(storage.NewService("")))
			m.Screen = ScreenSyncing
			m.PrevScreen = tt.prevScreen

			view := m.syncingView()
			if !strings.Contains(view, tt.wantTitle) {
				t.Errorf("syncingView() title = %q, want it to contain %q", view, tt.wantTitle)
			}
		})
	}
}
