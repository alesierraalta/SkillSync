package ui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"skillsync/tui/internal/storage"
)

func syncProvidersModel(t *testing.T) Model {
	t.Helper()
	return NewModel(NewBackend(storage.NewService(t.TempDir())))
}

func TestListYOpensSyncProviders(t *testing.T) {
	m := syncProvidersModel(t)
	m.Screen = ScreenList
	m.List.selected = nil

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	nm := next.(Model)

	if nm.Screen != ScreenSyncProviders {
		t.Fatalf("expected ScreenSyncProviders, got %v", nm.Screen)
	}
	if len(nm.syncProviderSel) != len(syncProviderList) {
		t.Fatalf("selection slice size mismatch: %d vs %d", len(nm.syncProviderSel), len(syncProviderList))
	}
	// OpenCode is pre-selected by default.
	if !nm.syncProviderSel[0] || syncProviderList[0].Dir != ".opencode" {
		t.Errorf("expected OpenCode pre-selected")
	}
}

func TestSyncProvidersPreselectsExistingDirs(t *testing.T) {
	m := syncProvidersModel(t)
	root := t.TempDir()
	m.rootPath = root
	// Create a .claude dir so it should be pre-checked.
	if err := os.MkdirAll(filepath.Join(root, ".claude"), 0755); err != nil {
		t.Fatal(err)
	}
	m.Screen = ScreenList

	next, _ := m.openSyncProviders()
	nm := next.(Model)

	claudeIdx := -1
	for i, p := range syncProviderList {
		if p.Dir == ".claude" {
			claudeIdx = i
		}
	}
	if claudeIdx < 0 || !nm.syncProviderSel[claudeIdx] {
		t.Errorf("expected .claude pre-selected because its dir exists")
	}
}

func TestSyncProvidersToggleAndConfirm(t *testing.T) {
	m := syncProvidersModel(t)
	m.Screen = ScreenSyncProviders
	m.syncProviderSel = make([]bool, len(syncProviderList))
	m.syncProviderSel[0] = true // OpenCode
	m.syncProviderCursor = 1     // Claude

	// space toggles Claude on
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	nm := next.(Model)
	if !nm.syncProviderSel[1] {
		t.Fatalf("space did not toggle provider on")
	}

	// enter confirms → goes to syncing and issues a command
	next2, cmd := nm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	nm2 := next2.(Model)
	if nm2.Screen != ScreenSyncing {
		t.Fatalf("expected ScreenSyncing after enter, got %v", nm2.Screen)
	}
	if cmd == nil {
		t.Fatal("expected a sync command")
	}
}

func TestSyncProvidersEnterWithNoneSelected(t *testing.T) {
	m := syncProvidersModel(t)
	m.Screen = ScreenSyncProviders
	m.syncProviderSel = make([]bool, len(syncProviderList)) // all off

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	nm := next.(Model)
	if nm.Screen != ScreenSyncProviders {
		t.Errorf("expected to stay on ScreenSyncProviders when nothing selected, got %v", nm.Screen)
	}
}

func TestSyncProvidersEscCancels(t *testing.T) {
	m := syncProvidersModel(t)
	m.Screen = ScreenList
	next, _ := m.openSyncProviders()
	nm := next.(Model)

	next2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyEsc})
	nm2 := next2.(Model)
	if nm2.Screen != ScreenList {
		t.Errorf("expected esc to return to ScreenList, got %v", nm2.Screen)
	}
}

func TestSelectedProviderDirs(t *testing.T) {
	m := syncProvidersModel(t)
	m.syncProviderSel = make([]bool, len(syncProviderList))
	m.syncProviderSel[0] = true // OpenCode
	m.syncProviderSel[1] = true // Claude

	dirs := m.selectedProviderDirs()
	if len(dirs) != 2 || dirs[0] != ".opencode" || dirs[1] != ".claude" {
		t.Errorf("unexpected selected dirs: %v", dirs)
	}
}
