package ui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"skillsync/tui/internal/storage"
)

type InstallerModel struct {
	Cursor       int
	Mode         bool // false = Recommended, true = Advanced
	Providers    []bool
	Skills       []bool
	Global       bool
	StoredSkills []bool
	
	// External state needed for rendering/logic
	Width        int
	Height       int
	RootPath     string
	AllStored    []storage.StoredSkill
	Backend      AppService
}

func NewInstallerModel(backend AppService, rootPath string) InstallerModel {
	return InstallerModel{
		Providers:    []bool{true, false, true, false, true},
		Skills:       []bool{true, true, true},
		Global:       true,
		RootPath:     rootPath,
		Backend:      backend,
	}
}

func (m InstallerModel) Init() tea.Cmd {
	return nil
}

func (m InstallerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < 9+len(m.AllStored) {
				m.Cursor++
			}
		case "m", "M":
			m.Mode = !m.Mode
		case " ", "space", "enter":
			storageOffset := len(m.AllStored)
			if m.Cursor >= 0 && m.Cursor < 5 {
				m.Providers[m.Cursor] = !m.Providers[m.Cursor]
			} else if m.Cursor >= 5 && m.Cursor < 8 {
				m.Skills[m.Cursor-5] = !m.Skills[m.Cursor-5]
			} else if m.Cursor >= 8 && m.Cursor < 8+storageOffset {
				idx := m.Cursor - 8
				if idx < len(m.StoredSkills) {
					m.StoredSkills[idx] = !m.StoredSkills[idx]
				}
			} else if m.Cursor == 8+storageOffset {
				m.Global = !m.Global
			} else if m.Cursor == 9+storageOffset && msg.String() == "enter" {
				return m, func() tea.Msg {
					return syncRequestMsg{
						Installer: m,
					}
				}
			}
		}
	}
	return m, nil
}

type syncRequestMsg struct {
	Installer InstallerModel
}

func (m InstallerModel) View() string {
	listWidth := int(float64(m.Width) * 0.5)
	previewWidth := m.Width - listWidth

	// Height fallback
	var localCardStyle = cardStyle
	if m.Height < 24 {
		localCardStyle = localCardStyle.Border(lipgloss.NormalBorder()).MarginBottom(0)
	} else {
		localCardStyle = localCardStyle.Border(lipgloss.RoundedBorder()).MarginBottom(1)
	}

	left := lipgloss.NewStyle().Width(listWidth).PaddingRight(2).Render(m.OptionsView())
	right := viewportStyle.Width(previewWidth).PaddingLeft(2).Render(m.PreviewView())

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m InstallerModel) renderCard(title, content string, width int) string {
	styledTitle := cardTitleStyle.Render(title)
	return cardStyle.Width(width).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			styledTitle,
			content,
		),
	)
}

func (m InstallerModel) OptionsView() string {
	width := int(float64(m.Width)*0.5) - 4

	// Header
	banner := bannerStyle.Render("SYNCK INSTALLER")
	pathInfo := fmt.Sprintf("Target: %s", m.RootPath)
	header := lipgloss.JoinVertical(lipgloss.Left, banner, pathInfo)

	recommendedCheck := checkmarkStyle.Render("[x]")
	advancedCheck := "[ ]"
	if m.Mode {
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
		if m.Cursor == i {
			cursor = "> "
		}
		check := "[ ]"
		if m.Providers[i] {
			check = checkmarkStyle.Render("[x]")
		}
		providersView += fmt.Sprintf("%s%s %s\n", cursor, check, p)

		if i == 4 && m.Cursor == 4 {
			providersView += hintStyle.Render("    Effect: Synced AGENTS.md -> OPENCODE.MD") + "\n"
		}
	}
	providersCard := m.renderCard("[1] SELECT AI PROVIDERS", providersView, width)

	// Skills
	var skillsView string
	skills := []string{"skill-creator", "skill-sync", "find-skills"}
	for i, sk := range skills {
		cursor := "  "
		if m.Cursor == i+5 {
			cursor = "> "
		}
		check := "[ ]"
		if m.Skills[i] {
			check = checkmarkStyle.Render("[x]")
		}
		skillsView += fmt.Sprintf("%s%s %s\n", cursor, check, sk)
	}
	skillsCard := m.renderCard("[2] INSTALL CORE SKILLS", skillsView, width)

	// Storage
	var storageView string
	for i, sk := range m.AllStored {
		cursor := "  "
		if m.Cursor == i+8 {
			cursor = "> "
		}
		check := "[ ]"
		if i < len(m.StoredSkills) && m.StoredSkills[i] {
			check = checkmarkStyle.Render("[x]")
		}
		storageView += fmt.Sprintf("%s%s %s\n", cursor, check, sk.Metadata.SkillName)
	}
	storageCard := m.renderCard("[3] INSTALL FROM STORAGE", storageView, width)

	// Global & Action
	var footer string
	storageOffset := len(m.AllStored)

	cursorGlobal := "  "
	if m.Cursor == 8+storageOffset {
		cursorGlobal = "> "
	}
	checkGlobal := "[ ]"
	if m.Global {
		checkGlobal = checkmarkStyle.Render("[x]")
	}
	footer += fmt.Sprintf("%s%s Add shell aliases to profile\n", cursorGlobal, checkGlobal)

	cursorAction := "  "
	if m.Cursor == 9+storageOffset {
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

func (m InstallerModel) PreviewView() string {
	width := m.Width - int(float64(m.Width)*0.5) - 4

	var content string

	// Directories
	content += lipgloss.NewStyle().Bold(true).Render("DIRECTORIES TO CREATE:") + "\n"
	if m.Providers[0] {
		content += "  + .claude/skills/\n"
	}
	if m.Providers[1] {
		content += "  + .gemini/skills/\n"
	}
	if m.Providers[2] {
		content += "  + .codex/skills/\n"
	}
	if m.Providers[3] {
		content += "  + .github/\n"
	}
	if m.Providers[4] {
		content += "  + .opencode/skills/\n"
	}
	content += "\n"

	// Skills
	content += lipgloss.NewStyle().Bold(true).Render("SKILLS TO COPY/LINK:") + "\n"
	if m.Skills[0] {
		content += "  + skill-creator -> .agent/skills/skill-creator\n"
	}
	if m.Skills[1] {
		content += "  + skill-sync    -> .agent/skills/skill-sync\n"
	}
	if m.Skills[2] {
		content += "  + find-skills   -> .agent/skills/find-skills\n"
	}
	content += "\n"

	// Configs
	content += lipgloss.NewStyle().Bold(true).Render("CONFIG FILES TO SYNC:") + "\n"
	if m.Providers[0] {
		content += "  + AGENTS.md -> CLAUDE.md\n"
	}
	if m.Providers[1] {
		content += "  + AGENTS.md -> GEMINI.md\n"
	}
	if m.Providers[3] {
		content += "  + AGENTS.md -> .github/copilot-instructions.md\n"
	}
	if m.Providers[4] {
		content += "  + AGENTS.md -> OPENCODE.MD\n"
	}
	content += "\n"

	if m.Global {
		content += lipgloss.NewStyle().Bold(true).Render("GLOBAL ALIASES:") + "\n"
		content += "  + Injecting 4 aliases into shell profile\n"
	}

	return m.renderCard("📋 INSTALLATION PREVIEW", content, width)
}
