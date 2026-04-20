package ui

import (
	"os"
	"path/filepath"
	"testing"
	"skillsync/tui/internal/types"
	"github.com/charmbracelet/bubbles/list"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
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
				m.list.SetItems([]list.Item{item{skill: types.Skill{
					Name:    "Markdown Skill",
					Prefix:  "# Welcome\n",
					RawBody: "This is a **test** body.\n\n- Item 1\n- Item 2",
				}}})
				m.updatePreview()
			}
			if tt.screen == ScreenContentView {
				m.selected = &types.Skill{
					Name:    "Test Skill",
					Prefix:  "# Welcome\n",
					RawBody: "This is a **test** body.\n\n- Item 1\n- Item 2",
				}
				m.viewport.Width = 80
				m.viewport.Height = 18
				m.updatePreview()
			}

			output := m.View()
			golden := filepath.Join("testdata", tt.name+".golden")

			// Force update if file doesn't exist
			if _, err := os.Stat(golden); os.IsNotExist(err) || os.Getenv("UPDATE_GOLDEN") == "true" {
				os.MkdirAll(filepath.Dir(golden), 0755)
				os.WriteFile(golden, []byte(output), 0644)
			}

			expected, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("failed to read golden file: %v", err)
			}

			if output != string(expected) {
				t.Errorf("output mismatch for screen %v", tt.screen)
			}
		})
	}
}
