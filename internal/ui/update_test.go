package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"testing"
	"skillsync/tui/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSelectionSync(t *testing.T) {
	tests := []struct {
		name          string
		initialID     string
		moveKey       string
		expectContent bool
	}{
		{
			name:          "cursor move updates content",
			initialID:     "1",
			moveKey:       "j",
			expectContent: true,
		},
		{
			name:          "no movement keeps content same",
			initialID:     "1",
			moveKey:       "",
			expectContent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.Screen = ScreenList
			m.Width = 100
			m.Height = 50

			s1 := types.Skill{ID: "1", Name: "Skill 1", RawBody: "Body 1"}
			s2 := types.Skill{ID: "2", Name: "Skill 2", RawBody: "Body 2"}
			m.list.SetItems([]list.Item{item{skill: s1}, item{skill: s2}})
			m.viewport.Height = 10
			m.viewport.Width = 30
			m.lastSelectedID = tt.initialID
			m.updatePreview()

			var msg tea.Msg
			if tt.moveKey != "" {
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.moveKey)}
			} else {
				msg = tea.KeyMsg{Type: tea.KeySpace}
			}

			newModel, _ := m.Update(msg)
			m = newModel.(Model)

			if tt.expectContent {
				if m.viewport.Height == 0 {
					t.Error("viewport height is 0")
				}
				if m.viewport.View() == "" {
					t.Errorf("expected viewport content for skill %s, but got empty. ID: %s", m.selected.Name, m.lastSelectedID)
				}
				if m.lastSelectedID != "2" {
					t.Errorf("expected lastSelectedID to be '2', got '%s'", m.lastSelectedID)
				}
			}
		})
	}
}

func TestScreenTransitions(t *testing.T) {
	tests := []struct {
		name         string
		startScreen  Screen
		msg          tea.Msg
		expectScreen Screen
	}{
		{
			name:         "list to preview on enter",
			startScreen:  ScreenList,
			msg:          tea.KeyMsg{Type: tea.KeyEnter},
			expectScreen: ScreenContentView,
		},
		{
			name:         "list to detail on e",
			startScreen:  ScreenList,
			msg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")},
			expectScreen: ScreenDetail,
		},
		{
			name:         "list to syncing on s",
			startScreen:  ScreenList,
			msg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")},
			expectScreen: ScreenSyncing,
		},
		{
			name:         "detail to list on esc",
			startScreen:  ScreenDetail,
			msg:          tea.KeyMsg{Type: tea.KeyEsc},
			expectScreen: ScreenList,
		},
		{
			name:         "syncing to list on esc",
			startScreen:  ScreenSyncing,
			msg:          tea.KeyMsg{Type: tea.KeyEsc},
			expectScreen: ScreenList,
		},
		{
			name:         "list to content on v",
			startScreen:  ScreenList,
			msg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")},
			expectScreen: ScreenContentView,
		},
		{
			name:         "content to list on esc",
			startScreen:  ScreenContentView,
			msg:          tea.KeyMsg{Type: tea.KeyEsc},
			expectScreen: ScreenList,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.Screen = tt.startScreen
			m.PrevScreen = ScreenList

			// Mock list selection for enter
			if tt.startScreen == ScreenList && (tt.expectScreen == ScreenDetail || tt.expectScreen == ScreenContentView) {
				m.list.SetItems([]list.Item{item{}})
			}

			newModel, _ := m.Update(tt.msg)
			res := newModel.(Model)

			if res.Screen != tt.expectScreen {
				t.Errorf("expected screen %v, got %v", tt.expectScreen, res.Screen)
			}
		})
	}
}

func TestResizeMarkdownReflow(t *testing.T) {
	m := NewModel()
	m.Screen = ScreenList
	m.Width = 100
	m.Height = 50
	
	s1 := types.Skill{Name: "Skill 1", RawBody: "# Long Markdown\nThis is a long sentence that should wrap differently at different widths."}
	m.list.SetItems([]list.Item{item{skill: s1}})
	m.updatePreview()

	// Initial render
	msg1 := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel1, _ := m.Update(msg1)
	m1 := newModel1.(Model)
	content1 := m1.viewport.View()

	// Resize
	msg2 := tea.WindowSizeMsg{Width: 50, Height: 50}
	newModel2, _ := m1.Update(msg2)
	m2 := newModel2.(Model)
	content2 := m2.viewport.View()

	if content1 == content2 {
		t.Error("expected markdown to reflow on resize, but content is identical")
	}
}

func TestViewportScrollInList(t *testing.T) {
	m := NewModel()
	m.Screen = ScreenList
	m.Width = 100
	m.Height = 20
	
	s1 := types.Skill{Name: "Skill 1", RawBody: "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10"}
	m.list.SetItems([]list.Item{item{skill: s1}})
	m.updatePreview()
	m.viewport.Height = 3 // Small viewport to force scroll

	initialOffset := m.viewport.YOffset

	// Send pgdown
	msg := tea.KeyMsg{Type: tea.KeyPgDown}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.viewport.YOffset <= initialOffset {
		t.Errorf("expected viewport to scroll down, offset %d stays <= %d", m.viewport.YOffset, initialOffset)
	}
	
	if m.list.Index() != 0 {
		t.Error("list selection changed when it should only scroll viewport")
	}
}

func TestLoadSkills_VirtualInjection(t *testing.T) {
	// Setup: Create agents.md in root
	m := NewModel()
	m.rootPath = "../../.." // relative to internal/ui
	
	// We need to mock the filesystem or just rely on actual file for this run
	// Since I'm in the real environment, I'll check if it exists
	t.Run("injects virtual skill if agents.md exists", func(t *testing.T) {
		cmd := m.loadSkills()
		msg := cmd()
		
		skills, ok := msg.(skillsLoadedMsg)
		if !ok {
			t.Fatalf("expected skillsLoadedMsg, got %T", msg)
		}

		found := false
		for _, s := range skills {
			if s.ID == "virtual:agents" {
				found = true
				if s.Name != "★ AGENTS.md" {
					t.Errorf("expected name ★ AGENTS.md, got %s", s.Name)
				}
				break
			}
		}
		
		if !found {
			t.Error("virtual:agents skill not found in loaded skills")
		}
	})
}

func TestHandleListKeys_Matrix(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		expectScreen Screen
	}{
		{
			name:         "Enter goes to Preview (ScreenContentView)",
			key:          "enter",
			expectScreen: ScreenContentView,
		},
		{
			name:         "v goes to Preview (ScreenContentView)",
			key:          "v",
			expectScreen: ScreenContentView,
		},
		{
			name:         "e goes to Detail (ScreenDetail)",
			key:          "e",
			expectScreen: ScreenDetail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.Screen = ScreenList
			s1 := types.Skill{Name: "Skill 1", RawBody: "Body 1"}
			m.list.SetItems([]list.Item{item{skill: s1}})

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "enter" {
				msg = tea.KeyMsg{Type: tea.KeyEnter}
			}

			newModel, _ := m.Update(msg)
			res := newModel.(Model)

			if res.Screen != tt.expectScreen {
				t.Errorf("expected screen %v, got %v", tt.expectScreen, res.Screen)
			}
			if res.PrevScreen != ScreenList {
				t.Errorf("expected PrevScreen to be ScreenList, got %v", res.PrevScreen)
			}
		})
	}
}

func TestHandleDetailKeys_ReadOnly(t *testing.T) {
	tests := []struct {
		name        string
		skillID     string
		key         string
		expectCmd   bool
		expectScreen Screen
	}{
		{
			name:        "Virtual skill blocks ctrl+s",
			skillID:     "virtual:agents",
			key:         "ctrl+s",
			expectCmd:   false,
			expectScreen: ScreenDetail,
		},
		{
			name:        "Normal skill allows ctrl+s",
			skillID:     "normal:skill",
			key:         "ctrl+s",
			expectCmd:   true,
			expectScreen: ScreenDetail,
		},
		{
			name:        "Esc returns to PrevScreen",
			skillID:     "virtual:agents",
			key:         "esc",
			expectCmd:   false,
			expectScreen: ScreenList,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.Screen = ScreenDetail
			m.PrevScreen = ScreenList
			s := &types.Skill{ID: tt.skillID, Name: "Test Skill"}
			m.selected = s
			m.setupInputs()

			var msg tea.KeyMsg
			switch tt.key {
			case "ctrl+s":
				msg = tea.KeyMsg{Type: tea.KeyCtrlS}
			case "esc":
				msg = tea.KeyMsg{Type: tea.KeyEsc}
			}

			_, cmd := m.handleDetailKeys(msg)

			if tt.expectCmd && cmd == nil {
				t.Error("expected a command, got nil")
			}
			if !tt.expectCmd && cmd != nil {
				t.Errorf("expected no command, got %v", cmd)
			}
		})
	}
}
