package ui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	switch m.Screen {
	case ScreenList:
		return m.splitView()
	case ScreenDetail:
		return m.detailView()
	case ScreenSyncing:
		return m.syncingView()
	case ScreenContentView:
		return m.contentView()
	}
	return ""
}

func (m Model) splitView() string {
	listWidth := int(float64(m.Width) * 0.4)
	previewWidth := m.Width - listWidth

	left := listStyle.Width(listWidth).Render(m.list.View())
	right := viewportStyle.Width(previewWidth).Render(m.viewport.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m Model) detailView() string {
	if m.selected == nil {
		return "No skill selected"
	}
	var s string
	s += titleStyle.Render(fmt.Sprintf("Editing Skill: %s", m.selected.Name)) + "\n\n"
	for _, i := range m.inputs {
		s += i.View() + "\n"
	}
	s += "\n(esc: back, ctrl+s: save)\n"
	return s
}

func (m Model) syncingView() string {
	return titleStyle.Render("Syncing...") + "\n\n" + m.syncOutput + "\n\n(esc: back)"
}

func (m Model) contentView() string {
	if m.selected == nil {
		return "No skill selected"
	}
	header := titleStyle.Render(fmt.Sprintf("Content: %s", m.selected.Name))
	footer := fmt.Sprintf("\n\n(esc: back, j/k: scroll)")
	return header + "\n\n" + m.viewport.View() + footer
}
