package ui

import (
	"os"
	"regexp"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"skillsync/tui/internal/agentdetect"
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
	re := regexp.MustCompile(`lipgloss\.Color\("[0-9]+"\)`)
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
