package ui

import (
	"testing"

	"skillsync/tui/internal/agentdetect"

	tea "github.com/charmbracelet/bubbletea"
)

// helper: build a minimal initialized model for integration tests
func newTestModel(t *testing.T) Model {
	t.Helper()
	backend := &MockAppService{}
	m := NewModel(backend)
	// Simulate window size
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return updated.(Model)
}

// ─── Home key "6" → ScreenAgentEcosystem ─────────────────────────────────────

func TestHomeKey6_TransitionsToAgentEcosystem(t *testing.T) {
	m := newTestModel(t)
	// Simulate pressing "6"
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("6")})
	updated := newModel.(Model)
	if updated.Screen != ScreenAgentEcosystem {
		t.Errorf("Screen: got %v, want ScreenAgentEcosystem", updated.Screen)
	}
	if updated.HomeCursor != 5 {
		t.Errorf("HomeCursor: got %d, want 5", updated.HomeCursor)
	}
}

// ─── Enter at cursor 5 → ScreenAgentEcosystem ─────────────────────────────────

func TestHomeEnter_Cursor5_TransitionsToAgentEcosystem(t *testing.T) {
	m := newTestModel(t)
	m.HomeCursor = 5

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := newModel.(Model)
	if updated.Screen != ScreenAgentEcosystem {
		t.Errorf("Screen: got %v, want ScreenAgentEcosystem", updated.Screen)
	}
}

// ─── Guard: cursor stops at 5 ─────────────────────────────────────────────────

func TestHomeGuard_AllowsCursor5(t *testing.T) {
	m := newTestModel(t)
	// Press "down" 10 times — should stop at 5 (not 4 as before the guard change)
	for i := 0; i < 10; i++ {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(Model)
	}
	if m.HomeCursor != 5 {
		t.Errorf("HomeCursor after 10 downs: got %d, want 5", m.HomeCursor)
	}
}

// ─── Regression: keys 1–5 still navigate to original screens ────────────────────

func TestHomeNavigation_Keys1Through5_Unaffected(t *testing.T) {
	tests := []struct {
		key        string
		wantScreen Screen
	}{
		{"1", ScreenInstaller},
		{"2", ScreenList},
		{"3", ScreenStorage},
		// key "4" triggers sync (stays on ScreenHome with a cmd), not a screen change
		{"5", ScreenProjects},
	}

	for _, tt := range tests {
		t.Run("key "+tt.key, func(t *testing.T) {
			m := newTestModel(t)
			newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			updated := newModel.(Model)
			if updated.Screen != tt.wantScreen {
				t.Errorf("key %q: Screen got %v, want %v", tt.key, updated.Screen, tt.wantScreen)
			}
		})
	}
}

// ─── agentEcosystemLoadedMsg populates model ─────────────────────────────────────

func TestIntegration_AgentEcosystemLoadedMsg_PopulatesModel(t *testing.T) {
	m := newTestModel(t)
	m.Screen = ScreenAgentEcosystem

	agents := []agentdetect.AgentInfo{
		{Name: "Claude Code", Present: true, Status: agentdetect.StatusOK},
		{Name: "Gemini CLI", Present: true, Status: agentdetect.StatusOK},
	}

	newModel, _ := m.Update(agentEcosystemLoadedMsg{list: agents})
	updated := newModel.(Model)

	if len(updated.agentEcosystem) != 2 {
		t.Errorf("agentEcosystem len: got %d, want 2", len(updated.agentEcosystem))
	}
	if updated.selectedAgent != 0 {
		t.Errorf("selectedAgent: got %d, want 0", updated.selectedAgent)
	}
}

// ─── Esc from AgentEcosystem returns to Home ─────────────────────────────────────

func TestHandleAgentEcosystemEscReturnsHome(t *testing.T) {
	m := newTestModel(t)
	m.Screen = ScreenAgentEcosystem
	m.agentEcosystem = []agentdetect.AgentInfo{{Name: "Test", Present: true}}

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated := newModel.(Model)

	if updated.Screen != ScreenHome {
		t.Errorf("Screen: got %v, want ScreenHome", updated.Screen)
	}
	if updated.HomeCursor != 5 {
		t.Errorf("HomeCursor: got %d, want 5", updated.HomeCursor)
	}
}
