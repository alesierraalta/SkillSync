package ui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			MarginLeft(2).
			Foreground(DefaultTheme.Primary).
			Bold(true)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(DefaultTheme.Secondary).
				Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(DefaultTheme.Error).
			Bold(true)

	docStyle = lipgloss.NewStyle().Margin(1, 2)

	listStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(DefaultTheme.Muted)

	viewportStyle = lipgloss.NewStyle().
			Padding(0, 2)

	footerStyle = lipgloss.NewStyle().
			Foreground(DefaultTheme.Muted).
			Padding(0, 2)

	focusedTextareaStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(DefaultTheme.Primary)

	blurredTextareaStyle = lipgloss.NewStyle().
				Border(lipgloss.HiddenBorder())

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DefaultTheme.Accent).
			Padding(0, 1).
			MarginBottom(1)

	cardTitleStyle = lipgloss.NewStyle().
			Foreground(DefaultTheme.Primary).
			Bold(true).
			Padding(0, 1)

	hintStyle = lipgloss.NewStyle().
			Foreground(DefaultTheme.Muted).
			Italic(true)

	bannerStyle = lipgloss.NewStyle().
			Foreground(DefaultTheme.Secondary).
			Bold(true).
			MarginBottom(1)

	checkmarkStyle = lipgloss.NewStyle().
			Foreground(DefaultTheme.Success)

	searchBarFocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(DefaultTheme.Primary).
				Padding(0, 1)

	searchBarBlurred = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(DefaultTheme.Muted).
				Padding(0, 1)

	// Extracted from inline per-frame allocations in view.go detailView.
	// labelMutedStyle is used for unfocused input labels.
	labelMutedStyle = lipgloss.NewStyle().MarginLeft(2).Foreground(DefaultTheme.Muted)
	// labelActiveStyle is used for the currently focused input label.
	labelActiveStyle = lipgloss.NewStyle().MarginLeft(2).Foreground(DefaultTheme.Primary).Bold(true)
	// inputWrapStyle wraps each textarea input in detailView.
	inputWrapStyle = lipgloss.NewStyle().MarginLeft(2)
	// syncOutputStyle is used for live sync output text.
	syncOutputStyle = lipgloss.NewStyle().Foreground(DefaultTheme.SyncOutput)
	// warningNoticeStyle is used for license disclosure warnings.
	warningNoticeStyle = lipgloss.NewStyle().Bold(true).Foreground(DefaultTheme.Warning)
	// footerKeyStyle renders key labels in the footer with bold Primary color.
	footerKeyStyle = lipgloss.NewStyle().Foreground(DefaultTheme.Primary).Bold(true)
)
