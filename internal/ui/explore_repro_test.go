package ui

import (
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/syncengine"
	"skillsync/tui/internal/types"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func TestContentViewHeaderSyncBug(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenContentView
	m.Width = 100
	m.Height = 50

	s1 := types.Skill{ID: "s1", Name: "Skill 1", Path: "s1.md"}
	s2 := types.Skill{ID: "s2", Name: "Skill 2", Path: "s2.md"}

	m.List.list.SetItems([]list.Item{item{skill: s1}, item{skill: s2}})
	m.List.list.Select(0)
	m.selected = &s1
	m.List.updatePreview()

	// Initial check
	if m.selected.Name != "Skill 1" {
		t.Errorf("expected selected name 'Skill 1', got '%s'", m.selected.Name)
	}

	// Manually move the list selection to Skill 2
	m.List.list.Select(1)

	// Send a dummy key that is not intercepted (e.g. "a")
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// The list is at Skill 2
	if m.List.list.Index() != 1 {
		t.Errorf("expected list index 1, got %d", m.List.list.Index())
	}

	// BUG: m.selected was NOT updated in handleContentViewKeys
	if m.selected.Name == "Skill 1" {
		t.Errorf("BUG DETECTED: ScreenContentView header out of sync! Expected 'Skill 2', got '%s'", m.selected.Name)
	}
}

func TestContentViewScrollBug(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenContentView
	m.Width = 100
	m.Height = 50

	s1 := types.Skill{ID: "s1", Name: "Skill 1", Path: "s1.md"}
	m.List.list.SetItems([]list.Item{item{skill: s1}})
	m.List.list.Select(0)
	m.selected = &s1

	// Set long content to make it scrollable
	m.List.viewport.SetContent("line 1\nline 2\nline 3\nline 4\nline 5")
	initialYOffset := m.List.viewport.YOffset

	// Press 'j' to scroll
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// BUG: viewport did not scroll because 'j' was consumed by list (or just not passed to viewport)
	if m.List.viewport.YOffset == initialYOffset {
		t.Errorf("BUG DETECTED: ScreenContentView 'j' did not scroll the viewport")
	}
}

func TestSyncFinishedResetBug(t *testing.T) {
	m := NewModel(&MockAppService{
		SyncFunc: func(root string, opts syncengine.SyncOptions) (*runner.SyncReport, error) {
			return &runner.SyncReport{}, nil
		},
		InstallCoreSkillFunc: func(name string) error { return nil },
		EnsureAgentsMDFunc:   func(root string) error { return nil },
		RegisterProjectFunc:  func(path string) error { return nil },
	})

	// 1. First sync finishes
	m.Screen = ScreenSyncing
	newModel, _ := m.Update(runner.SyncResult{ExitCode: 0})
	m = newModel.(Model)
	if !m.SyncFinished {
		t.Fatalf("expected SyncFinished to be true after first sync")
	}

	// 2. Start second sync from List
	m.Screen = ScreenList
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m = newModel.(Model)

	// BUG: SyncFinished is still true
	if m.SyncFinished {
		t.Errorf("BUG DETECTED: SyncFinished not reset when starting new sync! View will show stale data.")
	}
}

func TestInstallerStoredSkillsInitBug(t *testing.T) {
	backend := &MockAppService{
		ListStoredSkillsFunc: func() ([]storage.StoredSkill, error) {
			return []storage.StoredSkill{{ID: "s1", Metadata: storage.StoredMetadata{SkillName: "S1"}}}, nil
		},
	}
	m := NewModel(backend)

	// 1. Load stored skills while on Home screen
	msg := m.loadStoredSkillsCmd()()
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if len(m.storedSkills) == 0 {
		t.Fatalf("expected storedSkills to be loaded")
	}

	// 2. Switch to Installer
	// Simulate Home Screen Cursor 0 Enter
	m.HomeCursor = 0
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	if m.Screen != ScreenInstaller {
		t.Fatalf("expected ScreenInstaller, got %v", m.Screen)
	}

	// BUG: Installer.StoredSkills was NOT initialized because the msg arrived while on ScreenHome
	if m.Installer.StoredSkills == nil || len(m.Installer.StoredSkills) == 0 {
		t.Errorf("BUG DETECTED: Installer.StoredSkills not initialized when switching screens with pre-loaded skills")
	}
}

func TestESCListBug(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenList
	m.HomeCursor = 2 // Set to something non-zero

	// Press ESC
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.Screen != ScreenHome {
		t.Fatalf("expected ScreenHome, got %v", m.Screen)
	}

	// BUG: HomeCursor is reset to 0 in handleListKeys
	if m.HomeCursor != 2 {
		t.Errorf("BUG DETECTED: HomeCursor reset to %d after ESC from list, expected 2", m.HomeCursor)
	}
}
