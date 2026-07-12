package ui

import (
	"strings"
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

func TestHomeGuard_AllowsCursor6(t *testing.T) {
	m := newTestModel(t)
	// Press "down" 10 times — should stop at 6
	for i := 0; i < 10; i++ {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = newModel.(Model)
	}
	if m.HomeCursor != 6 {
		t.Errorf("HomeCursor after 10 downs: got %d, want 6", m.HomeCursor)
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

// ─── End-to-end: navigation + card layout + affordance + MCP names ──────────────

// TestIntegration_AgentEcosystem_EndToEnd_FullView exercises the full flow
// described in spec INT-1.1: load 3 agents, press j/j/k/enter, assert the
// selected row carries the affordance escape AND the rendered card body
// contains the selected agent's MCP server name. The "enter" press is a
// no-op for screen change in the Agent Eco flow (the screen stays the same).
// Spec: INT-1, INT-1.1, SL-1, SL-1.1, SL-2, SL-2.1, DP-1.1.
func TestIntegration_AgentEcosystem_EndToEnd_FullView(t *testing.T) {
	withANSI256(func() {
		m := newTestModel(t)
		m.Screen = ScreenAgentEcosystem

		agents := []agentdetect.AgentInfo{
			{
				Name:       "claude",
				Present:    true,
				Status:     agentdetect.StatusPresentOnly,
				MCPServers: []agentdetect.MCPServer{{Name: "filesystem", Transport: "stdio"}},
			},
			{
				Name:       "opencode",
				Present:    true,
				Status:     agentdetect.StatusOK,
				MCPServers: []agentdetect.MCPServer{{Name: "synck-tools", Transport: "stdio"}},
			},
			{
				Name:    "aider",
				Present: true,
				Status:  agentdetect.StatusUnreadable,
			},
		}
		m.agentEcosystem = agents

		// Navigation: j (1) -> j (2) -> k (1). Final selectedAgent == 1.
		for _, key := range []string{"j", "j", "k"} {
			newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
			m = newM.(Model)
		}
		if m.selectedAgent != 1 {
			t.Fatalf("selectedAgent after j/j/k: got %d, want 1", m.selectedAgent)
		}

		// Render and assert content before entering the menu.
		view := m.View()

		// (1) Banner — REQ-LY-1
		if !strings.Contains(view, "AGENT ECOSYSTEM") {
			t.Errorf("expected view to contain 'AGENT ECOSYSTEM' banner, got:\n%s", view)
		}

		// (2) Card title
		if !strings.Contains(view, "[1] DETECTED AGENTS") {
			t.Errorf("expected view to contain '[1] DETECTED AGENTS' card title, got:\n%s", view)
		}

		// (3) Selected agent's name on a line that carries the affordance escape.
		var opencodeLine string
		for _, line := range strings.Split(view, "\n") {
			if strings.Contains(line, "opencode") {
				opencodeLine = line
				break
			}
		}
		if opencodeLine == "" {
			t.Fatal("expected to find 'opencode' line in view, got:\n" + view)
		}
		if !strings.Contains(opencodeLine, ansiSelectedRow) {
			t.Errorf("expected 'opencode' line to carry affordance escape %q, got:\n%q",
				ansiSelectedRow, opencodeLine)
		}

		// (4) Selected agent's MCP servers count appears in the card body.
		if !strings.Contains(view, "1 MCP Servers") {
			t.Errorf("expected view to contain '1 MCP Servers', got:\n%s", view)
		}

		// Enter transitions to ScreenAgentMenu.
		newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = newM.(Model)
		if m.Screen != ScreenAgentMenu {
			t.Errorf("Screen after enter: got %v, want ScreenAgentMenu", m.Screen)
		}
	})
}
