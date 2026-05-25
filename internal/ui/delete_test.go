package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDeleteConfirm_Init(t *testing.T) {
	m := DeleteConfirmModel{}
	cmd := m.Init()
	if cmd != nil {
		t.Fatal("Init() should return nil command")
	}
}

func TestDeleteConfirm_UpdateY(t *testing.T) {
	m := DeleteConfirmModel{skillName: "test-skill"}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	if !updated.confirmed {
		t.Error("expected confirmed=true after 'y'")
	}
	if !updated.deleting {
		t.Error("expected deleting=true after 'y'")
	}

	// cmd should be non-nil (the delete command)
	if cmd == nil {
		t.Fatal("expected non-nil cmd after 'y'")
	}
}

func TestDeleteConfirm_UpdateN(t *testing.T) {
	m := DeleteConfirmModel{skillName: "test-skill"}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	if updated.confirmed {
		t.Error("expected confirmed=false after 'n'")
	}
	if updated.deleting {
		t.Error("expected deleting=false after 'n'")
	}
	if cmd != nil {
		t.Error("expected nil cmd after 'n'")
	}
}

func TestDeleteConfirm_UpdateEsc(t *testing.T) {
	m := DeleteConfirmModel{skillName: "test-skill"}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if updated.confirmed {
		t.Error("expected confirmed=false after 'esc'")
	}
	if updated.deleting {
		t.Error("expected deleting=false after 'esc'")
	}
	if cmd != nil {
		t.Error("expected nil cmd after 'esc'")
	}
}

func TestDeleteConfirm_UpdateCapitalY(t *testing.T) {
	m := DeleteConfirmModel{skillName: "test-skill"}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}})

	if !updated.confirmed {
		t.Error("expected confirmed=true after 'Y'")
	}
}

func TestDeleteConfirm_UpdateCapitalN(t *testing.T) {
	m := DeleteConfirmModel{skillName: "test-skill"}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})

	if updated.confirmed {
		t.Error("expected confirmed=false after 'N'")
	}
}

func TestDeleteConfirm_UpdateFinishedMsg(t *testing.T) {
	m := DeleteConfirmModel{skillName: "test-skill", deleting: true}
	updated, _ := m.Update(deleteSkillFinishedMsg{name: "test-skill", err: nil})

	if updated.deleting {
		t.Error("expected deleting=false after success msg")
	}
	if !updated.success {
		t.Error("expected success=true after success msg")
	}

	// Test error case
	m2 := DeleteConfirmModel{skillName: "test-skill", deleting: true}
	updated2, _ := m2.Update(deleteSkillFinishedMsg{name: "test-skill", err: fmt.Errorf("permission denied")})

	if updated2.deleting {
		t.Error("expected deleting=false after error msg")
	}
	if updated2.err == nil {
		t.Fatal("expected err to be set after error msg")
	}
	if !strings.Contains(updated2.err.Error(), "permission denied") {
		t.Errorf("expected 'permission denied' error, got %v", updated2.err)
	}
}

func TestDeleteConfirm_ViewConfirm(t *testing.T) {
	m := DeleteConfirmModel{skillName: "test-skill"}
	output := m.View()

	expected := "Remove skill 'test-skill'? This cannot be undone. [y/N]"
	if output != expected {
		t.Errorf("unexpected confirm view:\nexpected: %q\n     got: %q", expected, output)
	}
}

func TestDeleteConfirm_ViewDeleting(t *testing.T) {
	m := DeleteConfirmModel{skillName: "test-skill", deleting: true}
	output := m.View()

	expected := "Deleting skill 'test-skill'..."
	if output != expected {
		t.Errorf("unexpected deleting view:\nexpected: %q\n     got: %q", expected, output)
	}
}

func TestDeleteConfirm_ViewSuccess(t *testing.T) {
	m := DeleteConfirmModel{skillName: "test-skill", success: true}
	output := m.View()

	expected := "Skill 'test-skill' deleted."
	if output != expected {
		t.Errorf("unexpected success view:\nexpected: %q\n     got: %q", expected, output)
	}
}

func TestDeleteConfirm_ViewError(t *testing.T) {
	m := DeleteConfirmModel{skillName: "test-skill", err: fmt.Errorf("permission denied")}
	output := m.View()

	expected := "Error deleting skill 'test-skill': permission denied"
	if output != expected {
		t.Errorf("unexpected error view:\nexpected: %q\n     got: %q", expected, output)
	}
}

// Golden tests — test full Model.View() output for ScreenDeleteConfirm
func TestDeleteConfirmViewGolden(t *testing.T) {
	if os.Getenv("UPDATE_GOLDEN") != "" {
		t.Log("UPDATE_GOLDEN is set — golden files will be updated")
	}

	tests := []struct {
		name       string
		setupModel func() Model
		goldenFile string
	}{
		{
			name: "confirm",
			setupModel: func() Model {
				m := Model{Screen: ScreenDeleteConfirm}
				m.deleteConfirm.skillName = "test-skill"
				return m
			},
			goldenFile: "delete-confirm.golden",
		},
		{
			name: "success",
			setupModel: func() Model {
				m := Model{Screen: ScreenDeleteConfirm}
				m.deleteConfirm.skillName = "test-skill"
				m.deleteConfirm.success = true
				return m
			},
			goldenFile: "delete-success.golden",
		},
		{
			name: "error",
			setupModel: func() Model {
				m := Model{Screen: ScreenDeleteConfirm}
				m.deleteConfirm.skillName = "test-skill"
				m.deleteConfirm.err = fmt.Errorf("permission denied")
				return m
			},
			goldenFile: "delete-error.golden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setupModel()
			m.Width = 80
			m.Height = 24

			output := m.View()
			golden := filepath.Join("testdata", tt.goldenFile)

			if _, err := os.Stat(golden); os.IsNotExist(err) || os.Getenv("UPDATE_GOLDEN") != "" {
				os.MkdirAll(filepath.Dir(golden), 0755)
				os.WriteFile(golden, []byte(output), 0644)
			}

			expected, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("failed to read golden file %s: %v", golden, err)
			}

			if output != string(expected) {
				// Accept whitespace-only differences
				if strings.TrimSpace(output) == strings.TrimSpace(string(expected)) {
					return
				}
				t.Errorf("output mismatch for %s:\nexpected:\n%s\n\ngot:\n%s", tt.goldenFile, string(expected), output)
			}
		})
	}
}
