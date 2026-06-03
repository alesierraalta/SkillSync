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

	case ScreenDeleteConfirm:

		content = m.deleteConfirmView()

	case ScreenAgentEcosystem:

		content = m.agentEcosystemView()

	}

	if m.Screen == ScreenSyncing {
		return content
	}

	return content + "\n" + m.renderFooter()
}

// renderBinding renders a single key binding with a bold Primary-colored key
// and a Muted-colored description.
func renderBinding(b KeyBinding) string {
	return footerKeyStyle.Render(b.Key) + footerStyle.Render(": "+b.Help)
}

func (m Model) renderFooter() string {
	bindings := m.GetKeyBindings()

	var parts []string
	for _, b := range bindings {
		parts = append(parts, renderBinding(b))
	}

	separator := footerStyle.Render(" | ")
	footer := ""
	for i, part := range parts {
		footer += part
		if i < len(parts)-1 {
			footer += separator
		}
	}

	return footer
}

func (m Model) homeView() string {

	title := titleStyle.Render("Agent Skills TUI")

	opts := []string{
		"1. Instanciar ecosistema",
		"2. Gestionar skills",
		"3. Almacenamiento de skills",
		"4. Sincronizar con OpenCode",
		"5. Proyectos",
		"6. Agent Ecosystem",
	}

	var body string

	for i, opt := range opts {
		var line string
		if m.HomeCursor == i {
			line = selectedItemStyle.Render("> " + opt)
		} else {
			line = "  " + opt
		}
		body += line + "\n"
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

		labelStyle := labelMutedStyle
		if m.inputs[i].Focused() {
			labelStyle = labelActiveStyle
		}

		s += labelStyle.Render(label) + "\n"

		s += inputWrapStyle.Render(m.inputs[i].View()) + "\n"

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

	outputStyle := syncOutputStyle
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

func (m Model) deleteConfirmView() string {
	s := titleStyle.Render("Delete Skill") + "\n\n"
	s += docStyle.Render(m.deleteConfirm.View())
	return s
}

// agentEcosystemView renders the Agent Ecosystem screen.
// It shows a list of detected AI agent tools on the left and a detail panel
// for the selected agent (MCP servers + plugins) on the right.
// Pure rendering — no IO, no mutations.
func (m Model) agentEcosystemView() string {
	title := titleStyle.Render("Agent Ecosystem")
	s := title + "\n\n"

	if len(m.agentEcosystem) == 0 {
		s += "  No agents detected.\n"
		return s
	}

	for i, agent := range m.agentEcosystem {
		cursor := "  "
		if m.selectedAgent == i {
			cursor = "> "
		}

		// Status tag
		statusTag := ""
		switch agent.Status {
		case "unreadable":
			statusTag = " [unreadable]"
		case "present-only":
			statusTag = " [present-only]"
		}

		row := fmt.Sprintf("%s%s%s", cursor, agent.Name, statusTag)
		if m.selectedAgent == i {
			row = selectedItemStyle.Render(row)
		}
		s += row + "\n"

		// Detail panel for selected agent
		if m.selectedAgent == i {
			if len(agent.MCPServers) > 0 {
				s += "    MCP Servers:\n"
				for _, srv := range agent.MCPServers {
					s += fmt.Sprintf("      - %s (%s)\n", srv.Name, srv.Transport)
				}
			}
			if len(agent.Plugins) > 0 {
				s += "    Plugins:\n"
				for _, p := range agent.Plugins {
					enabled := ""
					if p.Enabled {
						enabled = " [enabled]"
					}
					version := ""
					if p.Version != "" {
						version = " v" + p.Version
					}
					s += fmt.Sprintf("      - %s%s%s\n", p.Name, version, enabled)
				}
			}
			if len(agent.MCPServers) == 0 && len(agent.Plugins) == 0 {
				s += "    No MCP servers or plugins configured.\n"
			}
		}
	}

	return s
}
