package ui

import (
	"strings"
	"testing"

	"skillsync/tui/internal/agentdetect"

	tea "github.com/charmbracelet/bubbletea"
)

// ─── Compile-level check: ScreenAgentEcosystem is a distinct Screen value ─────

func TestScreenAgentEcosystem_ConstantExists(t *testing.T) {
	// This test verifies at compile time that ScreenAgentEcosystem is defined and
	// is a Screen value distinct from all existing screens.
	var s Screen = ScreenAgentEcosystem
	existing := []Screen{
		ScreenHome,
		ScreenList,
		ScreenDetail,
		ScreenSyncing,
		ScreenContentView,
		ScreenInstaller,
		ScreenStorage,
		ScreenProjects,
		ScreenDeleteConfirm,
		ScreenLicenseDisclosure,
	}
	for _, e := range existing {
		if s == e {
			t.Errorf("ScreenAgentEcosystem (%d) collides with existing screen %d", s, e)
		}
	}
}

// ─── GetKeyBindings for AgentEcosystem screen ─────────────────────────────────

func TestGetKeyBindings_AgentEcosystem(t *testing.T) {
	backend := &MockAppService{}
	m := NewModel(backend)
	m.Screen = ScreenAgentEcosystem

	bindings := m.GetKeyBindings()
	if len(bindings) == 0 {
		t.Fatal("expected at least one key binding for ScreenAgentEcosystem")
	}

	// Must include esc/q and up/down hints
	allText := ""
	for _, b := range bindings {
		allText += b.Key + " " + b.Help + " "
	}
	if !strings.Contains(strings.ToLower(allText), "esc") {
		t.Errorf("expected esc binding, got bindings: %q", allText)
	}
	if !strings.Contains(strings.ToLower(allText), "up") && !strings.Contains(strings.ToLower(allText), "down") {
		t.Errorf("expected up/down binding, got bindings: %q", allText)
	}
}

// ─── View rendering tests ─────────────────────────────────────────────────────

func newMockWithAgents(agents []agentdetect.AgentInfo) *MockAppService {
	return &MockAppService{
		DetectAgentEcosystemFunc: func() ([]agentdetect.AgentInfo, error) {
			return agents, nil
		},
	}
}

func TestAgentEcosystemView_RendersAgents(t *testing.T) {
	agents := []agentdetect.AgentInfo{
		{Name: "Claude Code", Present: true, Status: agentdetect.StatusOK,
			MCPServers: []agentdetect.MCPServer{{Name: "claude-mcp", Transport: "stdio"}}},
		{Name: "Gemini CLI", Present: true, Status: agentdetect.StatusOK},
		{Name: "Cursor", Present: true, Status: agentdetect.StatusPresentOnly},
	}

	backend := newMockWithAgents(agents)
	m := NewModel(backend)
	m.Width = 120
	m.Height = 40
	m.Screen = ScreenAgentEcosystem
	m.agentEcosystem = agents

	view := m.View()
	for _, agent := range agents {
		if !strings.Contains(view, agent.Name) {
			t.Errorf("expected View() to contain agent name %q", agent.Name)
		}
	}
}

func TestAgentEcosystemView_EmptyState(t *testing.T) {
	backend := newMockWithAgents(nil)
	m := NewModel(backend)
	m.Width = 120
	m.Height = 40
	m.Screen = ScreenAgentEcosystem
	m.agentEcosystem = nil

	// Should not panic and should display some empty-state text
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view even in empty state")
	}
}

func TestAgentEcosystemView_ErrorState(t *testing.T) {
	backend := &MockAppService{
		DetectAgentEcosystemFunc: func() ([]agentdetect.AgentInfo, error) {
			return nil, nil
		},
	}
	m := NewModel(backend)
	m.Width = 120
	m.Height = 40
	m.Screen = ScreenAgentEcosystem
	m.StatusMsg = "Error detectando agentes: some error"
	m.agentEcosystem = nil

	// Should not panic
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view in error state")
	}
}

func TestAgentEcosystemView_DetailPanel(t *testing.T) {
	agents := []agentdetect.AgentInfo{
		{Name: "Claude Code", Present: true, Status: agentdetect.StatusOK},
		{
			Name:    "Gemini CLI",
			Present: true,
			Status:  agentdetect.StatusOK,
			MCPServers: []agentdetect.MCPServer{
				{Name: "fetch-server", Transport: "stdio"},
				{Name: "remote-server", Transport: "http"},
			},
		},
	}

	backend := newMockWithAgents(agents)
	m := NewModel(backend)
	m.Width = 120
	m.Height = 40
	m.Screen = ScreenAgentEcosystem
	m.agentEcosystem = agents
	m.selectedAgent = 1 // Gemini is selected

	view := m.View()
	if !strings.Contains(view, "fetch-server") {
		t.Errorf("expected View() to contain MCP server name 'fetch-server', got:\n%s", view)
	}
	if !strings.Contains(view, "remote-server") {
		t.Errorf("expected View() to contain MCP server name 'remote-server', got:\n%s", view)
	}
}

// ─── agentEcosystemLoadedMsg handling ─────────────────────────────────────────

func TestAgentEcosystemLoadedMsg_PopulatesModel(t *testing.T) {
	backend := &MockAppService{}
	m := NewModel(backend)

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

func TestAgentEcosystemLoadedMsg_ErrorSetsStatusMsg(t *testing.T) {
	backend := &MockAppService{}
	m := NewModel(backend)
	m.Screen = ScreenAgentEcosystem

	errMsg := agentEcosystemLoadedMsg{err: &errStub{msg: "detect failed"}}
	newModel, _ := m.Update(errMsg)
	updated := newModel.(Model)

	if !strings.Contains(updated.StatusMsg, "detect failed") {
		t.Errorf("expected StatusMsg to contain 'detect failed', got %q", updated.StatusMsg)
	}
}

// errStub is a minimal error for testing.
type errStub struct{ msg string }

func (e *errStub) Error() string { return e.msg }

// ─── Navigation: key bindings on ScreenAgentEcosystem ─────────────────────────

func TestHandleAgentEcosystemKeys_SelectDown(t *testing.T) {
	backend := &MockAppService{}
	m := NewModel(backend)
	m.Screen = ScreenAgentEcosystem
	m.agentEcosystem = []agentdetect.AgentInfo{
		{Name: "A", Present: true},
		{Name: "B", Present: true},
		{Name: "C", Present: true},
	}
	m.selectedAgent = 0

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	updated := newModel.(Model)
	if updated.selectedAgent != 1 {
		t.Errorf("selectedAgent after down: got %d, want 1", updated.selectedAgent)
	}
}

func TestHandleAgentEcosystemKeys_SelectUp(t *testing.T) {
	backend := &MockAppService{}
	m := NewModel(backend)
	m.Screen = ScreenAgentEcosystem
	m.agentEcosystem = []agentdetect.AgentInfo{
		{Name: "A", Present: true},
		{Name: "B", Present: true},
	}
	m.selectedAgent = 1

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	updated := newModel.(Model)
	if updated.selectedAgent != 0 {
		t.Errorf("selectedAgent after up: got %d, want 0", updated.selectedAgent)
	}
}

func TestHandleAgentEcosystemKeys_EscReturnsHome(t *testing.T) {
	backend := &MockAppService{}
	m := NewModel(backend)
	m.Screen = ScreenAgentEcosystem
	m.agentEcosystem = []agentdetect.AgentInfo{{Name: "A", Present: true}}
	m.selectedAgent = 0

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated := newModel.(Model)
	if updated.Screen != ScreenHome {
		t.Errorf("Screen after esc: got %v, want ScreenHome", updated.Screen)
	}
	if updated.HomeCursor != 5 {
		t.Errorf("HomeCursor after esc: got %d, want 5", updated.HomeCursor)
	}
}
