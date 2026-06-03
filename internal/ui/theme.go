package ui

import "github.com/charmbracelet/lipgloss"

// Theme holds the semantic color tokens for the TUI.
// All ANSI-256 color values in internal/ui must flow from DefaultTheme.
// Layer 1 (tokens): theme.go is the sole owner of raw lipgloss.Color literals.
// Layer 2 (styles): styles.go consumes DefaultTheme.* tokens.
// Layer 3 (views): view files consume package-level style vars from styles.go.
type Theme struct {
	Primary    lipgloss.Color // 205 hot pink  — title, focus borders, card titles, focused label
	Secondary  lipgloss.Color // 170 orchid    — selected item, banner, active accents
	Accent     lipgloss.Color // 62  cornflower — card borders (distinct structural hue)
	Success    lipgloss.Color // 42  green      — checkmark
	Warning    lipgloss.Color // 208 orange     — license notice
	Error      lipgloss.Color // 196 red        — errors
	Muted      lipgloss.Color // 240 gray       — footer, hints, unfocused borders/labels
	Text       lipgloss.Color // ""  terminal default — body text (no foreground escape)
	SyncOutput lipgloss.Color // 212 salmon     — live sync output
}

// DefaultTheme is the single source of truth for all ANSI-256 color values.
var DefaultTheme = Theme{
	Primary:    lipgloss.Color("205"),
	Secondary:  lipgloss.Color("170"),
	Accent:     lipgloss.Color("62"),
	Success:    lipgloss.Color("42"),
	Warning:    lipgloss.Color("208"),
	Error:      lipgloss.Color("196"),
	Muted:      lipgloss.Color("240"),
	Text:       lipgloss.Color(""),
	SyncOutput: lipgloss.Color("212"),
}
