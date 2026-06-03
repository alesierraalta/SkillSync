package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"skillsync/tui/internal/agentdetect"
	"skillsync/tui/internal/storage"
)

// withANSI256 temporarily sets the lipgloss color profile to ANSI256 for the
// duration of fn, then restores the original profile. This is necessary for
// affordance tests that assert ANSI escape sequences, since lipgloss strips
// colors when there is no terminal (CI / go test with no TTY).
func withANSI256(fn func()) {
	orig := lipgloss.DefaultRenderer().ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	defer lipgloss.SetColorProfile(orig)
	fn()
}

// ansiSelectedRow is the combined bold+Secondary(170) ANSI escape sequence emitted
// by selectedItemStyle (Bold=true, Foreground=170). Lipgloss combines bold and color
// in a single CSI sequence: \x1b[1;38;5;170m.
const ansiSelectedRow = "\x1b[1;38;5;170m"

// ansiFooterKey is the combined bold+Primary(205) ANSI escape sequence emitted
// by footerKeyStyle (Bold=true, Foreground=205). Lipgloss emits \x1b[1;38;5;205m.
const ansiFooterKey = "\x1b[1;38;5;205m"

// TestHomeView_SelectedItemStyled asserts that the selected row (cursor=1) carries
// the Secondary (170) color escape, and the non-selected row (cursor=0) does not.
// Spec: THEME-4.
func TestHomeView_SelectedItemStyled(t *testing.T) {
	withANSI256(func() {
		m := NewModel(NewBackend(storage.NewService("")))
		m.Screen = ScreenHome
		m.HomeCursor = 1

		output := m.homeView()
		lines := strings.Split(output, "\n")

		// Find the line for index 1 (second option: "2. Gestionar skills")
		var selectedLine, unselectedLine string
		for _, line := range lines {
			if strings.Contains(line, "2. Gestionar skills") {
				selectedLine = line
			}
			if strings.Contains(line, "1. Instanciar ecosistema") {
				unselectedLine = line
			}
		}

		if selectedLine == "" {
			t.Fatal("could not find selected option line '2. Gestionar skills' in homeView output")
		}
		if unselectedLine == "" {
			t.Fatal("could not find unselected option line '1. Instanciar ecosistema' in homeView output")
		}

		if !strings.Contains(selectedLine, ansiSelectedRow) {
			t.Errorf("selected row should contain Bold+Secondary escape %q, got line:\n%q", ansiSelectedRow, selectedLine)
		}
		if strings.Contains(unselectedLine, ansiSelectedRow) {
			t.Errorf("unselected row should NOT contain Bold+Secondary escape %q, got line:\n%q", ansiSelectedRow, unselectedLine)
		}
	})
}

// TestAgentEcosystemView_SelectedItemStyled asserts that the selected agent row (index 0)
// carries Secondary (170) color, and the second row (index 1) does not.
// Spec: THEME-4.
func TestAgentEcosystemView_SelectedItemStyled(t *testing.T) {
	withANSI256(func() {
		m := NewModel(NewBackend(storage.NewService("")))
		m.Screen = ScreenAgentEcosystem
		m.selectedAgent = 0
		m.agentEcosystem = []agentdetect.AgentInfo{
			{Name: "claude", Status: "present-only"},
			{Name: "opencode", Status: "present-only"},
		}

		output := m.agentEcosystemView()
		lines := strings.Split(output, "\n")

		var claudeLine, opencodeLine string
		for _, line := range lines {
			if strings.Contains(line, "claude") && !strings.Contains(line, "opencode") {
				if claudeLine == "" {
					claudeLine = line
				}
			}
			if strings.Contains(line, "opencode") {
				if opencodeLine == "" {
					opencodeLine = line
				}
			}
		}

		if claudeLine == "" {
			t.Fatal("could not find 'claude' agent line in agentEcosystemView output")
		}
		if opencodeLine == "" {
			t.Fatal("could not find 'opencode' agent line in agentEcosystemView output")
		}

		if !strings.Contains(claudeLine, ansiSelectedRow) {
			t.Errorf("selected agent row 'claude' should contain Bold+Secondary escape %q, got line:\n%q", ansiSelectedRow, claudeLine)
		}
		if strings.Contains(opencodeLine, ansiSelectedRow) {
			t.Errorf("unselected agent row 'opencode' should NOT contain Bold+Secondary escape %q, got line:\n%q", ansiSelectedRow, opencodeLine)
		}
	})
}

// TestFooterKeyEmphasis asserts that footer key labels are rendered with bold Primary color.
// Spec: THEME-5.
func TestFooterKeyEmphasis(t *testing.T) {
	withANSI256(func() {
		m := NewModel(NewBackend(storage.NewService("")))
		m.Screen = ScreenHome

		footer := m.renderFooter()

		// Footer must contain Bold+Primary (205) combined escape for the key labels
		if !strings.Contains(footer, ansiFooterKey) {
			t.Errorf("footer should contain Bold+Primary escape %q for key labels, got:\n%q", ansiFooterKey, footer)
		}
	})
}
