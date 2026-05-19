package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/types"
	"testing"
)

func TestNewModel_SearchInitialization(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))

	// Task 1.2: Check fields
	if m.List.searchInput.Placeholder != "Search skills..." {
		t.Errorf("expected searchInput placeholder 'Search skills...', got %q", m.List.searchInput.Placeholder)
	}

	if m.List.searchFocused != false {
		t.Error("expected searchFocused to be false initially")
	}

	// Task 1.3: Check list filter disabled
	if m.List.list.KeyMap.Filter.Enabled() {
		t.Error("expected list filter key to be disabled")
	}
}

func TestModel_FilterSkills(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		allSkills     []string
		expectedCount int
	}{
		{
			name:          "match substring",
			query:         "git",
			allSkills:     []string{"git helper", "docker setup", "git sync"},
			expectedCount: 2,
		},
		{
			name:          "case insensitive",
			query:         "GIT",
			allSkills:     []string{"git helper", "docker setup"},
			expectedCount: 1,
		},
		{
			name:          "empty query returns all",
			query:         "",
			allSkills:     []string{"a", "b", "c"},
			expectedCount: 3,
		},
		{
			name:          "no match returns zero",
			query:         "zzz",
			allSkills:     []string{"a", "b", "c"},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(NewBackend(storage.NewService("")))
			var items []list.Item
			for _, s := range tt.allSkills {
				items = append(items, item{skill: types.Skill{Name: s}})
			}
			m.List.allSkills = items

			m.filterSkills(tt.query)

			if len(m.List.list.Items()) != tt.expectedCount {
				t.Errorf("expected %d items, got %d", tt.expectedCount, len(m.List.list.Items()))
			}
		})
	}
}

func TestModel_FocusTransitions(t *testing.T) {
	tests := []struct {
		name          string
		initialFocus  bool // true = search, false = list
		key           string
		expectedFocus bool
	}{
		{
			name:          "Tab from list to search",
			initialFocus:  false,
			key:           "tab",
			expectedFocus: true,
		},
		{
			name:          "Esc from search to list",
			initialFocus:  true,
			key:           "esc",
			expectedFocus: false,
		},
		{
			name:          "Tab from search to list",
			initialFocus:  true,
			key:           "tab",
			expectedFocus: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(NewBackend(storage.NewService("")))
			m.Screen = ScreenList
			m.List.searchFocused = tt.initialFocus
			if m.List.searchFocused {
				m.List.searchInput.Focus()
			} else {
				m.List.searchInput.Blur()
			}

			var msg tea.Msg
			switch tt.key {
			case "tab":
				msg = tea.KeyMsg{Type: tea.KeyTab}
			case "esc":
				msg = tea.KeyMsg{Type: tea.KeyEsc}
			}

			newModel, _ := m.Update(msg)
			res := newModel.(Model)

			if res.List.searchFocused != tt.expectedFocus {
				t.Errorf("expected searchFocused %v, got %v", tt.expectedFocus, res.List.searchFocused)
			}
			if res.List.searchInput.Focused() != tt.expectedFocus {
				t.Errorf("expected searchInput.Focused() %v, got %v", tt.expectedFocus, res.List.searchInput.Focused())
			}
		})
	}
}

func TestModel_SlashKeyDisabled(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenList
	m.List.searchFocused = false

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}
	newModel, _ := m.Update(msg)
	res := newModel.(Model)

	// If it doesn't activate anything, searchFocused should stay false
	if res.List.searchFocused {
		t.Error("expected searchFocused to stay false on '/'")
	}
}

func TestModel_SearchIntegration(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenList

	s1 := types.Skill{Name: "git helper"}
	s2 := types.Skill{Name: "docker setup"}
	m.List.allSkills = []list.Item{item{skill: s1}, item{skill: s2}}
	m.List.list.SetItems(m.List.allSkills)

	// 1. Tab to focus search
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyTab})
	if !m.List.searchFocused {
		t.Fatal("expected search to be focused after Tab")
	}

	// 2. Type 'g'
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	if len(m.List.list.Items()) != 1 {
		t.Errorf("expected 1 item after 'g', got %d", len(m.List.list.Items()))
	}
	if it, ok := m.List.list.Items()[0].(item); !ok || it.skill.Name != "git helper" {
		t.Errorf("expected 'git helper', got %v", m.List.list.Items()[0])
	}

	// 3. Backspace (delete 'g')
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyBackspace})
	if len(m.List.list.Items()) != 2 {
		t.Errorf("expected 2 items after backspace, got %d", len(m.List.list.Items()))
	}
}
