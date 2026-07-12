package ui

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"skillsync/tui/internal/agentdetect"
	"skillsync/tui/internal/storage"
)

// TestStatusStyleFor_Table asserts the (foreground, glyph) tuple returned by
// statusStyleFor for every agentdetect.Status value the helper supports
// (StatusOK, StatusPresentOnly, StatusUnreadable, unknown). Pure unit test —
// no rendering, no TTY required.
// Spec: ST-1, ST-1.1..ST-1.4, GP-1, GP-1.1.
func TestStatusStyleFor_Table(t *testing.T) {
	tests := []struct {
		name           string
		status         agentdetect.Status
		wantForeground lipgloss.Color
		wantGlyph      string
	}{
		{
			name:           "StatusOK returns Success color + checkmark",
			status:         agentdetect.StatusOK,
			wantForeground: DefaultTheme.Success,
			wantGlyph:      "✓",
		},
		{
			name:           "StatusPresentOnly returns Warning color + warning glyph",
			status:         agentdetect.StatusPresentOnly,
			wantForeground: DefaultTheme.Warning,
			wantGlyph:      "⚠",
		},
		{
			name:           "StatusUnreadable returns Muted color + question mark",
			status:         agentdetect.StatusUnreadable,
			wantForeground: DefaultTheme.Muted,
			wantGlyph:      "?",
		},
		{
			name:           "Unknown status returns Muted color + empty glyph",
			status:         agentdetect.Status("bogus"),
			wantForeground: DefaultTheme.Muted,
			wantGlyph:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style, glyph := statusStyleFor(tt.status)

			// (a) Foreground color must match the expected theme token.
			gotForeground, ok := style.GetForeground().(lipgloss.Color)
			if !ok {
				t.Fatalf("expected style.GetForeground() to be lipgloss.Color, got %T (%v)",
					style.GetForeground(), style.GetForeground())
			}
			if gotForeground != tt.wantForeground {
				t.Errorf("foreground color: got %q, want %q", gotForeground, tt.wantForeground)
			}

			// (b) Glyph must match the expected character verbatim.
			if glyph != tt.wantGlyph {
				t.Errorf("glyph: got %q, want %q", glyph, tt.wantGlyph)
			}
		})
	}
}

// TestNoRawColorLiteralsInAgentEcoFiles is a focused regression guard for
// WU-1 of agent-eco-ui-as-skills: it scans internal/ui/agent_eco_status.go
// and internal/ui/view.go for raw lipgloss.Color("NNN") literals. Any
// violation means a new style var leaked a raw color value instead of
// consuming a DefaultTheme.* token. Complements (does not replace) the
// global TestNoRawColorLiteralsOutsideTheme in theme_static_test.go.
// Spec: GP-2, GP-2.1.
func TestNoRawColorLiteralsInAgentEcoFiles(t *testing.T) {
	re := regexp.MustCompile(`lipgloss\.Color\("[0-9]+\"\)`)
	targets := []string{
		"agent_eco_status.go",
		"view.go",
	}
	for _, name := range targets {
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if loc := re.FindIndex(src); loc != nil {
			t.Errorf("%s contains raw color literal %q — move to theme.go/DefaultTheme",
				name, src[loc[0]:loc[1]])
		}
	}
}

// newBannerBorderAgents returns a fixed 3-agent fixture used by the banner +
// border tests. Mixed statuses; at least one agent has both an MCP and a
// plugin so the detail panel is exercised.
func newBannerBorderAgents() []agentdetect.AgentInfo {
	return []agentdetect.AgentInfo{
		{
			Name:       "claude",
			Present:    true,
			Status:     agentdetect.StatusOK,
			MCPServers: []agentdetect.MCPServer{{Name: "context7", Transport: "stdio"}},
		},
		{
			Name:    "opencode",
			Present: true,
			Status:  agentdetect.StatusPresentOnly,
			Plugins: []agentdetect.Plugin{{Name: "synck-bridge", Enabled: true, Version: "1.0.0"}},
		},
		{
			Name:    "aider",
			Present: true,
			Status:  agentdetect.StatusUnreadable,
		},
	}
}

// TestAgentEcosystemView_BannerAndCardBorders asserts the height-aware card
// border pattern (mirror of installer_model.go:128-133):
//   - tall terminals (m.Height >= 24) render with a RoundedBorder (╭ / ╰)
//   - short terminals (m.Height < 24)  render with a NormalBorder  (┌ / └)
//     and contain zero RoundedBorder corner glyphs.
//
// Both subtests also assert the "AGENT ECOSYSTEM" banner is present and the
// "[1] DETECTED AGENTS" card title is present.
// Spec: LY-1, LY-1.1, LY-1.2, LY-1.3.
func TestAgentEcosystemView_BannerAndCardBorders(t *testing.T) {
	// RoundedBorder corner characters: ╭ ╮ ╰ ╯.
	roundedCorners := []string{"╭", "╮", "╰", "╯"}
	// NormalBorder corner characters: ┌ ┐ └ ┘.
	normalCorners := []string{"┌", "┐", "└", "┘"}

	containsAny := func(haystack string, needles []string) bool {
		for _, n := range needles {
			if strings.Contains(haystack, n) {
				return true
			}
		}
		return false
	}

	t.Run("tall_terminal_uses_rounded_border", func(t *testing.T) {
		m := NewModel(NewBackend(storage.NewService("")))
		m.Screen = ScreenAgentEcosystem
		m.Width = 120
		m.Height = 40
		m.agentEcosystem = newBannerBorderAgents()
		m.selectedAgent = 0

		view := m.agentEcosystemView()
		if !strings.Contains(view, "AGENT ECOSYSTEM") {
			t.Errorf("expected banner 'AGENT ECOSYSTEM' in tall-terminal render, got:\n%s", view)
		}
		if !strings.Contains(view, "[1] DETECTED AGENTS") {
			t.Errorf("expected card title '[1] DETECTED AGENTS' in tall-terminal render, got:\n%s", view)
		}
		if !containsAny(view, roundedCorners) {
			t.Errorf("expected at least one RoundedBorder corner (╭/╮/╰/╯) in tall-terminal render, got:\n%s", view)
		}
	})

	t.Run("short_terminal_uses_normal_border", func(t *testing.T) {
		m := NewModel(NewBackend(storage.NewService("")))
		m.Screen = ScreenAgentEcosystem
		m.Width = 120
		m.Height = 20
		m.agentEcosystem = newBannerBorderAgents()
		m.selectedAgent = 0

		view := m.agentEcosystemView()
		if !strings.Contains(view, "AGENT ECOSYSTEM") {
			t.Errorf("expected banner 'AGENT ECOSYSTEM' in short-terminal render, got:\n%s", view)
		}
		if !containsAny(view, normalCorners) {
			t.Errorf("expected at least one NormalBorder corner (┌/┐/└/┘) in short-terminal render, got:\n%s", view)
		}
		if containsAny(view, roundedCorners) {
			t.Errorf("expected NO RoundedBorder corners in short-terminal render, got:\n%s", view)
		}
	})

	t.Run("detail=empty", func(t *testing.T) {
		m := NewModel(NewBackend(storage.NewService("")))
		m.Screen = ScreenAgentEcosystem
		m.Width, m.Height = 120, 40
		m.agentEcosystem = newBannerBorderAgents()
		m.selectedAgent = 2

		view := m.agentEcosystemView()
		if !strings.Contains(view, "No MCP servers or plugins configured.") {
			t.Errorf("expected empty-detail hint in render, got:\n%s", view)
		}
		if strings.Contains(view, "context7") || strings.Contains(view, "synck-bridge") {
			t.Errorf("expected NO MCP/plugin names in empty-detail render, got:\n%s", view)
		}
	})

	t.Run("detail=plugin", func(t *testing.T) {
		m := NewModel(NewBackend(storage.NewService("")))
		m.Screen = ScreenAgentEcosystem
		m.Width, m.Height = 120, 40
		m.agentEcosystem = newBannerBorderAgents()
		m.selectedAgent = 1

		view := m.agentEcosystemView()
		if !strings.Contains(view, "1 Plugins") {
			t.Errorf("expected summary count '1 Plugins' in render, got:\n%s", view)
		}
	})
}

func TestAgentEcosystemView_Menus(t *testing.T) {
	t.Run("agent_menu", func(t *testing.T) {
		m := NewModel(NewBackend(storage.NewService("")))
		m.Screen = ScreenAgentMenu
		m.agentEcosystem = newBannerBorderAgents()
		m.selectedAgent = 1

		view := m.agentMenuView()
		if !strings.Contains(view, "Agent Menu: opencode") {
			t.Errorf("expected title in render, got:\n%s", view)
		}
		if !strings.Contains(view, "Plugins (1)") {
			t.Errorf("expected Plugins (1) in render, got:\n%s", view)
		}
	})

	t.Run("plugins_menu", func(t *testing.T) {
		m := NewModel(NewBackend(storage.NewService("")))
		m.Screen = ScreenPluginsMenu
		m.agentEcosystem = newBannerBorderAgents()
		m.selectedAgent = 1

		view := m.pluginsMenuView()
		if !strings.Contains(view, "Plugins for opencode") {
			t.Errorf("expected title in render, got:\n%s", view)
		}
		if !strings.Contains(view, "synck-bridge") {
			t.Errorf("expected plugin name in render, got:\n%s", view)
		}
	})
}
