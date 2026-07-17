package ui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/types"
)

func menuTestModel(t *testing.T) Model {
	t.Helper()
	dir := writeSkillTree(t)
	m := NewModel(NewBackend(storage.NewService(t.TempDir())))
	m.selected = &types.Skill{Name: "my-skill", Path: filepath.Join(dir, "SKILL.md")}
	return m
}

func TestListEnterOpensSkillMenu(t *testing.T) {
	m := menuTestModel(t)
	m.Screen = ScreenList
	m.List.selected = m.selected

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	nm := next.(Model)

	if nm.Screen != ScreenSkillMenu {
		t.Fatalf("expected ScreenSkillMenu, got %v", nm.Screen)
	}
	if nm.skillMenuOrigin != ScreenList {
		t.Fatalf("expected origin ScreenList, got %v", nm.skillMenuOrigin)
	}
}

func TestSkillMenuOptions(t *testing.T) {
	cases := []struct {
		name   string
		cursor int
		want   Screen
	}{
		{"preview", 0, ScreenContentView},
		{"edit", 1, ScreenDetail},
		{"files", 2, ScreenSkillFiles},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := menuTestModel(t)
			m.Screen = ScreenSkillMenu
			m.skillMenuOrigin = ScreenList
			m.skillMenuCursor = tc.cursor

			next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			nm := next.(Model)

			if nm.Screen != tc.want {
				t.Fatalf("option %s: expected %v, got %v", tc.name, tc.want, nm.Screen)
			}
		})
	}
}

func TestSkillMenuSaveToStorage(t *testing.T) {
	m := menuTestModel(t)
	m.Screen = ScreenSkillMenu
	m.skillMenuOrigin = ScreenGlobalSkillsList
	// Save option is the last entry in the menu.
	m.skillMenuCursor = len(skillMenuOptions) - 1
	if skillMenuOptions[m.skillMenuCursor] == "" {
		t.Fatal("no menu options")
	}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	nm := next.(Model)

	// Selecting save returns to the origin screen and issues a save command.
	if nm.Screen != ScreenGlobalSkillsList {
		t.Fatalf("expected return to ScreenGlobalSkillsList, got %v", nm.Screen)
	}
	if cmd == nil {
		t.Fatal("expected a save command, got nil")
	}
}

func TestSkillMenuHasSaveOption(t *testing.T) {
	found := false
	for _, opt := range skillMenuOptions {
		if strings.Contains(strings.ToLower(opt), "storage") || strings.Contains(strings.ToLower(opt), "almacen") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a save-to-storage option in menu, got %v", skillMenuOptions)
	}
}

func TestSkillMenuEscReturnsToOrigin(t *testing.T) {
	m := menuTestModel(t)
	m.Screen = ScreenSkillMenu
	m.skillMenuOrigin = ScreenGlobalSkillsList

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	nm := next.(Model)

	if nm.Screen != ScreenGlobalSkillsList {
		t.Fatalf("expected ScreenGlobalSkillsList, got %v", nm.Screen)
	}
}
