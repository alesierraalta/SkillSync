package ui

import (
	"context"
	"fmt"
	"skillsync/tui/internal/discovery"
	"skillsync/tui/internal/parser"
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/types"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type skillsLoadedMsg []types.Skill
type errorMsg error

func (m Model) loadSkills() tea.Cmd {
	return func() tea.Msg {
		paths, err := discovery.DiscoverSkills(m.rootPath)
		if err != nil {
			return errorMsg(err)
		}

		var skills []types.Skill
		for _, p := range paths {
			s, err := parser.Parse(p)
			if err == nil {
				skills = append(skills, *s)
			}
		}
		return skillsLoadedMsg(skills)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case skillsLoadedMsg:
		var items []list.Item
		seen := make(map[string]bool)
		for _, s := range msg {
			if !seen[s.Name] {
				items = append(items, item{skill: s})
				seen[s.Name] = true
			}
		}
		cmd = m.list.SetItems(items)
		return m, cmd

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		listWidth := int(float64(m.Width) * 0.4)
		previewWidth := m.Width - listWidth

		m.list.SetSize(listWidth, m.Height)
		m.viewport.Width = previewWidth
		m.viewport.Height = m.Height

		if m.Screen == ScreenContentView {
			m.viewport.Width = m.Width
			m.viewport.Height = m.Height - 6
		}

		// Force re-render of preview to reflow markdown
		m.updatePreview()

		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case runner.SyncResult:
		m.syncOutput = fmt.Sprintf("Exit: %d\nOutput: %s\nErr: %s", msg.ExitCode, msg.Stdout, msg.Stderr)
		// Don't auto-switch back, let user see result
		return m, nil

	case errorMsg:
		m.err = msg
		return m, nil
	}

	return m.handleComponentUpdate(msg)
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.Screen {
	case ScreenList:
		return m.handleListKeys(msg)
	case ScreenDetail:
		return m.handleDetailKeys(msg)
	case ScreenSyncing:
		return m.handleSyncingKeys(msg)
	case ScreenContentView:
		return m.handleContentViewKeys(msg)
	}

	return m, nil
}

func (m Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if i, ok := m.list.SelectedItem().(item); ok {
			m.selected = &i.skill
			m.PrevScreen = m.Screen
			m.Screen = ScreenDetail
			m.setupInputs()
			return m, nil
		}
	case "s":
		m.PrevScreen = m.Screen
		m.Screen = ScreenSyncing
		return m, m.startSync()
	case "v":
		if i, ok := m.list.SelectedItem().(item); ok {
			m.selected = &i.skill
			m.PrevScreen = m.Screen
			m.Screen = ScreenContentView
			m.viewport.Width = m.Width
			m.viewport.Height = m.Height - 6
			m.updatePreview()
			return m, nil
		}
	case "pgup", "pgdown":
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	if i, ok := m.list.SelectedItem().(item); ok {
		if i.skill.Name != m.lastSelectedID {
			m.updatePreview()
		}
	}

	return m, cmd
}

func (m Model) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.Screen = m.PrevScreen
		return m, nil
	case "ctrl+s":
		return m, m.saveSkill()
	}

	var cmd tea.Cmd
	for i := range m.inputs {
		m.inputs[i], cmd = m.inputs[i].Update(msg)
	}
	return m, cmd
}

func (m Model) handleSyncingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.Screen = m.PrevScreen
		return m, nil
	}
	return m, nil
}

func (m Model) handleContentViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.Screen = m.PrevScreen
		return m, nil
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) handleComponentUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.Screen {
	case ScreenList:
		m.list, cmd = m.list.Update(msg)
	case ScreenDetail:
		for i := range m.inputs {
			m.inputs[i], cmd = m.inputs[i].Update(msg)
		}
	case ScreenContentView:
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

func (m *Model) setupInputs() {
	m.inputs = make([]textinput.Model, 2)

	d := textinput.New()
	d.Placeholder = "Description"
	if m.selected != nil {
		d.SetValue(m.selected.Metadata.Description)
	}
	d.Focus()
	m.inputs[0] = d

	s := textinput.New()
	s.Placeholder = "Scope"
	if m.selected != nil {
		s.SetValue(m.selected.Metadata.Scope)
	}
	m.inputs[1] = s
}

func (m Model) startSync() tea.Cmd {
	return func() tea.Msg {
		r := runner.NewRunner("./.agents/skills/skill-sync/assets/sync.sh")
		resChan := r.ExecuteSync(context.Background(), nil)
		return <-resChan
	}
}

func (m Model) saveSkill() tea.Cmd {
	if m.selected == nil {
		return nil
	}
	m.selected.Metadata.Description = m.inputs[0].Value()
	m.selected.Metadata.Scope = m.inputs[1].Value()

	return func() tea.Msg {
		err := parser.Save(m.selected.Path, m.selected)
		if err != nil {
			return errorMsg(err)
		}
		return m.loadSkills()()
	}
}
