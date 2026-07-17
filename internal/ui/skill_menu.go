package ui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// skillMenuOptions are the actions offered when a skill is selected.
// Order must match the cursor indices handled in handleSkillMenuKeys.
var skillMenuOptions = []string{
	"Preview content",
	"Edit skill",
	"Browse files (references, assets, ...)",
	"Save to global storage",
}

// openSkillMenu switches to the skill action submenu, remembering the list
// screen that launched it so esc can return there.
func (m Model) openSkillMenu() (tea.Model, tea.Cmd) {
	if m.selected == nil {
		return m, nil
	}
	m.skillMenuOrigin = m.Screen
	m.skillMenuCursor = 0
	m.PrevScreen = m.Screen
	m.Screen = ScreenSkillMenu
	return m, nil
}

// loadSelectedSkillContent reads the selected skill's SKILL.md into the
// shared viewport so ScreenContentView shows fresh content.
func (m *Model) loadSelectedSkillContent() {
	contentBytes, err := os.ReadFile(m.selected.Path)
	if err != nil {
		m.List.viewport.SetContent(fmt.Sprintf("Error reading file: %v", err))
	} else {
		m.List.viewport.SetContent(string(contentBytes))
	}
	m.List.viewport.GotoTop()
}

func (m Model) handleSkillMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.Screen = m.skillMenuOrigin
		return m, nil
	case "j", "down":
		if m.skillMenuCursor < len(skillMenuOptions)-1 {
			m.skillMenuCursor++
		}
	case "k", "up":
		if m.skillMenuCursor > 0 {
			m.skillMenuCursor--
		}
	case "enter":
		if m.selected == nil {
			return m, nil
		}
		switch m.skillMenuCursor {
		case 0: // Preview content
			m.loadSelectedSkillContent()
			m.PrevScreen = ScreenSkillMenu
			m.Screen = ScreenContentView
			m.List.viewport.Width = m.Width
			m.List.viewport.Height = m.Height - 6
			return m, nil
		case 1: // Edit skill
			m.PrevScreen = ScreenSkillMenu
			m.Screen = ScreenDetail
			m.setupInputs()
			return m, nil
		case 2: // Browse files
			// Load SKILL.md first so the content view behind the browser's
			// esc chain (files → content → menu) is coherent.
			m.loadSelectedSkillContent()
			m.PrevScreen = ScreenSkillMenu
			return m.openSkillFileBrowser()
		case 3: // Save to global storage
			// Copies SKILL.md plus references/assets into the vault (see
			// saveToStorageCmd → storage.Save). Return to the launching
			// screen; the status message reports the result.
			cmd := m.saveToStorageCmd()
			m.Screen = m.skillMenuOrigin
			return m, cmd
		}
	}
	return m, nil
}

func (m Model) viewSkillMenu() string {
	if m.selected == nil {
		return "No skill selected"
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Skill: %s", m.selected.Name)) + "\n\n")
	b.WriteString(hintStyle.Render("enter: select · j/k: navigate · esc: back") + "\n\n")
	for i, opt := range skillMenuOptions {
		cursor := "  "
		if i == m.skillMenuCursor {
			cursor = "> "
		}
		b.WriteString(cursor + opt + "\n")
	}
	return docStyle.Render(b.String())
}
