package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"skillsync/tui/internal/types"
)

type Screen int

const (
	ScreenList Screen = iota
	ScreenDetail
	ScreenSyncing
	ScreenContentView
)

type Model struct {
	Screen     Screen
	PrevScreen Screen
	Width      int
	Height     int
	list       list.Model
	viewport       viewport.Model
	lastSelectedID string
	selected       *types.Skill
	inputs     []textinput.Model
	syncOutput string
	err        error
	rootPath   string
	renderer   *glamour.TermRenderer
}

type item struct {
	skill types.Skill
}

func (i item) Title() string       { return i.skill.Name }
func (i item) Description() string { return i.skill.Metadata.Description }
func (i item) FilterValue() string { return i.skill.Name }

func NewModel() Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Skillsync TUI"

	return Model{
		Screen:     ScreenList,
		PrevScreen: ScreenList,
		list:       l,
		viewport:   viewport.New(0, 0),
		rootPath:   ".",
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadSkills()
}

func (m *Model) initRenderer() error {
	width := m.Width - 4
	if m.Screen == ScreenList {
		width = int(float64(m.Width)*0.6) - 4
	}
	return m.initRendererWithWidth(width)
}

func (m Model) renderMarkdown(content string) (string, error) {
	if m.renderer == nil {
		return content, nil
	}
	return m.renderer.Render(content)
}

func (m *Model) updatePreview() {
	if i, ok := m.list.SelectedItem().(item); ok {
		m.selected = &i.skill
		m.lastSelectedID = i.skill.Name
		content := m.selected.Prefix + m.selected.RawBody
		if content == "" {
			content = "No content"
		}

		// Calculate preview width (60% of total)
		previewWidth := int(float64(m.Width) * 0.6)
		if previewWidth <= 0 {
			previewWidth = m.Width // Fallback for full screen or uninitialized
		}

		// Re-init renderer with correct width
		_ = m.initRendererWithWidth(previewWidth - 4)

		rendered, err := m.renderMarkdown(content)
		if err == nil {
			content = rendered
		}
		m.viewport.SetContent(content)
	}
}

func (m *Model) initRendererWithWidth(width int) error {
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
