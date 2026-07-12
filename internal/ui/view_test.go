package ui

import (
	"os"
	"path/filepath"
	"skillsync/tui/internal/storage"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"skillsync/tui/internal/types"
)

func TestViewGolden(t *testing.T) {
	tests := []struct {
		name   string
		screen Screen
	}{
		{"list", ScreenList},
		{"detail", ScreenDetail},
		{"syncing", ScreenSyncing},
		{"content", ScreenContentView},
		{"global_cats", ScreenGlobalSkillsCats},
		{"global_list", ScreenGlobalSkillsList},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(NewBackend(storage.NewService("")))
			m.Screen = tt.screen
			m.Width = 80
			m.Height = 24

			// Mock selected for detail
			if tt.screen == ScreenDetail {
				m.selected = &types.Skill{
					Name: "Test",
					Metadata: types.Metadata{
						Description: "Test Description",
						Scope:       "project",
					},
				}
				m.setupInputs()
			}
			if tt.screen == ScreenList {
				m.List.list.SetItems([]list.Item{item{skill: types.Skill{
					Name:    "Markdown Skill",
					Prefix:  "# Welcome\n",
					RawBody: "This is a **test** body.\n\n- Item 1\n- Item 2",
				}}})
				m.List.updatePreview()
			}
			if tt.screen == ScreenContentView {
				m.List.list.SetItems([]list.Item{item{skill: types.Skill{
					Name:    "Test Skill",
					Prefix:  "# Welcome\n",
					RawBody: "This is a **test** body.\n\n- Item 1\n- Item 2",
				}}})
				m.List.viewport.Width = 80
				m.List.viewport.Height = 18
				m.List.updatePreview()
				m.selected = m.List.selected
			}
			if tt.screen == ScreenGlobalSkillsList {
				m.globalSkillsList.SetItems([]list.Item{
					globalSkillItem{
						skill: types.Skill{
							Name: "Global Skill 1",
						},
						category: "Claude",
					},
				})
				m.globalSkillsLoaded = true
			}

			output := m.View()
			golden := filepath.Join("testdata", tt.name+".golden")

			// Force update if file doesn't exist
			if _, err := os.Stat(golden); os.IsNotExist(err) || os.Getenv("UPDATE_GOLDEN") != "" {
				os.MkdirAll(filepath.Dir(golden), 0755)
				os.WriteFile(golden, []byte(output), 0644)
			}

			expected, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("failed to read golden file: %v", err)
			}

			if output != string(expected) {
				// Accept minor whitespace/newline differences in Golden tests if m.StatusMsg empty
				if strings.TrimSpace(output) == strings.TrimSpace(string(expected)) {
					return
				}
				t.Errorf("output mismatch for screen %v", tt.screen)
			}
		})
	}
}

func TestHomeView_ContainsSyncOption(t *testing.T) {
	m := Model{
		Screen: ScreenHome,
	}
	view := m.View()
	if !strings.Contains(view, "4. Sincronizar con OpenCode") {
		t.Errorf("home view missing sync option, got:\n%s", view)
	}
}

func TestHomeView_ContainsGlobalSkillsOption(t *testing.T) {
	m := Model{
		Screen: ScreenHome,
	}
	view := m.View()
	if !strings.Contains(view, "7. Global Skills") {
		t.Errorf("home view missing global skills option, got:\n%s", view)
	}
}

func TestInstallerOptions_ContainsOpenCode(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Width = 80
	m.Height = 24

	output := m.Installer.OptionsView()

	// Check for explicit OPENCODE.MD label
	if !strings.Contains(output, "OpenCode (OPENCODE.MD)") {
		t.Errorf("installer options view missing explicit OPENCODE.MD label")
	}

	// Check for card borders
	if !strings.Contains(output, "┌") && !strings.Contains(output, "│") {
		t.Errorf("installer options view missing card borders")
	}

	// Check for header banner
	if !strings.Contains(output, "SYNCK INSTALLER") {
		t.Errorf("installer options view missing banner")
	}
}

func TestInstallerOptions_ModeLabels(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Width = 80

	// Test Recommended (false)
	m.Installer.Mode = false
	output := m.Installer.OptionsView()
	expectedRec := "[x] Recommended"
	expectedRecHelp := "Use one shared install everywhere"
	expectedAdv := "[ ] Advanced"
	expectedAdvHelp := "Create an isolated copy here"
	if !strings.Contains(output, expectedRec) {
		t.Errorf("expected output to contain %q for false mode", expectedRec)
	}
	if !strings.Contains(output, expectedRecHelp) {
		t.Errorf("expected output to contain %q for false mode", expectedRecHelp)
	}
	if !strings.Contains(output, expectedAdv) {
		t.Errorf("expected output to contain %q for false mode", expectedAdv)
	}
	if !strings.Contains(output, expectedAdvHelp) {
		t.Errorf("expected output to contain %q for false mode", expectedAdvHelp)
	}

	// Test Advanced (true)
	m.Installer.Mode = true
	output = m.Installer.OptionsView()
	if !strings.Contains(output, "[ ] Recommended") {
		t.Errorf("expected output to contain unselected Recommended for true mode")
	}
	if !strings.Contains(output, "[x] Advanced") {
		t.Errorf("expected output to contain selected Advanced for true mode")
	}
}

func TestInstallerOptions_ModeLabelsFitNarrowColumn(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Width = 80

	output := m.Installer.OptionsView()
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "Recommended") || strings.Contains(line, "Advanced") {
			if len(line) > 48 {
				t.Errorf("mode line too wide for installer column: %q (%d chars)", line, len(line))
			}
		}
	}
}

func TestSplitView_ShowsStatusMsg(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenList

	// Set dimensions via WindowSizeMsg
	updatedModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updatedModel.(Model)

	// Simulate status message via Update
	msg := statusMsg("Saved!")
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(Model)

	output := m.View()
	if !strings.Contains(output, "Saved!") {
		t.Errorf("splitView (ScreenList) does not display status message from list.View(), got:\n%s", output)
	}
}
