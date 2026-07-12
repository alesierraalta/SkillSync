package ui

import (
	"github.com/charmbracelet/lipgloss"

	"skillsync/tui/internal/agentdetect"
)

// statusStyleFor returns the (style, glyph) pair to render next to an agent
// row, keyed by detection status. Pure function — no lipgloss.NewStyle calls
// per invocation, all styles are package-level in styles.go.
func statusStyleFor(s agentdetect.Status) (lipgloss.Style, string) {
	switch s {
	case agentdetect.StatusOK:
		return agentEcoStatusOKStyle, "✓"
	case agentdetect.StatusPresentOnly:
		return agentEcoStatusWarningStyle, "⚠"
	case agentdetect.StatusUnreadable:
		return agentEcoStatusMutedStyle, "?"
	default:
		return agentEcoStatusMutedStyle, ""
	}
}
