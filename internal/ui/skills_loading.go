package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// newSkillsSpinner builds the spinner shown while the skill list loads.
func newSkillsSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(DefaultTheme.Primary)
	return s
}

// beginLoadSkills marks the list as loading and kicks off both the concurrent
// skill load and the spinner animation. Callers that navigate to (or refresh)
// the manage-skills list should use this instead of loadSkills() directly so
// the user gets a loading indicator with many skills.
func (m *Model) beginLoadSkills() tea.Cmd {
	m.skillsLoading = true
	return tea.Batch(m.loadSkills(), m.spinner.Tick)
}

// loadingView renders the spinner + message shown while skills are loading.
func (m Model) loadingView() string {
	return docStyle.Render(
		titleStyle.Render("Skillsync TUI") + "\n\n" +
			fmt.Sprintf("%s Cargando skills...", m.spinner.View()),
	)
}
