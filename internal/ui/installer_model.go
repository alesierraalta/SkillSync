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
	Autoskills   bool
	Global       bool
	StoredSkills []bool
	Screen       Screen

	// External state needed for rendering/logic
	Width     int
	Height    int
	RootPath  string
	AllStored []storage.StoredSkill
	Backend   AppService
}

func NewInstallerModel(backend AppService, rootPath string) InstallerModel {
	return InstallerModel{
		Providers: []bool{true, false, true, false, true},
		Skills:    []bool{true, true, true},
		Global:    true,
		RootPath:  rootPath,
		Backend:   backend,
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
			if m.Screen == ScreenLicenseDisclosure {
				return m, nil
			}
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Screen == ScreenLicenseDisclosure {
				return m, nil
			}
			if m.Cursor < 10+len(m.AllStored) {
				m.Cursor++
			}
		case "m", "M":
			m.Mode = !m.Mode
		case " ", "space", "enter":
			if m.Screen == ScreenLicenseDisclosure {
				return m, nil
			}

			storageOffset := len(m.AllStored)
			if m.Cursor >= 0 && m.Cursor < 5 {
				m.Providers[m.Cursor] = !m.Providers[m.Cursor]
			} else if m.Cursor >= 5 && m.Cursor < 8 {
				m.Skills[m.Cursor-5] = !m.Skills[m.Cursor-5]
			} else if m.Cursor == 8 {
				m.Autoskills = !m.Autoskills
			} else if m.Cursor >= 9 && m.Cursor < 9+storageOffset {
				idx := m.Cursor - 9
				if m.StoredSkills == nil {
					m.StoredSkills = make([]bool, len(m.AllStored))
				}
				if len(m.StoredSkills) <= idx {
					newStored := make([]bool, idx+1)
					copy(newStored, m.StoredSkills)
					m.StoredSkills = newStored
				}
				m.StoredSkills[idx] = !m.StoredSkills[idx]
			} else if m.Cursor == 9+storageOffset {
				m.Global = !m.Global
			} else if m.Cursor == 10+storageOffset {
				if m.Autoskills {
					m.Screen = ScreenLicenseDisclosure
					return m, nil
				}
				return m, func() tea.Msg {
					return syncRequestMsg{
						Installer: m,
					}
				}
			}
		case "y", "Y", "n", "N", "esc":
			if m.Screen == ScreenLicenseDisclosure {
				if msg.String() == "y" || msg.String() == "Y" {
					m.Screen = ScreenInstaller
					return m, func() tea.Msg {
						return syncRequestMsg{
							Installer: m,
						}
					}
				}
				m.Screen = ScreenInstaller
				return m, nil
			}
		}
	}
	return m, nil
}

type syncRequestMsg struct {
	Installer InstallerModel
}

func (m InstallerModel) getCardStyle() lipgloss.Style {
	if m.Height < 24 {
		return cardStyle.Border(lipgloss.NormalBorder()).MarginBottom(0)
	}
	return cardStyle.Border(lipgloss.RoundedBorder()).MarginBottom(1)
}

func (m InstallerModel) View() string {
	if m.Screen == ScreenLicenseDisclosure {
		return m.LicenseDisclosureView()
	}

	listWidth := int(float64(m.Width) * 0.5)
	previewWidth := m.Width - listWidth

	left := lipgloss.NewStyle().Width(listWidth).PaddingRight(2).Render(m.OptionsView())
	right := viewportStyle.Width(previewWidth).PaddingLeft(2).Render(m.PreviewView())

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m InstallerModel) LicenseDisclosureView() string {
	width := m.Width - 8
	title := titleStyle.Render("AUTOSKILLS LICENSE DISCLOSURE")

	content := "\nThis action will execute 'midudev/autoskills' via npx.\n\n" +
		warningNoticeStyle.Render("LICENSE: Creative Commons Attribution-NonCommercial 4.0 International (CC-BY-NC-4.0)") + "\n\n" +
		"By proceeding, you acknowledge that this tool is for NON-COMMERCIAL use only.\n" +
		"The discovery process will analyze your project structure to suggest relevant skills.\n\n" +
		"Do you accept and want to proceed? (y/n)"

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center,
		m.getCardStyle().Width(width).Render(
			lipgloss.JoinVertical(lipgloss.Center, title, content),
		),
	)
}

func (m InstallerModel) renderCard(title, content string, width int) string {
	styledTitle := cardTitleStyle.Render(title)
	return m.getCardStyle().Width(width).Render(
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

	// Smart Scan
	cursorAuto := "  "
	if m.Cursor == 8 {
		cursorAuto = "> "
	}
	checkAuto := "[ ]"
	if m.Autoskills {
		checkAuto = checkmarkStyle.Render("[x]")
	}
	autoView := fmt.Sprintf("%s%s Smart Scan (autoskills)\n", cursorAuto, checkAuto)
	if m.Cursor == 8 {
		autoView += hintStyle.Render("    Detect technologies & suggest skills") + "\n"
	}
	autoCard := m.renderCard("[3] DISCOVERY", autoView, width)

	// Storage
	var storageView string
	for i, sk := range m.AllStored {
		cursor := "  "
		if m.Cursor == i+9 {
			cursor = "> "
		}
		check := "[ ]"
		if i < len(m.StoredSkills) && m.StoredSkills[i] {
			check = checkmarkStyle.Render("[x]")
		}
		storageView += fmt.Sprintf("%s%s %s\n", cursor, check, sk.Metadata.SkillName)
	}
	storageCard := m.renderCard("[4] INSTALL FROM STORAGE", storageView, width)

	// Global & Action
	var footer string
	storageOffset := len(m.AllStored)

	cursorGlobal := "  "
	if m.Cursor == 9+storageOffset {
		cursorGlobal = "> "
	}
	checkGlobal := "[ ]"
	if m.Global {
		checkGlobal = checkmarkStyle.Render("[x]")
	}
	footer += fmt.Sprintf("%s%s Add shell aliases to profile\n", cursorGlobal, checkGlobal)

	cursorAction := "  "
	if m.Cursor == 10+storageOffset {
		cursorAction = "> "
	}
	footer += fmt.Sprintf("\n%s[ Execute Install ]\n", cursorAction)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		providersCard,
		skillsCard,
		autoCard,
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
		content += "  + skill-creator -> .agents/skills/skill-creator\n"
	}
	if m.Skills[1] {
		content += "  + skill-sync    -> .agents/skills/skill-sync\n"
	}
	if m.Skills[2] {
		content += "  + find-skills   -> .agents/skills/find-skills\n"
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

	if m.Autoskills {
		content += "\n" + lipgloss.NewStyle().Bold(true).Render("SMART SCAN:") + "\n"
		content += "  + Executing autoskills discovery\n"
	}

	return m.renderCard("📋 INSTALLATION PREVIEW", content, width)
}
