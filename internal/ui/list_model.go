package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"skillsync/tui/internal/types"
)

type ListModel struct {
	list           list.Model
	viewport       viewport.Model
	lastSelectedID string
	selected       *types.Skill
	allSkills      []list.Item
	searchInput    textinput.Model
	searchFocused  bool
	Width          int
	Height         int
	renderer       *glamour.TermRenderer
	backend        AppService
	rootPath       string
}

func NewListModel(backend AppService, rootPath string) ListModel {
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(4)
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Skillsync TUI"
	l.KeyMap.Filter.SetEnabled(false)

	ti := textinput.New()
	ti.Placeholder = "Search skills..."
	ti.Blur()

	return ListModel{
		list:        l,
		viewport:    viewport.New(0, 0),
		searchInput: ti,
		backend:     backend,
		rootPath:    rootPath,
	}
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case skillsLoadedMsg:
		var items []list.Item
		seen := make(map[string]bool)
		for _, s := range msg {
			if !seen[s.Path] {
				items = append(items, item{skill: s})
				seen[s.Path] = true
			}
		}
		m.allSkills = items
		cmd = m.list.SetItems(items)
		return m, cmd

	case tea.WindowSizeMsg:
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if m.searchFocused {
			switch msg.String() {
			case "esc", "tab":
				m.searchFocused = false
				m.searchInput.Blur()
				return m, nil
			}

			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			m.filterSkills(m.searchInput.Value())
			return m, cmd
		}

		switch msg.String() {
		case "tab":
			m.searchFocused = true
			return m, m.searchInput.Focus()
		case "enter", "v":
			if i, ok := m.list.SelectedItem().(item); ok {
				m.selected = &i.skill
				// Navigation handled by root model capturing state change
				// but we update preview
				m.updatePreview()
			}
		case "pgup", "pgdown":
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

		m.list, cmd = m.list.Update(msg)
		if i, ok := m.list.SelectedItem().(item); ok {
			if i.skill.ID != m.lastSelectedID {
				m.updatePreview()
			}
		}
		return m, cmd
	}

	// Forward any unhandled message to the underlying list model
	// (status messages, clear-status ticks, etc.)
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

type editRequestMsg struct {
	Skill *types.Skill
}

func (m ListModel) View() string {
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

func (m *ListModel) filterSkills(query string) {
	if query == "" {
		m.list.SetItems(m.allSkills)
		return
	}
	q := strings.ToLower(query)
	var filtered []list.Item
	for _, it := range m.allSkills {
		if itm, ok := it.(item); ok {
			if strings.Contains(strings.ToLower(itm.FilterValue()), q) {
				filtered = append(filtered, it)
			}
		}
	}
	m.list.SetItems(filtered)
}

func (m *ListModel) updatePreview() {
	if i, ok := m.list.SelectedItem().(item); ok {
		m.selected = &i.skill
		m.lastSelectedID = i.skill.ID

		desc := m.selected.Metadata.Description
		if desc == "" {
			desc = "No description provided"
		}
		scope := m.selected.Metadata.Scope
		if scope == "" {
			scope = "No scope specified"
		}

		metadata := fmt.Sprintf("# %s\n\n**Description:** %s\n**Scope:** %s\n\n*e Edit content*\n\n---\n\n",
			m.selected.Name, desc, scope)

		rawContent := m.selected.Prefix + m.selected.RawBody
		if rawContent == "" {
			rawContent = "_No content available_"
		}

		content := metadata + rawContent

		previewWidth := int(float64(m.Width) * 0.6)
		if previewWidth <= 0 {
			previewWidth = m.Width
		}

		_ = m.initRendererWithWidth(previewWidth - 4)

		rendered, err := m.renderMarkdown(content)
		if err == nil {
			content = rendered
		}
		m.viewport.SetContent(content)
	}
}

func (m *ListModel) initRendererWithWidth(width int) error {
	if width <= 0 {
		width = 80
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return err
	}
	m.renderer = r
	return nil
}

func (m ListModel) renderMarkdown(content string) (string, error) {
	if m.renderer == nil {
		return content, nil
	}
	return m.renderer.Render(content)
}
