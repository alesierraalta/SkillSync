package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/syncengine"
	"skillsync/tui/internal/types"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type skillsLoadedMsg []types.Skill
type errorMsg error

type installerProgressMsg struct {
	percent float64
	task    string
}

type installerFinishedMsg struct {
	err error
}

type storedSkillsLoadedMsg []storage.StoredSkill
type projectsLoadedMsg []storage.ProjectInfo
type statusMsg string

func (m Model) loadSkills() tea.Cmd {
	return func() tea.Msg {
		paths, err := m.backend.DiscoverSkills(m.rootPath)
		if err != nil {
			return errorMsg(err)
		}

		var skills []types.Skill
		for _, p := range paths {
			s, err := m.backend.ParseSkill(p)
			if err == nil {
				skills = append(skills, *s)
			}
		}

		// Inject virtual AGENTS.md skill if exists
		agentsPath := filepath.Join(m.rootPath, "AGENTS.md")
		if _, err := os.Stat(agentsPath); err == nil {
			content, err := os.ReadFile(agentsPath)
			if err == nil {
				virtualAgent := types.Skill{
					ID:      "virtual:agents",
					Name:    "★ AGENTS.md",
					Path:    "AGENTS.md",
					RawBody: string(content),
				}
				// Prepend to the slice
				skills = append([]types.Skill{virtualAgent}, skills...)
			}
		}

		return skillsLoadedMsg(skills)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ecosystemMsg:
		if msg.err != nil {
			m.StatusMsg = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.StatusMsg = "Ecosistema instanciado"
		}
		return m, nil
	case skillsLoadedMsg:
		newModel, cmd := m.List.Update(msg)
		m.List = newModel.(ListModel)
		return m, cmd

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		m.Installer.Width = m.Width
		m.Installer.Height = m.Height
		m.List.Width = m.Width
		m.List.Height = m.Height

		listWidth := int(float64(m.Width) * 0.4)
		previewWidth := m.Width - listWidth

		searchBarHeight := 3
		m.List.list.SetSize(listWidth, m.Height-searchBarHeight-2)
		m.List.viewport.Width = previewWidth
		m.List.viewport.Height = m.Height - 2

		m.storageList.SetSize(m.Width-4, m.Height-8)
		m.projectList.SetSize(m.Width-4, m.Height-8)

		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case editRequestMsg:
		m.selected = msg.Skill
		m.PrevScreen = m.Screen
		m.Screen = ScreenDetail
		m.setupInputs()
		return m, nil

	case syncRequestMsg:
		m.PrevScreen = m.Screen
		m.Screen = ScreenSyncing
		m.SyncFailed = false
		m.syncOutput = "Preparando instalación..."
		return m, runInstallerCmd(m)

	case runner.SyncResult:
		m.SyncFinished = true
		m.syncOutput = fmt.Sprintf("Exit: %d\nOutput: %s\nErr: %s", msg.ExitCode, msg.Stdout, msg.Stderr)
		if msg.ExitCode != 0 {
			m.SyncFailed = true
		} else {
			m.syncOutput = "✅ Sync Successful!\n\n" + m.syncOutput
			m.Progress.SetPercent(1.0)

			// Register project after successful sync
			absRoot, _ := filepath.Abs(m.rootPath)
			_ = m.backend.RegisterProject(absRoot)
		}
		return m, nil

	case installerProgressMsg:
		m.syncOutput = msg.task
		cmd = m.Progress.SetPercent(msg.percent)
		return m, tea.Batch(cmd, nextInstallerStep(m, msg.percent))

	case progress.FrameMsg:
		progressModel, cmd := m.Progress.Update(msg)
		m.Progress = progressModel.(progress.Model)
		return m, cmd

	case installerFinishedMsg:
		if msg.err != nil {
			m.StatusMsg = fmt.Sprintf("Error: %v", msg.err)
			m.syncOutput = fmt.Sprintf("Error: %v", msg.err)
			m.SyncFailed = true
			return m, nil // Stay on ScreenSyncing so user can read error
		}
		m.StatusMsg = "Instalación completada con éxito"
		m.Screen = ScreenHome
		return m, m.loadSkills()

	case storedSkillsLoadedMsg:
		var items []list.Item
		m.storedSkills = []storage.StoredSkill(msg)
		for _, s := range msg {
			items = append(items, storageItem{stored: s})
		}
		cmd = m.storageList.SetItems(items)
		if m.Screen == ScreenInstaller {
			m.Installer.AllStored = m.storedSkills
			m.Installer.StoredSkills = make([]bool, len(msg))
		}
		return m, cmd

	case projectsLoadedMsg:
		var items []list.Item
		for _, p := range msg {
			items = append(items, projectItem{project: p})
		}
		cmd = m.projectList.SetItems(items)
		return m, cmd

	case statusMsg:
		m.StatusMsg = string(msg)
		if m.Screen == ScreenList {
			cmd = m.List.list.NewStatusMessage(string(msg))
			return m, cmd
		}
		return m, nil

	case errorMsg:
		m.err = msg
		return m, nil

	case syncReportMsg:
		m.SyncFinished = true
		m.syncReport = msg.report
		m.err = msg.err
		if msg.err != nil {
			m.SyncFailed = true
		}
		return m, nil

	case syncFinishedMsg:
		if msg.err != nil {
			m.StatusMsg = fmt.Sprintf("Error en sincronización: %v", msg.err)
		} else {
			m.StatusMsg = "OpenCode synchronization successful"
		}
		return m, nil

	case deleteSkillFinishedMsg:
		updated, _ := m.deleteConfirm.Update(msg)
		m.deleteConfirm = updated
		if m.deleteConfirm.success {
			m.Screen = m.PrevScreen
			m.StatusMsg = fmt.Sprintf("Skill '%s' deleted.", msg.name)
			m.resetDeleteConfirm()
			return m, m.loadSkills()
		}
		// On error, stay on confirm screen to show error
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
	case ScreenHome:
		return m.handleHomeKeys(msg)
	case ScreenList:
		return m.handleListKeys(msg)
	case ScreenDetail:
		return m.handleDetailKeys(msg)
	case ScreenSyncing:
		return m.handleSyncingKeys(msg)
	case ScreenContentView:
		return m.handleContentViewKeys(msg)
	case ScreenInstaller:
		return m.handleInstallerKeys(msg)
	case ScreenStorage:
		return m.handleStorageKeys(msg)
	case ScreenProjects:
		return m.handleProjectsKeys(msg)
	case ScreenDeleteConfirm:
		return m.handleDeleteConfirmKeys(msg)
	}

	return m, nil
}

func (m Model) handleHomeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.HomeCursor > 0 {
			m.HomeCursor--
		}
	case "down", "j":
		if m.HomeCursor < 4 {
			m.HomeCursor++
		}
	case "enter", " ":
		if m.HomeCursor == 0 {
			m.Screen = ScreenInstaller
			m.Installer.Cursor = 0
			m.Installer.AllStored = m.storedSkills
			m.Installer.StoredSkills = make([]bool, len(m.storedSkills))
			return m, m.loadStoredSkillsCmd()
		} else if m.HomeCursor == 1 {
			m.Screen = ScreenList
			return m, m.loadSkills()
		} else if m.HomeCursor == 2 {
			m.Screen = ScreenStorage
			return m, m.loadStoredSkillsCmd()
		} else if m.HomeCursor == 3 {
			m.StatusMsg = "Sincronizando con OpenCode..."
			return m, m.syncOpenCodeCmd()
		} else if m.HomeCursor == 4 {
			m.Screen = ScreenProjects
			return m, m.loadProjectsCmd()
		}
	case "1":
		m.HomeCursor = 0
		m.Screen = ScreenInstaller
		m.Installer.Cursor = 0
		m.Installer.AllStored = m.storedSkills
		m.Installer.StoredSkills = make([]bool, len(m.storedSkills))
		return m, m.loadStoredSkillsCmd()
	case "2":
		m.HomeCursor = 1
		m.Screen = ScreenList
		return m, m.loadSkills()
	case "3":
		m.HomeCursor = 2
		m.Screen = ScreenStorage
		return m, m.loadStoredSkillsCmd()
	case "4":
		m.HomeCursor = 3
		m.StatusMsg = "Sincronizando con OpenCode..."
		return m, m.syncOpenCodeCmd()
	case "5":
		m.HomeCursor = 4
		m.Screen = ScreenProjects
		return m, m.loadProjectsCmd()
	case "esc", "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) syncOpenCodeCmd() tea.Cmd {
	return func() tea.Msg {
		err := m.backend.RegisterOpenCodeTools()
		return syncFinishedMsg{err: err}
	}
}

func (m Model) handleListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	newModel, cmd := m.List.Update(msg)
	m.List = newModel.(ListModel)

	// Coordination logic
	if m.List.selected != nil {
		m.selected = m.List.selected
	}

	switch msg.String() {
	case "enter", "v":
		if m.List.selected != nil {
			m.PrevScreen = m.Screen
			m.Screen = ScreenContentView
			m.List.Width = m.Width
			m.List.Height = m.Height
			m.List.viewport.Width = m.Width
			m.List.viewport.Height = m.Height - 6
			return m, cmd
		}
	case "e":
		if m.List.selected != nil {
			m.PrevScreen = m.Screen
			m.Screen = ScreenDetail
			m.setupInputs()
			return m, cmd
		}
	case "y":
		m.PrevScreen = m.Screen
		m.Screen = ScreenSyncing
		m.SyncFailed = false
		m.SyncFinished = false
		return m, m.startSync()
	case "s":
		return m, m.saveToStorageCmd()
	case "d":
		if m.List.selected != nil {
			name := m.List.selected.Name
			if !strings.HasPrefix(m.List.selected.ID, "virtual:") && !isCoreSkill(name) {
				m.deleteConfirm.skillName = name
				m.PrevScreen = m.Screen
				m.Screen = ScreenDeleteConfirm
				return m, nil
			}
		}
	case "esc":
		m.Screen = ScreenHome
		return m, nil
	}

	return m, cmd
}

func (m Model) handleDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	isReadOnly := m.selected != nil && m.selected.ID == "virtual:agents"

	switch msg.String() {
	case "esc":
		m.Screen = m.PrevScreen
		return m, nil
	case "ctrl+s":
		if isReadOnly {
			return m, nil
		}
		m.selected.Metadata.Description = m.inputs[0].Value()
		m.selected.Metadata.Scope = m.inputs[1].Value()
		m.selected.RawBody = m.inputs[2].Value()
		m.Screen = m.PrevScreen
		return m, m.saveSkill()
	case "tab":
		if m.inputs[0].Focused() {
			m.inputs[0].Blur()
			m.inputs[1].Focus()
		} else if m.inputs[1].Focused() {
			m.inputs[1].Blur()
			m.inputs[2].Focus()
		} else {
			m.inputs[2].Blur()
			m.inputs[0].Focus()
		}
		return m, nil
	case "enter":
		if m.inputs[2].Focused() {
			// Allow Enter for content textarea
			break
		}
		// Suppress Enter for Description/Scope to prevent newlines
		return m, nil
	}

	if isReadOnly {
		return m, nil
	}

	var cmd tea.Cmd
	for i := range m.inputs {
		if m.inputs[i].Focused() {
			m.inputs[i], cmd = m.inputs[i].Update(msg)
			break
		}
	}
	return m, cmd
}

func (m Model) handleSyncingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		if m.SyncFailed {
			m.Screen = ScreenHome
			return m, nil
		}
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
	case "e":
		if m.selected != nil {
			m.PrevScreen = m.Screen
			m.Screen = ScreenDetail
			m.setupInputs()
			return m, nil
		}
	case "j", "k", "up", "down", "pgup", "pgdown":
		var cmd tea.Cmd
		m.List.viewport, cmd = m.List.viewport.Update(msg)
		return m, cmd
	}

	newModel, cmd := m.List.Update(msg)
	m.List = newModel.(ListModel)

	// Sync header if list selection changed
	if i, ok := m.List.list.SelectedItem().(item); ok {
		m.selected = &i.skill
	}

	return m, cmd
}

func (m Model) handleInstallerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" || msg.String() == "q" {
		m.Screen = ScreenHome
		return m, nil
	}

	newInstaller, cmd := m.Installer.Update(msg)
	m.Installer = newInstaller.(InstallerModel)
	return m, cmd
}

func (m Model) handleComponentUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.Screen {
	case ScreenList, ScreenContentView:
		newModel, cmd := m.List.Update(msg)
		m.List = newModel.(ListModel)
		return m, cmd
	case ScreenDetail:
		for i := range m.inputs {
			m.inputs[i], cmd = m.inputs[i].Update(msg)
		}
	case ScreenStorage:

		m.storageList, cmd = m.storageList.Update(msg)
	case ScreenProjects:
		m.projectList, cmd = m.projectList.Update(msg)
	}
	return m, cmd
}

func (m Model) handleProjectsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.Screen = ScreenHome
		m.HomeCursor = 4
		return m, nil
	case "r":
		m.StatusMsg = "Buscando proyectos..."
		return m, m.scanProjectsCmd()
	}

	var cmd tea.Cmd
	m.projectList, cmd = m.projectList.Update(msg)
	return m, cmd
}

func (m Model) scanProjectsCmd() tea.Cmd {
	return func() tea.Msg {
		// Scan relevant roots: HOME and parent of current root
		var roots []string
		if home, err := os.UserHomeDir(); err == nil {
			roots = append(roots, home)
		}
		if absRoot, err := filepath.Abs(m.rootPath); err == nil {
			roots = append(roots, filepath.Dir(absRoot))
		}

		// Depth 3 is a good balance
		paths, err := m.backend.ScanProjects(roots, 3)
		if err != nil {
			return errorMsg(err)
		}

		// Register found projects
		for _, p := range paths {
			_ = m.backend.RegisterProjectInitial(p)
		}

		// Reload list from storage (which filters dead paths)
		return m.loadProjectsCmd()()
	}
}

func (m *Model) setupInputs() {
	m.inputs = make([]textarea.Model, 3)
	isReadOnly := m.selected != nil && m.selected.ID == "virtual:agents"

	// Overhead calculation for ScreenDetail to ensure content fits
	// Title: 2
	// Labels: 3
	// Metadata areas: 4 (Desc) + 3 (Scope) = 7
	// Gaps: 3
	// Footer: 2
	// Total: 17
	overhead := 17
	contentHeight := m.Height - overhead
	if contentHeight < 4 {
		contentHeight = 4 // Absolute minimum
	}

	d := textarea.New()
	d.Placeholder = "Description"
	d.SetWidth(m.Width - 6)
	d.SetHeight(2)
	d.FocusedStyle.Base = focusedTextareaStyle
	d.BlurredStyle.Base = blurredTextareaStyle
	if m.selected != nil {
		d.SetValue(m.selected.Metadata.Description)
	}
	if !isReadOnly {
		d.Focus()
	}
	m.inputs[0] = d

	s := textarea.New()
	s.Placeholder = "Scope"
	s.SetWidth(m.Width - 6)
	s.SetHeight(1)
	s.FocusedStyle.Base = focusedTextareaStyle
	s.BlurredStyle.Base = blurredTextareaStyle
	if m.selected != nil {
		s.SetValue(m.selected.Metadata.Scope)
	}
	m.inputs[1] = s

	c := textarea.New()
	c.Placeholder = "Content (SKILL.md)"
	c.SetWidth(m.Width - 6)
	c.SetHeight(contentHeight)
	c.FocusedStyle.Base = focusedTextareaStyle
	c.BlurredStyle.Base = blurredTextareaStyle
	if m.selected != nil {
		c.SetValue(m.selected.RawBody)
	}
	m.inputs[2] = c
}

func (m Model) startSync() tea.Cmd {
	return func() tea.Msg {
		// Ensure core shared library and AGENTS.md exist before syncing
		_ = m.backend.InstallCoreSkill("skill-sync")
		_ = m.backend.EnsureAgentsMD()

		opts := syncengine.SyncOptions{
			DryRun: false,
		}

		report, err := m.backend.Sync(m.rootPath, opts)
		return syncReportMsg{report: report, err: err}
	}
}

func (m Model) saveSkill() tea.Cmd {
	if m.selected == nil {
		return nil
	}
	m.selected.Metadata.Description = m.inputs[0].Value()
	m.selected.Metadata.Scope = m.inputs[1].Value()
	m.selected.RawBody = m.inputs[2].Value()

	return func() tea.Msg {
		_ = m.backend.SaveSkill(m.selected.Path, m.selected)
		return tea.Batch(
			func() tea.Msg { return statusMsg(fmt.Sprintf("Skill '%s' guardada correctamente", m.selected.Name)) },
			m.loadSkills(),
		)()
	}
}

func (m Model) loadStoredSkillsCmd() tea.Cmd {
	return func() tea.Msg {
		skills, _ := m.backend.ListStoredSkills()
		return storedSkillsLoadedMsg(skills)
	}
}

func (m Model) loadProjectsCmd() tea.Cmd {
	return func() tea.Msg {
		projects, _ := m.backend.GetProjects()
		return projectsLoadedMsg(projects)
	}
}

func (m Model) saveToStorageCmd() tea.Cmd {
	target := m.selected
	if target == nil {
		if i, ok := m.List.list.SelectedItem().(item); ok {
			target = &i.skill
		}
	}

	if target == nil {
		return nil
	}

	// Create a copy to avoid race conditions if needed
	skill := *target

	// Metadata from the current context
	absPath, _ := filepath.Abs(m.rootPath)
	metadata := storage.StoredMetadata{
		SkillName:     skill.Name,
		Description:   skill.Metadata.Description,
		OriginProject: absPath,
		OriginPath:    skill.Path,
		SavedAt:       time.Now(),
	}

	return func() tea.Msg {
		_ = m.backend.SaveToStorage(&skill, metadata)
		return statusMsg(fmt.Sprintf("Skill '%s' guardada en almacenamiento global", skill.Name))
	}
}

func (m Model) handleStorageKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.Screen = ScreenHome
		m.HomeCursor = 2
		return m, nil
	case "i":
		if itm, ok := m.storageList.SelectedItem().(storageItem); ok {
			m.PrevScreen = m.Screen
			m.Screen = ScreenSyncing
			m.SyncFailed = false
			m.SyncFinished = false
			return m, m.installFromStorageAndSyncCmd(itm.stored)
		}
	case "d":
		if itm, ok := m.storageList.SelectedItem().(storageItem); ok {
			name := itm.stored.Metadata.SkillName
			if !isCoreSkill(name) {
				m.deleteConfirm.skillName = name
				m.PrevScreen = m.Screen
				m.Screen = ScreenDeleteConfirm
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.storageList, cmd = m.storageList.Update(msg)
	return m, cmd
}

func (m Model) handleDeleteConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	updated, cmd := m.deleteConfirm.Update(msg)
	m.deleteConfirm = updated

	// User cancelled (n/esc) — go back
	if !m.deleteConfirm.deleting && !m.deleteConfirm.success && m.deleteConfirm.err == nil {
		// Only go back if a key was actually pressed (not an unrelated msg)
		if msg.String() == "n" || msg.String() == "N" || msg.String() == "esc" {
			m.Screen = m.PrevScreen
			m.resetDeleteConfirm()
			return m, nil
		}
	}

	return m, cmd
}

func (m Model) installFromStorageAndSyncCmd(stored storage.StoredSkill) tea.Cmd {
	return func() tea.Msg {
		// 1. Load from storage
		content, err := m.backend.LoadFromStorage(stored.ID)
		if err != nil {
			return runner.SyncResult{
				ExitCode: 1,
				Stderr:   fmt.Sprintf("Failed to load skill from storage: %v", err),
			}
		}

		// 2. Parse content to get metadata (especially Skill Name)
		// We use parser.ParseContent to handle the raw SKILL.md content
		skill, err := m.backend.ParseSkillContent(content)
		if err != nil {
			return runner.SyncResult{
				ExitCode: 1,
				Stderr:   fmt.Sprintf("Malformed YAML: Edit and fix\n\nDetails: %v", err),
			}
		}

		// Ensure we use the name from stored metadata if parser fails to find one
		if skill.Name == "" {
			skill.Name = stored.Metadata.SkillName
		}

		// 3. Write to local project
		skillDir := filepath.Join(m.rootPath, ".agents", "skills", skill.Name)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return runner.SyncResult{
				ExitCode: 1,
				Stderr:   fmt.Sprintf("Failed to create skill directory %s: %v", skillDir, err),
			}
		}

		skillPath := filepath.Join(skillDir, "SKILL.md")
		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			return runner.SyncResult{
				ExitCode: 1,
				Stderr:   fmt.Sprintf("Failed to write skill file %s: %v", skillPath, err),
			}
		}

		// 4. Trigger Sync
		return m.startSync()()
	}
}

func (m *Model) filterSkills(query string) {
	m.List.filterSkills(query)
}
