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

	title := titleStyle.Render("Installing Ecosistema...")

	

	currentTask := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render(m.syncOutput)

	

	s := title + "\n\n"

	s += currentTask + "\n\n"

	s += m.Progress.View() + "\n\n"

	s += footerStyle.Render("esc: back")

	

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



	left := lipgloss.NewStyle().Width(listWidth).PaddingRight(2).Render(m.installerOptionsView())

	right := viewportStyle.Width(previewWidth).PaddingLeft(2).Render(m.installerPreviewView())



	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)

}



func (m Model) installerOptionsView() string {

	title := titleStyle.Render("Target Project: " + m.rootPath)

	

	modeStr := "[ Symlink (Global) ] / Copy (Local)"

	if m.installerMode {

		modeStr = "Symlink (Global) / [ Copy (Local) ]"

	}

	

	s := title + "\nMode (M to toggle): " + modeStr + "\n\n"

	

	// Providers

	s += lipgloss.NewStyle().Bold(true).Render("[1] SELECT AI PROVIDERS") + "\n"

	providers := []string{"Claude Code", "Gemini CLI", "Codex (OpenAI)", "GitHub Copilot", "OpenCode"}

	for i, p := range providers {

		cursor := "  "

		if m.installerCursor == i {

			cursor = "> "

		}

		check := "[ ]"

		if m.installerProviders[i] {

			check = "[x]"

		}

		s += fmt.Sprintf("%s%s %s\n", cursor, check, p)

	}

	s += "\n"

	

	// Skills

	s += lipgloss.NewStyle().Bold(true).Render("[2] INSTALL CORE SKILLS") + "\n"

	skills := []string{"skill-creator", "skill-sync", "find-skills"}

	for i, sk := range skills {

		cursor := "  "

		if m.installerCursor == i+5 {

			cursor = "> "

		}

		check := "[ ]"

		if m.installerSkills[i] {

			check = "[x]"

		}

		s += fmt.Sprintf("%s%s %s\n", cursor, check, sk)

	}

	s += "\n"

	

	// Storage
	s += lipgloss.NewStyle().Bold(true).Render("[3] INSTALL FROM STORAGE") + "\n"
	for i, sk := range m.storedSkills {
		cursor := "  "
		if m.installerCursor == i+8 {
			cursor = "> "
		}
		check := "[ ]"
		if i < len(m.installerStoredSkills) && m.installerStoredSkills[i] {
			check = "[x]"
		}
		s += fmt.Sprintf("%s%s %s\n", cursor, check, sk.Metadata.SkillName)
	}
	s += "\n"

	// Global
	s += lipgloss.NewStyle().Bold(true).Render("[4] GLOBAL INSTALL") + "\n"
	cursorGlobal := "  "
	storageOffset := len(m.storedSkills)
	if m.installerCursor == 8+storageOffset {
		cursorGlobal = "> "
	}
	checkGlobal := "[ ]"
	if m.installerGlobal {
		checkGlobal = "[x]"
	}
	s += fmt.Sprintf("%s%s Add shell aliases to profile\n\n", cursorGlobal, checkGlobal)
	
	// Action
	cursorAction := "  "
	if m.installerCursor == 9+storageOffset {
		cursorAction = "> "
	}
	s += fmt.Sprintf("%s[ Execute Install ]\n", cursorAction)

	

	return s

}



func (m Model) installerPreviewView() string {

	s := titleStyle.Render("đź“‹ INSTALLATION PREVIEW") + "\n"

	s += "--------------------------------------------------\n"

	

	s += lipgloss.NewStyle().Bold(true).Render("DIRECTORIES TO CREATE:") + "\n"

	if m.installerProviders[0] { s += "  + .claude/skills/\n" }

	if m.installerProviders[1] { s += "  + .gemini/skills/\n" }

	if m.installerProviders[2] { s += "  + .codex/skills/\n" }

	if m.installerProviders[3] { s += "  + .github/\n" }

	if m.installerProviders[4] { s += "  + .opencode/skills/\n" }

	s += "\n"

	

	s += lipgloss.NewStyle().Bold(true).Render("SKILLS TO COPY/LINK:") + "\n"

	if m.installerSkills[0] { s += "  + skill-creator -> .agent/skills/skill-creator\n" }

	if m.installerSkills[1] { s += "  + skill-sync    -> .agent/skills/skill-sync\n" }

	if m.installerSkills[2] { s += "  + find-skills   -> .agent/skills/find-skills\n" }

	s += "\n"

	

	s += lipgloss.NewStyle().Bold(true).Render("CONFIG FILES TO SYNC:") + "\n"

	if m.installerProviders[0] { s += "  + AGENTS.md -> CLAUDE.md\n" }

	if m.installerProviders[1] { s += "  + AGENTS.md -> GEMINI.md\n" }

	if m.installerProviders[3] { s += "  + AGENTS.md -> .github/copilot-instructions.md\n" }

	if m.installerProviders[4] { s += "  + AGENTS.md -> OPENCODE.md\n" }

	s += "\n"

	

	if m.installerGlobal {

		s += lipgloss.NewStyle().Bold(true).Render("GLOBAL ALIASES:") + "\n"

		s += "  + Injecting 4 aliases into shell profile\n"

	}

	

	return s

}


func (m Model) storageView() string {
	s := titleStyle.Render("Almacenamiento Global de Skills") + "\n\n"
	if len(m.storedSkills) == 0 {
		return s + lipgloss.NewStyle().MarginLeft(4).Render("No hay skills almacenadas.\nPresioná 's' en la vista de gestión para guardar una.")
	}
	return s + docStyle.Render(m.storageList.View())
}

