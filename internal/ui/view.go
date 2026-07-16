package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"skillsync/tui/internal/agentdetect"
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
	case ScreenSkillFiles:
		content = m.viewSkillFiles()

	case ScreenInstaller:

		content = m.Installer.View()

	case ScreenStorage:

		content = m.storageView()

	case ScreenBundleImport:

		content = m.bundleImportView()

	case ScreenProjects:

		content = m.projectsView()

	case ScreenDeleteConfirm:

		content = m.deleteConfirmView()

	case ScreenAgentEcosystem:
		content = m.agentEcosystemView()
	case ScreenAgentMenu:
		content = m.agentMenuView()
	case ScreenPluginsMenu:
		content = m.pluginsMenuView()
	case ScreenMCPServersMenu:
		content = m.mcpServersMenuView()
	case ScreenGlobalSkillsCats:
		content = m.viewGlobalSkillsCats()
	case ScreenGlobalSkillsList:
		content = m.viewGlobalSkillsList()
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

	// Trim trailing whitespace: footerStyle's horizontal padding leaves
	// trailing spaces on the last binding, which are visually noise and
	// trip `git diff --check` in the golden files.
	return strings.TrimRight(footer, " \t")
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
		"7. Global Skills",
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
	if m.selectMode {
		return s + m.storageSelectView()
	}
	hint := hintStyle.Render("space: seleccionar · m: importar bundle")
	return s + hint + "\n\n" + docStyle.Render(m.storageList.View())
}

// storageSelectView renders the vault as a checklist with a cursor and
// [x]/[ ] markers while multi-select is active. Rows are windowed to the
// available height so a large vault never overflows the terminal. Filtering
// is disabled on storageList (see NewModel), so storageList.Index() stays
// aligned with m.storedSkills here.
func (m Model) storageSelectView() string {
	var b strings.Builder
	b.WriteString(hintStyle.Render("space: (des)seleccionar · e: exportar selección · m: importar · esc: salir") + "\n\n")

	idx := m.storageList.Index()
	start, end := windowBounds(idx, len(m.storedSkills), m.Height-8)
	for i := start; i < end; i++ {
		name := m.storedSkills[i].Metadata.SkillName
		cursor := "  "
		if i == idx {
			cursor = "> "
		}
		marker := "[ ]"
		if m.vaultSelected[name] {
			marker = "[x]"
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, marker, name))
	}
	b.WriteString(fmt.Sprintf("\n%d seleccionada(s) de %d", len(m.selectedVaultNames()), len(m.storedSkills)))
	return docStyle.Render(b.String())
}

// windowBounds returns a [start,end) slice window of size maxRows that keeps
// index idx visible. If maxRows <= 0 or exceeds total, the full range is used.
func windowBounds(idx, total, maxRows int) (int, int) {
	if maxRows <= 0 || maxRows >= total {
		return 0, total
	}
	start := 0
	if idx >= maxRows {
		start = idx - maxRows + 1
	}
	end := start + maxRows
	if end > total {
		end = total
		start = end - maxRows
	}
	return start, end
}

// bundleImportView renders the .skillsync import screen with a path input.
func (m Model) bundleImportView() string {
	s := titleStyle.Render("Importar Bundle .skillsync") + "\n\n"
	body := "Ingresá la ruta del bundle a instalar en este proyecto:\n\n" +
		m.bundleImportIn.View() + "\n\n" +
		hintStyle.Render("enter: importar · esc: cancelar")
	return s + docStyle.Render(body)
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
// Mirrors InstallerModel.OptionsView card layout (banner + card + selected
// row + detail panel). All styles consumed are package-level vars in
// styles.go; zero lipgloss.NewStyle() per call. Pure rendering — no IO, no
// mutations. Spec: LY-1, LY-1.1, LY-1.2, LY-1.3, SL-1, SL-2, DP-1, DP-2.
func (m Model) agentEcosystemView() string {
	// Card body width: leave 2 columns of breathing room for the border.
	width := m.Width - 2
	if width < 0 {
		width = 0
	}

	// Height-aware card border, identical to installer_model.go:128-133.
	var localCardStyle = cardStyle
	if m.Height < 24 {
		localCardStyle = localCardStyle.Border(lipgloss.NormalBorder()).MarginBottom(0)
	} else {
		localCardStyle = localCardStyle.Border(lipgloss.RoundedBorder()).MarginBottom(1)
	}

	banner := bannerStyle.Render("AGENT ECOSYSTEM")

	// Empty state — no card, just the banner + a hint line.
	if len(m.agentEcosystem) == 0 {
		return banner + "\n" + hintStyle.Render("  No agents detected.") + "\n"
	}

	// Build the card body as a []string and JoinVertical to assemble.
	// This is the same primitive InstallerModel.OptionsView uses, and it
	// keeps the per-line assembly allocation-cheap (each entry is already a
	// fully-rendered string). spec LY-1.
	lines := make([]string, 0, len(m.agentEcosystem)+2)
	lines = append(lines, cardTitleStyle.Render("[1] DETECTED AGENTS"))
	for i, agent := range m.agentEcosystem {
		lines = append(lines, m.renderAgentRow(agent, i == m.selectedAgent, width))
	}
	if m.selectedAgent >= 0 && m.selectedAgent < len(m.agentEcosystem) {
		lines = append(lines, m.renderAgentDetail(m.agentEcosystem[m.selectedAgent], width))
	}
	body := lipgloss.JoinVertical(lipgloss.Left, lines...)

	return banner + "\n" + localCardStyle.Width(width).Render(body) + "\n"
}

// renderAgentRow renders one agent row inside the Agent Ecosystem card.
// When selected, the whole row is wrapped in selectedItemStyle — the
// affordance escape lands on the same line as the agent name so the line-scan
// invariant in TestAgentEcosystemView_SelectedItemStyled holds.
// Pure rendering — no IO, no mutations.
func (m Model) renderAgentRow(agent agentdetect.AgentInfo, selected bool, width int) string {
	_ = width // reserved for future row-width truncation; not used today.
	cursor := "  "
	if selected {
		cursor = "> "
	}
	statusStyle, glyph := statusStyleFor(agent.Status)
	// Plain string concat (not lipgloss.JoinHorizontal) — the cursor, glyph
	// and name are all single-line, so the visual alignment is implicit.
	// This keeps the alloc footprint to 1 (statusStyle.Render) + 1 (selectedItemStyle.Render)
	// per row instead of the JoinHorizontal slice + width computation.
	row := cursor + statusStyle.Render(glyph) + " " + agent.Name
	if selected {
		row = selectedItemStyle.Render(row)
	}
	return row
}

func (m Model) renderAgentDetail(agent agentdetect.AgentInfo, width int) string {
	_ = width
	if len(agent.MCPServers) == 0 && len(agent.Plugins) == 0 {
		return hintStyle.Render("  No MCP servers or plugins configured.")
	}

	lines := make([]string, 0, 2)
	if len(agent.MCPServers) > 0 {
		lines = append(lines, fmt.Sprintf("  %d MCP Servers", len(agent.MCPServers)))
	}
	if len(agent.Plugins) > 0 {
		lines = append(lines, fmt.Sprintf("  %d Plugins", len(agent.Plugins)))
	}
	return hintStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m Model) agentMenuView() string {
	if len(m.agentEcosystem) == 0 {
		return "No agent selected."
	}
	agent := m.agentEcosystem[m.selectedAgent]
	title := titleStyle.Render(fmt.Sprintf("Agent Menu: %s", agent.Name))

	opts := []string{
		fmt.Sprintf("Plugins (%d)", len(agent.Plugins)),
		fmt.Sprintf("MCP Servers (%d)", len(agent.MCPServers)),
	}

	var body string
	for i, opt := range opts {
		line := "  " + opt
		if m.agentMenuCursor == i {
			line = selectedItemStyle.Render("> " + opt)
		}
		body += line + "\n"
	}

	footer := hintStyle.Render(fmt.Sprintf("\n  Config Path: %s\n  (Press 'o' to open in explorer)", agent.ConfigPath))
	return title + "\n\n" + body + footer
}

func (m Model) pluginsMenuView() string {
	if len(m.agentEcosystem) == 0 {
		return "No agent selected."
	}
	agent := m.agentEcosystem[m.selectedAgent]
	title := titleStyle.Render(fmt.Sprintf("Plugins for %s", agent.Name))

	footer := hintStyle.Render(fmt.Sprintf("\n  Config Path: %s\n  (Press 'o' to open in explorer)", agent.ConfigPath))

	if len(agent.Plugins) == 0 {
		return title + "\n\n" + hintStyle.Render("  No plugins configured.") + footer
	}

	var body string
	for i, p := range agent.Plugins {
		marker := "[ ]"
		if p.Enabled {
			marker = checkmarkStyle.Render("[x]")
		}
		version := ""
		if p.Version != "" {
			version = " v" + p.Version
		}

		line := fmt.Sprintf("%s %s%s", marker, p.Name, version)
		if m.pluginsMenuCursor == i {
			line = selectedItemStyle.Render("> " + line)
		} else {
			line = "  " + line
		}
		body += line + "\n"
	}

	return title + "\n\n" + body + footer
}

func (m Model) mcpServersMenuView() string {
	if len(m.agentEcosystem) == 0 {
		return "No agent selected."
	}
	agent := m.agentEcosystem[m.selectedAgent]
	title := titleStyle.Render(fmt.Sprintf("MCP Servers for %s", agent.Name))

	footer := hintStyle.Render(fmt.Sprintf("\n  Config Path: %s\n  (Press 'o' to open in explorer)", agent.ConfigPath))

	if len(agent.MCPServers) == 0 {
		return title + "\n\n" + hintStyle.Render("  No MCP servers configured.") + footer
	}

	var body string
	for i, s := range agent.MCPServers {
		marker := "[ ]"
		// MCPServer doesn't have Enabled, default to just listing them or assume active

		line := fmt.Sprintf("%s %s", marker, s.Name)
		if m.mcpServersMenuCursor == i {
			line = selectedItemStyle.Render("> " + line)
		} else {
			line = "  " + line
		}
		body += line + "\n"
	}

	return title + "\n\n" + body + footer
}

func (m Model) viewGlobalSkillsCats() string {
	title := titleStyle.Render("Global Skills - Categories")

	opts := globalSkillCategories

	var body string
	for i, opt := range opts {
		var line string
		if m.globalCategoryCursor == i {
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

func (m Model) viewGlobalSkillsList() string {
	if !m.globalSkillsLoaded {
		return titleStyle.Render("Global Skills: "+m.globalCategory) + "\n\n  Buscando..."
	}
	if m.globalSkillsErr != nil {
		title := titleStyle.Render(m.globalSkillsList.Title)
		errLine := errorStyle.Render(fmt.Sprintf("  Error: %v", m.globalSkillsErr))
		hint := hintStyle.Render("  Press esc to go back and pick a different category.")
		return title + "\n\n" + errLine + "\n\n" + hint
	}
	if len(m.globalSkillsList.Items()) == 0 {
		return titleStyle.Render(m.globalSkillsList.Title) + "\n\n  No se encontraron skills en esta categoría."
	}
	return docStyle.Render(m.globalSkillsList.View())
}
