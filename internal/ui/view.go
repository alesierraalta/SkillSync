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

		content = m.splitView()

	case ScreenDetail:

		content = m.detailView()

	case ScreenSyncing:

		content = m.syncingView()

	case ScreenContentView:

		content = m.contentView()

	case ScreenInstaller:

		content = m.splitInstallerView()

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
	listWidth := int(float64(m.Width) * 0.4)
	previewWidth := m.Width - listWidth

	var searchBar string
	if m.searchFocused {
		searchBar = searchBarFocused.Width(listWidth - 2).Render(m.searchInput.View())
	} else {
		searchBar = searchBarBlurred.Width(listWidth - 2).Render(m.searchInput.View())
	}

	leftViewItems := []string{searchBar, m.list.View()}

	leftView := lipgloss.JoinVertical(lipgloss.Left, leftViewItems...)

	left := listStyle.Width(listWidth).Render(leftView)
	right := viewportStyle.Width(previewWidth).Render(m.viewport.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
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

	return header + "\n\n" + m.viewport.View()

}

func (m Model) splitInstallerView() string {
	listWidth := int(float64(m.Width) * 0.5)
	previewWidth := m.Width - listWidth

	// Task 4.1: Height fallback
	if m.Height < 24 {
		cardStyle = cardStyle.Border(lipgloss.NormalBorder()).MarginBottom(0)
	} else {
		cardStyle = cardStyle.Border(lipgloss.RoundedBorder()).MarginBottom(1)
	}

	left := lipgloss.NewStyle().Width(listWidth).PaddingRight(2).Render(m.installerOptionsView())
	right := viewportStyle.Width(previewWidth).PaddingLeft(2).Render(m.installerPreviewView())

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m Model) renderCard(title, content string, width int) string {
	styledTitle := cardTitleStyle.Render(title)
	return cardStyle.Width(width).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			styledTitle,
			content,
		),
	)
}

func (m Model) installerOptionsView() string {
	width := int(float64(m.Width)*0.5) - 4

	// Header
	banner := bannerStyle.Render("SYNCK INSTALLER")
	pathInfo := fmt.Sprintf("Target: %s", m.rootPath)
	header := lipgloss.JoinVertical(lipgloss.Left, banner, pathInfo)

	recommendedCheck := checkmarkStyle.Render("[x]")
	advancedCheck := "[ ]"
	if m.installerMode {
		recommendedCheck = "[ ]"
		advancedCheck = checkmarkStyle.Render("[x]")
	}
	modeStr := fmt.Sprintf("Mode (M to toggle):\n  %s Recommended\n      Use one shared install everywhere\n  %s Advanced\n      Create an isolated copy here", recommendedCheck, advancedCheck)
	header += "\n" + modeStr + "\n\n"

	// Providers
	var providersView string
	providers := []string{"Claude Code", "Gemini CLI", "Codex (OpenAI)", "GitHub Copilot", "OpenCode (OPENCODE.MD)"}
	for i, p := range providers {
		cursor := "  "
		if m.installerCursor == i {
			cursor = "> "
		}
		check := "[ ]"
		if m.installerProviders[i] {
			check = checkmarkStyle.Render("[x]")
		}
		providersView += fmt.Sprintf("%s%s %s\n", cursor, check, p)

		// Task 3.2: Hint for OpenCode
		if i == 4 && m.installerCursor == 4 {
			providersView += hintStyle.Render("    Effect: Synced AGENTS.md -> OPENCODE.MD") + "\n"
		}
	}
	providersCard := m.renderCard("[1] SELECT AI PROVIDERS", providersView, width)

	// Skills
	var skillsView string
	skills := []string{"skill-creator", "skill-sync", "find-skills"}
	for i, sk := range skills {
		cursor := "  "
		if m.installerCursor == i+5 {
			cursor = "> "
		}
		check := "[ ]"
		if m.installerSkills[i] {
			check = checkmarkStyle.Render("[x]")
		}
		skillsView += fmt.Sprintf("%s%s %s\n", cursor, check, sk)
	}
	skillsCard := m.renderCard("[2] INSTALL CORE SKILLS", skillsView, width)

	// Storage
	var storageView string
	for i, sk := range m.storedSkills {
		cursor := "  "
		if m.installerCursor == i+8 {
			cursor = "> "
		}
		check := "[ ]"
		if i < len(m.installerStoredSkills) && m.installerStoredSkills[i] {
			check = checkmarkStyle.Render("[x]")
		}
		storageView += fmt.Sprintf("%s%s %s\n", cursor, check, sk.Metadata.SkillName)
	}
	storageCard := m.renderCard("[3] INSTALL FROM STORAGE", storageView, width)

	// Global & Action
	var footer string
	storageOffset := len(m.storedSkills)

	cursorGlobal := "  "
	if m.installerCursor == 8+storageOffset {
		cursorGlobal = "> "
	}
	checkGlobal := "[ ]"
	if m.installerGlobal {
		checkGlobal = checkmarkStyle.Render("[x]")
	}
	footer += fmt.Sprintf("%s%s Add shell aliases to profile\n", cursorGlobal, checkGlobal)

	cursorAction := "  "
	if m.installerCursor == 9+storageOffset {
		cursorAction = "> "
	}
	footer += fmt.Sprintf("\n%s[ Execute Install ]\n", cursorAction)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		providersCard,
		skillsCard,
		storageCard,
		footer,
	)
}

func (m Model) installerPreviewView() string {
	width := m.Width - int(float64(m.Width)*0.5) - 4

	var content string

	// Directories
	content += lipgloss.NewStyle().Bold(true).Render("DIRECTORIES TO CREATE:") + "\n"
	if m.installerProviders[0] {
		content += "  + .claude/skills/\n"
	}
	if m.installerProviders[1] {
		content += "  + .gemini/skills/\n"
	}
	if m.installerProviders[2] {
		content += "  + .codex/skills/\n"
	}
	if m.installerProviders[3] {
		content += "  + .github/\n"
	}
	if m.installerProviders[4] {
		content += "  + .opencode/skills/\n"
	}
	content += "\n"

	// Skills
	content += lipgloss.NewStyle().Bold(true).Render("SKILLS TO COPY/LINK:") + "\n"
	if m.installerSkills[0] {
		content += "  + skill-creator -> .agent/skills/skill-creator\n"
	}
	if m.installerSkills[1] {
		content += "  + skill-sync    -> .agent/skills/skill-sync\n"
	}
	if m.installerSkills[2] {
		content += "  + find-skills   -> .agent/skills/find-skills\n"
	}
	content += "\n"

	// Configs
	content += lipgloss.NewStyle().Bold(true).Render("CONFIG FILES TO SYNC:") + "\n"
	if m.installerProviders[0] {
		content += "  + AGENTS.md -> CLAUDE.md\n"
	}
	if m.installerProviders[1] {
		content += "  + AGENTS.md -> GEMINI.md\n"
	}
	if m.installerProviders[3] {
		content += "  + AGENTS.md -> .github/copilot-instructions.md\n"
	}
	if m.installerProviders[4] {
		content += "  + AGENTS.md -> OPENCODE.MD\n"
	}
	content += "\n"

	if m.installerGlobal {
		content += lipgloss.NewStyle().Bold(true).Render("GLOBAL ALIASES:") + "\n"
		content += "  + Injecting 4 aliases into shell profile\n"
	}

	return m.renderCard("📋 INSTALLATION PREVIEW", content, width)
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
