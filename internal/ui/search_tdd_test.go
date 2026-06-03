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

func TestModel_SearchFocusClash(t *testing.T) {
	tests := []struct {
		key           string
		initialScreen Screen
	}{
		{key: "e", initialScreen: ScreenList},
		{key: "s", initialScreen: ScreenList},
		{key: "y", initialScreen: ScreenList},
		{key: "d", initialScreen: ScreenList},
	}

	for _, tt := range tests {
		t.Run("key "+tt.key, func(t *testing.T) {
			m := NewModel(NewBackend(storage.NewService("")))
			m.Screen = tt.initialScreen
			m.List.searchFocused = true
			m.List.searchInput.Focus()

			// Set up a selected item so that actions like 'e' or 'd' could be triggered
			s1 := types.Skill{ID: "test-skill", Name: "test-skill-name"}
			m.List.selected = &s1
			m.List.allSkills = []list.Item{item{skill: s1}}
			m.List.list.SetItems(m.List.allSkills)

			var msg tea.Msg
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}

			newModel, _ := m.Update(msg)
			res := newModel.(Model)

			// Ensure screen has NOT changed
			if res.Screen != ScreenList {
				t.Errorf("expected screen to remain ScreenList, got %v for key %s", res.Screen, tt.key)
			}

			// Ensure no commands were returned (or at least not commands initiating sync/storage)
			// Actually, since List.Update(msg) might return some basic tea.Cmd, we should check specifically
			// that the Screen didn't change and we are still on the list screen.
			// Let's verify that the screen didn't change.
			if tt.key == "e" && res.Screen == ScreenDetail {
				t.Error("key 'e' triggered ScreenDetail even though search was focused")
			}
			if tt.key == "y" && res.Screen == ScreenSyncing {
				t.Error("key 'y' triggered ScreenSyncing even though search was focused")
			}
			if tt.key == "d" && res.Screen == ScreenDeleteConfirm {
				t.Error("key 'd' triggered ScreenDeleteConfirm even though search was focused")
			}

			// We can also verify that cmd is nil or not the action cmd.
			// (Note: saving to storage returns a cmd but doesn't change screen, so we can verify if a cmd is returned or we can inspect m's state or behavior if possible).
			// Wait, saveToStorageCmd returns a tea.Cmd. If we press 's', a cmd is returned. But when search is focused, 's' is just typed into the search bar, returning no cmd (or list cmd).
			// Let's check that if key is "s", we didn't trigger storage/saving command. Or better: we just type "s" and the search input value becomes "s"!
			if tt.key == "s" {
				// Since we sent 's', it should be passed to the searchInput
				if res.List.searchInput.Value() != "s" {
					t.Errorf("expected searchInput value to be 's', got %q", res.List.searchInput.Value())
				}
			}
		})
	}
}
