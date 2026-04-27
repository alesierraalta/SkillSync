package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"skillsync/tui/internal/types"
)

type Screen int

const (
	ScreenHome Screen = iota
	ScreenList
	ScreenDetail
	ScreenSyncing
	ScreenContentView
	ScreenInstaller
)

type Model struct {
	Screen         Screen
	PrevScreen     Screen
	Width          int
	Height         int
	list           list.Model
	viewport       viewport.Model
	lastSelectedID string
	selected       *types.Skill
	inputs         []textarea.Model
	syncOutput     string
	err            error
	rootPath       string
	renderer       *glamour.TermRenderer
	HomeCursor     int
	StatusMsg      string
	Progress       progress.Model

	// Installer State
	installerCursor    int
	installerMode      bool // false = Symlink, true = Copy
	installerProviders []bool
	installerSkills    []bool
	installerGlobal    bool
}

type item struct {
	skill types.Skill
}

func (i item) Title() string {
	if i.skill.ID == "virtual:agents" {
		return i.skill.Name
	}
	
	path := filepath.ToSlash(i.skill.Path)
	segments := strings.Split(path, "/")
	
	flag := ""
	for _, segment := range segments {
		switch segment {
		case ".opencode", ".claude", ".gemini", ".cursor", ".copilot", ".agents":
			flag = "[" + segment + "]"
			break
		}
		if flag != "" {
			break
		}
	}
	
	if flag == "" {
		return i.skill.Name
	}
	return fmt.Sprintf("%s %s", i.skill.Name, flag)
}
func (i item) Description() string { return i.skill.Metadata.Description }
func (i item) FilterValue() string { return i.skill.Name }

func NewModel() Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Skillsync TUI"

	return Model{
		Screen:             ScreenHome,
		PrevScreen:         ScreenHome,
		list:               l,
		viewport:           viewport.New(0, 0),
		rootPath:           ".",
		Progress:           progress.New(progress.WithDefaultGradient()),
		installerProviders: []bool{true, false, true, false, true},
		installerSkills:    []bool{true, true, false},
		installerGlobal:    true,
	}
}

func (m Model) GetKeyBindings() []KeyBinding {
	switch m.Screen {
	case ScreenHome:
		return []KeyBinding{
			{Key: "q/esc", Help: "quit"},
			{Key: "up/down", Help: "navigate"},
			{Key: "enter", Help: "select"},
		}
	case ScreenList:
		return []KeyBinding{
			{Key: "q", Help: "quit"},
			{Key: "enter", Help: "preview"},
			{Key: "e", Help: "edit skill"},
			{Key: "S", Help: "sync"},
		}
	case ScreenDetail:
		return []KeyBinding{
			{Key: "esc", Help: "back"},
			{Key: "tab", Help: "switch field"},
			{Key: "ctrl+s", Help: "save (desc/scope/content)"},
		}
	case ScreenContentView:
		return []KeyBinding{
			{Key: "esc", Help: "back"},
			{Key: "e", Help: "edit content"},
			{Key: "j/k", Help: "scroll"},
		}
	case ScreenInstaller:
		return []KeyBinding{
			{Key: "q/esc", Help: "back"},
			{Key: "up/down", Help: "navigate"},
			{Key: "space", Help: "toggle"},
			{Key: "enter", Help: "execute/select"},
		}
	default:
		return []KeyBinding{
			{Key: "esc", Help: "back"},
		}
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
