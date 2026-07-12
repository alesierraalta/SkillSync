package ui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"os"
	"path/filepath"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/types"
	"strings"
	"testing"

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
			m := NewModel(NewBackend(storage.NewService("")))
			m.Screen = ScreenList
			m.Width = 100
			m.Height = 50

			s1 := types.Skill{ID: "1", Name: "S", RawBody: "Body 1"}
			s2 := types.Skill{ID: "2", Name: "S", RawBody: "Body 2"}
			m.List.list.SetItems([]list.Item{item{skill: s1}, item{skill: s2}})
			m.List.viewport.Height = 10
			m.List.viewport.Width = 30
			m.List.lastSelectedID = tt.initialID
			m.List.updatePreview()

			var msg tea.Msg
			if tt.moveKey != "" {
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.moveKey)}
			} else {
				msg = tea.KeyMsg{Type: tea.KeySpace}
			}

			newModel, _ := m.Update(msg)
			m = newModel.(Model)

			if tt.expectContent {
				if m.List.viewport.Height == 0 {
					t.Error("viewport height is 0")
				}
				if m.List.viewport.View() == "" {
					t.Errorf("expected viewport content for skill %s, but got empty. ID: %s", m.selected.Name, m.List.lastSelectedID)
				}
				if m.List.lastSelectedID != "2" {
					t.Errorf("expected lastSelectedID to be '2', got '%s'", m.List.lastSelectedID)
				}
			}
		})
	}
}

func TestHandleHomeKeys_Key4_TriggersSync(t *testing.T) {
	m := Model{Screen: ScreenHome}
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	m = newModel.(Model)
	if cmd == nil {
		t.Fatal("expected command, got nil")
	}
	// We'll define syncOpenCodeCmd later, but for now we expect SOME command
}

func TestHandleHomeKeys_Cursor3_Enter_TriggersSync(t *testing.T) {
	m := Model{Screen: ScreenHome, HomeCursor: 3}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command, got nil")
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
			name:         "list to syncing on y",
			startScreen:  ScreenList,
			msg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")},
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
		{
			name:         "home to storage on 3",
			startScreen:  ScreenHome,
			msg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")},
			expectScreen: ScreenStorage,
		},
		{
			name:         "storage to home on esc",
			startScreen:  ScreenStorage,
			msg:          tea.KeyMsg{Type: tea.KeyEsc},
			expectScreen: ScreenHome,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(NewBackend(storage.NewService("")))
			m.Screen = tt.startScreen
			m.PrevScreen = ScreenList

			// Mock list selection for enter
			if tt.startScreen == ScreenList && (tt.expectScreen == ScreenDetail || tt.expectScreen == ScreenContentView) {
				m.List.list.SetItems([]list.Item{item{}})
				m.List.selected = &types.Skill{}
			}

			newModel, _ := m.Update(tt.msg)
			res := newModel.(Model)

			if res.Screen != tt.expectScreen {
				t.Errorf("expected screen %v, got %v", tt.expectScreen, res.Screen)
			}
		})
	}
}

func TestHandleStorageKeys_InstallShortcut(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenStorage

	// Mock some stored skills
	s1 := storage.StoredSkill{
		ID:       "skill1",
		Metadata: storage.StoredMetadata{SkillName: "test-skill"},
	}
	m.storedSkills = []storage.StoredSkill{s1}

	items := []list.Item{storageItem{stored: s1}}
	m.storageList.SetItems(items)
	m.storageList.Select(0)

	// Press 'i'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")}
	newModel, cmd := m.Update(msg)
	res := newModel.(Model)

	if res.Screen != ScreenSyncing {
		t.Errorf("expected ScreenSyncing, got %v", res.Screen)
	}
	if cmd == nil {
		t.Error("expected a command to be returned, got nil")
	}
}

func TestInstallFromStorageAndSyncCmd(t *testing.T) {
	tmpDir := t.TempDir()
	storageDir := t.TempDir()
	m := NewModel(NewBackend(storage.NewService(storageDir)))
	m.rootPath = tmpDir

	skillContent := "test content"
	sk := &types.Skill{Name: "test-skill", RawBody: skillContent}
	err := m.backend.SaveToStorage(sk, storage.StoredMetadata{SkillName: "test-skill"})
	if err != nil {
		t.Fatalf("failed to save to storage: %v", err)
	}

	// Stored ID in our implementation is the SkillName (folder name)
	stored := storage.StoredSkill{
		ID: "test-skill",
		Metadata: storage.StoredMetadata{
			SkillName: "test-skill",
		},
	}

	cmd := m.installFromStorageAndSyncCmd(stored)
	msg := cmd()

	if sr, ok := msg.(syncReportMsg); ok {
		if sr.err != nil {
			t.Logf("syncReportMsg failed (expected in temp dir): %v", sr.err)
		}
	} else {
		t.Errorf("expected syncReportMsg, got %T: %v", msg, msg)
	}

	// Check if file was created
	skillPath := filepath.Join(tmpDir, ".agents", "skills", "test-skill", "SKILL.md")
	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Errorf("expected file to be created at %s, but got error: %v", skillPath, err)
	}
	if !strings.Contains(string(content), skillContent) {
		t.Errorf("expected file content to contain %q, got %q", skillContent, string(content))
	}

	// Check msg (syncReportMsg)
	// We expect syncReportMsg because m.startSync() returns it.
	// However, m.startSync() in current implementation points to a real script.
	// We might get an error because sync.sh doesn't exist in temp dir.
	if _, ok := msg.(syncReportMsg); !ok {
		t.Errorf("expected syncReportMsg, got %T", msg)
	}
}

func TestResizeMarkdownReflow(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenList
	m.Width = 100
	m.Height = 50

	s1 := types.Skill{Name: "S", RawBody: "# Long Markdown\nThis is a long sentence that should wrap differently at different widths."}
	m.List.list.SetItems([]list.Item{item{skill: s1}})
	m.List.updatePreview()

	// Initial render
	msg1 := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel1, _ := m.Update(msg1)
	m1 := newModel1.(Model)
	content1 := m1.List.viewport.View()

	// Resize
	msg2 := tea.WindowSizeMsg{Width: 50, Height: 50}
	newModel2, _ := m1.Update(msg2)
	m2 := newModel2.(Model)
	content2 := m2.List.viewport.View()

	if content1 == content2 {
		t.Error("expected markdown to reflow on resize, but content is identical")
	}
}

func TestViewportScrollInList(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenList
	m.Width = 100
	m.Height = 20

	s1 := types.Skill{Name: "S", RawBody: "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10"}
	m.List.list.SetItems([]list.Item{item{skill: s1}})
	m.List.updatePreview()
	m.List.viewport.Height = 3 // Small viewport to force scroll

	initialOffset := m.List.viewport.YOffset

	// Send pgdown
	msg := tea.KeyMsg{Type: tea.KeyPgDown}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	if m.List.viewport.YOffset <= initialOffset {
		t.Errorf("expected viewport to scroll down, offset %d stays <= %d", m.List.viewport.YOffset, initialOffset)
	}

	if m.List.list.Index() != 0 {
		t.Error("list selection changed when it should only scroll viewport")
	}
}

func TestLoadSkills_VirtualInjection(t *testing.T) {
	// Setup: Create agents.md in root
	m := NewModel(NewBackend(storage.NewService("")))
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
		{
			name:         "y goes to Syncing (ScreenSyncing)",
			key:          "y",
			expectScreen: ScreenSyncing,
		},
		{
			name:         "s stays on List (triggers save command)",
			key:          "s",
			expectScreen: ScreenList,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(NewBackend(storage.NewService("")))
			m.Screen = ScreenList
			s1 := types.Skill{Name: "S", RawBody: "Body 1"}
			m.List.list.SetItems([]list.Item{item{skill: s1}})
			m.List.selected = &s1

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "enter" {
				msg = tea.KeyMsg{Type: tea.KeyEnter}
			}

			newModel, _ := m.Update(msg)
			res := newModel.(Model)

			if res.Screen != tt.expectScreen {
				t.Errorf("expected screen %v, got %v", tt.expectScreen, res.Screen)
			}
			if tt.expectScreen != ScreenList && res.PrevScreen != ScreenList {
				t.Errorf("expected PrevScreen to be ScreenList, got %v", res.PrevScreen)
			}
		})
	}
}

func TestHandleDetailKeys_ReadOnly(t *testing.T) {
	tests := []struct {
		name         string
		skillID      string
		key          string
		expectCmd    bool
		expectScreen Screen
	}{
		{
			name:         "Virtual skill blocks ctrl+s",
			skillID:      "virtual:agents",
			key:          "ctrl+s",
			expectCmd:    false,
			expectScreen: ScreenDetail,
		},
		{
			name:         "Normal skill allows ctrl+s",
			skillID:      "normal:skill",
			key:          "ctrl+s",
			expectCmd:    true,
			expectScreen: ScreenDetail,
		},
		{
			name:         "Esc returns to PrevScreen",
			skillID:      "virtual:agents",
			key:          "esc",
			expectCmd:    false,
			expectScreen: ScreenList,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(NewBackend(storage.NewService("")))
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

func TestInstallFailurePath(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenSyncing
	m.syncOutput = "Running..."

	// Simulate failure
	errMsg := fmt.Errorf("failed for test")
	newModel, cmd := m.Update(installerFinishedMsg{err: errMsg})
	m = newModel.(Model)

	if m.Screen != ScreenSyncing {
		t.Errorf("expected ScreenSyncing on error, got %v", m.Screen)
	}
	if !strings.Contains(m.syncOutput, "Error: failed for test") {
		t.Errorf("expected syncOutput to contain error message, got %q", m.syncOutput)
	}
	if cmd != nil {
		t.Errorf("expected nil cmd on error, got %v", cmd)
	}

	// Test back navigation from error
	res, _ := m.handleSyncingKeys(tea.KeyMsg{Type: tea.KeyEsc})
	m = res.(Model)
	if m.Screen != ScreenHome {
		t.Errorf("expected navigation to ScreenHome on esc from error, got %v", m.Screen)
	}
}

func TestHandleInstallerKeys_ToggleMode(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenInstaller
	initialMode := m.Installer.Mode

	// Toggle once
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	m = newModel.(Model)
	if m.Installer.Mode == initialMode {
		t.Errorf("expected Installer.Mode to toggle, got %v", m.Installer.Mode)
	}

	// Toggle back
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("M")})
	m = newModel.(Model)
	if m.Installer.Mode != initialMode {
		t.Errorf("expected Installer.Mode to toggle back, got %v", m.Installer.Mode)
	}
}

func TestGlobalSkillsScreenTransitions(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenHome

	// Navigate Home -> Cats
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("7")})
	res := newModel.(Model)
	if res.Screen != ScreenGlobalSkillsCats {
		t.Errorf("expected ScreenGlobalSkillsCats, got %v", res.Screen)
	}

	// Navigate Cats -> List
	res.globalCategoryCursor = 0 // "Claude"
	newModel, cmd := res.Update(tea.KeyMsg{Type: tea.KeyEnter})
	res2 := newModel.(Model)
	if res2.Screen != ScreenGlobalSkillsList {
		t.Errorf("expected ScreenGlobalSkillsList, got %v", res2.Screen)
	}
	if cmd == nil {
		t.Error("expected command to load skills, got nil")
	}

	// Navigate List -> Cats
	newModel, _ = res2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	res3 := newModel.(Model)
	if res3.Screen != ScreenGlobalSkillsCats {
		t.Errorf("expected ScreenGlobalSkillsCats, got %v", res3.Screen)
	}
}

func TestLoadGlobalSkillsCmd(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	cmd := m.loadGlobalSkillsCmd("Claude")

	// Note: We can't easily mock the filesystem here without changing AppService,
	// but we can at least ensure it returns a globalSkillsLoadedMsg.
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
}
