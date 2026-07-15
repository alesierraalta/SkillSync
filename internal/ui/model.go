package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/muesli/reflow/wordwrap"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"skillsync/tui/internal/agentdetect"
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/types"
)

type Screen int

type syncFinishedMsg struct {
	err error
}

type syncReportMsg struct {
	report *runner.SyncReport
	err    error
}

const (
	ScreenHome Screen = iota
	ScreenList
	ScreenDetail
	ScreenSyncing
	ScreenContentView
	ScreenInstaller
	ScreenStorage
	ScreenProjects
	ScreenDeleteConfirm
	ScreenLicenseDisclosure
	ScreenAgentEcosystem
	ScreenAgentMenu
	ScreenPluginsMenu
	ScreenMCPServersMenu
	ScreenGlobalSkillsCats
	ScreenGlobalSkillsList
	ScreenBundleImport
)

type Model struct {
	Screen       Screen
	PrevScreen   Screen
	Width        int
	Height       int
	HomeCursor   int
	StatusMsg    string
	Progress     progress.Model
	SyncFailed   bool
	SyncFinished bool
	syncReport   *runner.SyncReport
	err          error
	rootPath     string
	inputs       []textarea.Model
	syncOutput   string
	selected     *types.Skill

	// Sub-models
	Installer InstallerModel
	List      ListModel

	// Storage State
	storageList  list.Model
	storedSkills []storage.StoredSkill

	// Projects State
	projectList list.Model

	// Delete Confirm State
	deleteConfirm DeleteConfirmModel

	// Service Layer
	backend AppService

	// Agent Ecosystem State
	agentEcosystem       []agentdetect.AgentInfo
	selectedAgent        int
	agentEcosystemScroll int
	agentMenuCursor      int
	pluginsMenuCursor    int
	mcpServersMenuCursor int

	// Global Skills State
	globalSkillsList     list.Model
	globalCategory       string
	globalCategoryCursor int
	globalSkillsLoaded   bool
	globalSkillsErr      error

	// Vault selection / bundle state
	selectMode     bool            // multi-select active in Global Skills
	vaultSelected  map[string]bool // selected vault skill names
	bundleImportIn textinput.Model // path input on the import screen
}

type globalSkillItem struct {
	skill    types.Skill
	category string
}

func (i globalSkillItem) Title() string {
	if i.category == "All" {
		path := filepath.ToSlash(i.skill.Path)
		segments := strings.Split(path, "/")

		flag := ""
		for _, segment := range segments {
			switch segment {
			case ".opencode", ".claude", ".gemini", ".cursor", ".copilot", ".agents":
				flag = "[" + strings.TrimPrefix(segment, ".") + "]"
				break
			}
			if flag != "" {
				break
			}
		}

		if flag != "" {
			return fmt.Sprintf("%s %s", flag, i.skill.Name)
		}
	}
	return i.skill.Name
}
func (i globalSkillItem) Description() string {
	desc := i.skill.Metadata.Description
	if desc == "" {
		desc = "Sin descripción"
	}
	wrapped := wordwrap.String(desc, 70)
	return fmt.Sprintf("Path: %s\n%s", i.skill.Path, wrapped)
}
func (i globalSkillItem) FilterValue() string {
	return i.skill.Name
}

type item struct {
	skill types.Skill
}

type storageItem struct {
	stored storage.StoredSkill
}

func (i storageItem) Title() string {
	return fmt.Sprintf("%s [%s]", i.stored.Metadata.SkillName, filepath.Base(i.stored.Metadata.OriginProject))
}
func (i storageItem) Description() string {
	return fmt.Sprintf("Stored at: %s | Project: %s", i.stored.Metadata.SavedAt.Format("2006-01-02 15:04"), i.stored.Metadata.OriginProject)
}
func (i storageItem) FilterValue() string { return i.stored.Metadata.SkillName }

type projectItem struct {
	project storage.ProjectInfo
}

func (i projectItem) Title() string { return i.project.Path }
func (i projectItem) Description() string {
	if i.project.LastSynced.IsZero() {
		return "Último sync: Nunca"
	}
	return fmt.Sprintf("Último sync: %s", i.project.LastSynced.Format("2006-01-02 15:04"))
}
func (i projectItem) FilterValue() string { return i.project.Path }

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
func (i item) Description() string {
	desc := i.skill.Metadata.Description
	if desc == "" {
		desc = "Sin descripción"
	}
	return wordwrap.String(desc, 70)
}
func (i item) FilterValue() string { return i.skill.Name }

func NewModel(backend AppService) Model {
	sl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	sl.Title = "Almacenamiento Global"

	pl := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	pl.Title = "Proyectos Sincronizados"

	glDelegate := list.NewDefaultDelegate()
	glDelegate.SetHeight(5)
	gl := list.New([]list.Item{}, glDelegate, 0, 0)
	gl.Title = "Global Skills"

	importIn := textinput.New()
	importIn.Placeholder = "path/to/bundle.skillsync"
	importIn.Prompt = "> "

	return Model{
		Screen:           ScreenHome,
		PrevScreen:       ScreenHome,
		storageList:      sl,
		projectList:      pl,
		globalSkillsList: gl,
		rootPath:         ".",
		Progress:         progress.New(progress.WithDefaultGradient()),
		Installer:        NewInstallerModel(backend, "."),
		List:             NewListModel(backend, "."),
		deleteConfirm:    NewDeleteConfirmModel(backend),
		backend:          backend,
		vaultSelected:    make(map[string]bool),
		bundleImportIn:   importIn,
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
			{Key: "s", Help: "save globally"},
			{Key: "y", Help: "sync"},
			{Key: "d", Help: "delete skill"},
			{Key: "o", Help: "open folder"},
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
	case ScreenStorage:
		return []KeyBinding{
			{Key: "esc", Help: "back"},
			{Key: "up/down", Help: "navigate"},
			{Key: "i", Help: "install & sync"},
			{Key: "d", Help: "delete from storage"},
		}
	case ScreenProjects:
		return []KeyBinding{
			{Key: "esc/q", Help: "back"},
			{Key: "r", Help: "refresh projects"},
		}
	case ScreenDeleteConfirm:
		return []KeyBinding{
			{Key: "y", Help: "confirm delete"},
			{Key: "n/esc", Help: "cancel"},
		}
	case ScreenAgentEcosystem:
		return []KeyBinding{
			{Key: "esc/q", Help: "back"},
			{Key: "up/down", Help: "navigate"},
			{Key: "j/k", Help: "scroll"},
			{Key: "enter", Help: "select"},
			{Key: "o", Help: "open dir"},
		}
	case ScreenAgentMenu:
		return []KeyBinding{
			{Key: "esc/q", Help: "back"},
			{Key: "up/down", Help: "navigate"},
			{Key: "enter", Help: "select"},
			{Key: "o", Help: "open dir"},
		}
	case ScreenPluginsMenu:
		return []KeyBinding{
			{Key: "esc/q", Help: "back"},
			{Key: "up/down", Help: "navigate"},
			{Key: "o", Help: "open dir"},
		}
	case ScreenMCPServersMenu:
		return []KeyBinding{
			{Key: "esc/q", Help: "back"},
			{Key: "up/down", Help: "navigate"},
			{Key: "o", Help: "open dir"},
		}
	case ScreenGlobalSkillsCats:
		return []KeyBinding{
			{Key: "esc/q", Help: "back"},
			{Key: "up/down", Help: "navigate"},
			{Key: "enter", Help: "select"},
		}
	case ScreenGlobalSkillsList:
		return []KeyBinding{
			{Key: "esc/q", Help: "back"},
			{Key: "enter", Help: "preview"},
			{Key: "d", Help: "delete skill"},
			{Key: "o", Help: "open folder"},
		}
	default:
		return []KeyBinding{
			{Key: "esc", Help: "back"},
		}
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.List.Init(), m.loadSkills(), instantiateEcosystemCmd(m.backend, m.rootPath))
}

// isCoreSkill returns true if the given skill name is protected.
func isCoreSkill(name string) bool {
	switch name {
	case "skill-creator", "skill-sync", "find-skills":
		return true
	}
	return false
}

// resetDeleteConfirm clears the delete confirmation state.
func (m *Model) resetDeleteConfirm() {
	m.deleteConfirm = NewDeleteConfirmModel(m.backend)
}
