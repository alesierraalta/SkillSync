package ui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	var content string
	switch m.Screen {
	case ScreenList:
		content = m.splitView()
	case ScreenDetail:
		content = m.detailView()
	case ScreenSyncing:
		content = m.syncingView()
	case ScreenContentView:
		content = m.contentView()
	}

	return content + "\n" + m.renderFooter()
}

func (m Model) renderFooter() string {
	bindings := m.GetKeyBindings()
	footer := ""
	for i, b := range bindings {
		footer += fmt.Sprintf("%s: %s", b.Key, b.Help)
		if i < len(bindings)-1 {
			footer += " | "
		}
	}
	return footerStyle.Render(footer)
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
	for i := range m.inputs {
		indicator := "  "
		if m.inputs[i].Focused() {
			indicator = "> "
		}
		s += fmt.Sprintf("%s%s\n", indicator, m.inputs[i].View())
	}
	return s
}

func (m Model) syncingView() string {
	return titleStyle.Render("Syncing...") + "\n\n" + m.syncOutput
}

func (m Model) contentView() string {
	if m.selected == nil {
		return "No skill selected"
	}
	header := titleStyle.Render(fmt.Sprintf("Content: %s", m.selected.Name))
	return header + "\n\n" + m.viewport.View()
}
