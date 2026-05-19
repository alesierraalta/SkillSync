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

	case ScreenHome:

		content = m.homeView()

	case ScreenList:

		content = m.List.View()

	case ScreenDetail:

		content = m.detailView()

	case ScreenSyncing:

		content = m.syncingView()

	case ScreenContentView:

		content = m.contentView()

	case ScreenInstaller:

		content = m.Installer.View()

	case ScreenStorage:

		content = m.storageView()

	case ScreenProjects:

		content = m.projectsView()

	}

	if m.Screen == ScreenSyncing {
		return content
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

func (m Model) homeView() string {

	title := titleStyle.Render("Agent Skills TUI")

	opts := []string{
		"1. Instanciar ecosistema",
		"2. Gestionar skills",
		"3. Almacenamiento de skills",
		"4. Sincronizar con OpenCode",
		"5. Proyectos",
	}

	var body string

	for i, opt := range opts {

		cursor := "  "

		if m.HomeCursor == i {

			cursor = "> "

		}

		body += fmt.Sprintf("%s%s\n", cursor, opt)

	}

	if m.StatusMsg != "" {

		body += "\n" + m.StatusMsg + "\n"

	}

	return title + "\n\n" + body

}

func (m Model) splitView() string {
	return m.List.View()
}

func (m Model) detailView() string {

	if m.selected == nil {

		return "No skill selected"

	}

	var s string

	s += titleStyle.Render(fmt.Sprintf("Editing Skill: %s", m.selected.Name)) + "\n\n"

	labels := []string{"Description", "Scope", "Content (SKILL.md)"}

	for i := range m.inputs {

		label := labels[i]

		labelStyle := lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("240"))

		if m.inputs[i].Focused() {

			labelStyle = labelStyle.Foreground(lipgloss.Color("205")).Bold(true)

		}

		s += labelStyle.Render(label) + "\n"

		s += lipgloss.NewStyle().MarginLeft(2).Render(m.inputs[i].View()) + "\n"

	}

	return s

}

func (m Model) syncingView() string {
	var titleText string
	switch m.PrevScreen {
	case ScreenList, ScreenContentView:
		titleText = "Syncing Skills..."
	case ScreenStorage:
		titleText = "Installing Skill..."
	default:
		titleText = "Installing Ecosistema..."
	}
	title := titleStyle.Render(titleText)

	outputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	if m.SyncFailed {
		outputStyle = errorStyle
	}

	content := m.syncOutput
	if m.SyncFinished {
		if m.syncReport != nil {
			content = "✅ Sync Successful!\n\n"
			content += fmt.Sprintf("Skills synced: %d\n", len(m.syncReport.Changes))
			for _, change := range m.syncReport.Changes {
				content += fmt.Sprintf("  • %s (%s)\n", change.Path, change.Status)
			}
		} else if m.err != nil {
			content = errorStyle.Render("❌ Sync Failed") + "\n\n"
			content += "Error Details:\n"
			content += m.err.Error()
		}
	}

	currentTask := outputStyle.Render(content)

	s := title + "\n\n"
	s += currentTask + "\n\n"

	if m.SyncFailed || m.SyncFinished {
		s += "  esc: back"
	} else {
		s += m.Progress.View()
	}

	return s
}

func (m Model) contentView() string {

	if m.selected == nil {

		return "No skill selected"

	}

	header := titleStyle.Render(fmt.Sprintf("Content: %s", m.selected.Name))

	return header + "\n\n" + m.List.viewport.View()

}

func (m Model) storageView() string {
	s := titleStyle.Render("Almacenamiento Global de Skills") + "\n\n"
	if len(m.storedSkills) == 0 {
		return s + lipgloss.NewStyle().MarginLeft(4).Render("No hay skills almacenadas.\nPresioná 's' en la vista de gestión para guardar una.")
	}
	return s + docStyle.Render(m.storageList.View())
}

func (m Model) projectsView() string {
	s := titleStyle.Render("Proyectos Sincronizados") + "\n\n"
	if len(m.projectList.Items()) == 0 {
		return s + lipgloss.NewStyle().MarginLeft(4).Render("No se encontraron proyectos. Presioná '4' para sincronizar este proyecto y registrarlo.")
	}
	return s + docStyle.Render(m.projectList.View())
}
