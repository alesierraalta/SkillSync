package ui

import (
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/types"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSearchHotkeyClashRepro(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenList

	s1 := types.Skill{ID: "s1", Name: "git helper", Path: "git-helper.md"}
	m.List.allSkills = []list.Item{item{skill: s1}}
	m.List.list.SetItems(m.List.allSkills)
	m.List.list.Select(0)
	m.List.updatePreview()

	// 1. Focus the search bar via Tab
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyTab})
	if !m.List.searchFocused {
		t.Fatal("expected search to be focused after Tab")
	}

	// 2. Type 'e' while search is focused
	// Expected behavior: it should type 'e' into the search input and NOT open the edit screen.
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if m.Screen == ScreenDetail {
		t.Fatal("hotkey clash reproduced: typing 'e' in search bar opened edit skill screen!")
	}
	if m.List.searchInput.Value() != "e" {
		t.Errorf("expected searchInput value to be 'e', got %q", m.List.searchInput.Value())
	}
}
