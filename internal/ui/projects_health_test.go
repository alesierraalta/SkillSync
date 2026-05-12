package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"path/filepath"
	"skillsync/tui/internal/storage"
	"testing"
	"time"
)

func TestProjectsRefresh(t *testing.T) {
	// 3.1 RED: Test Update in internal/ui/update.go for 'r' key message handling.
	tmpHome := t.TempDir()
	t.Setenv("SKILLSYNC_HOME", tmpHome)

	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenProjects
	m.storageService = storage.NewService("") // Uses SKILLSYNC_HOME

	// Setup a project on disk
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "p1")
	os.MkdirAll(filepath.Join(projectDir, ".agents"), 0755)

	// Simulate 'r' key press
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	m = m2.(Model)

	if cmd == nil {
		t.Fatal("expected 'r' key to return a command")
	}

	// Trigger the resulting message.
	p1 := storage.ProjectInfo{Path: projectDir, LastSynced: time.Now()}
	m3, _ := m.Update(projectsLoadedMsg{p1})
	m = m3.(Model)

	items := m.projectList.Items()
	if len(items) != 1 {
		t.Errorf("expected 1 project item, got %d", len(items))
	} else {
		item := items[0].(projectItem)
		if item.project.Path != projectDir {
			t.Errorf("expected path %q, got %q", projectDir, item.project.Path)
		}
	}
}
