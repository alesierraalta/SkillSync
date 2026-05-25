package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"skillsync/tui/internal/remove"
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/syncengine"
	"skillsync/tui/internal/types"
	"strings"
	"testing"
)

// MockAppService implements AppService for testing.
type MockAppService struct {
	DiscoverSkillsFunc         func(rootPath string) ([]string, error)
	ScanProjectsFunc           func(roots []string, depth int) ([]string, error)
	ParseSkillFunc             func(path string) (*types.Skill, error)
	ParseSkillContentFunc      func(content string) (*types.Skill, error)
	SaveSkillFunc              func(path string, skill *types.Skill) error
	SyncFunc                   func(root string, opts syncengine.SyncOptions) (*runner.SyncReport, error)
	RegisterProjectFunc        func(path string) error
	RegisterProjectInitialFunc func(path string) error
	ListStoredSkillsFunc       func() ([]storage.StoredSkill, error)
	GetProjectsFunc            func() ([]storage.ProjectInfo, error)
	SaveToStorageFunc          func(skill *types.Skill, metadata storage.StoredMetadata) error
	LoadFromStorageFunc        func(id string) (string, error)
	InstallCoreSkillFunc       func(name string) error
	RegisterOpenCodeToolsFunc  func() error
	RegisterSkillManagerAgentFunc func() error
	EnsureAgentsMDFunc        func() error
	RemoveSkillFunc           func(name string, opts remove.Options) error
}

func (m *MockAppService) DiscoverSkills(rootPath string) ([]string, error) {
	return m.DiscoverSkillsFunc(rootPath)
}

func (m *MockAppService) ScanProjects(roots []string, depth int) ([]string, error) {
	return m.ScanProjectsFunc(roots, depth)
}

func (m *MockAppService) ParseSkill(path string) (*types.Skill, error) {
	return m.ParseSkillFunc(path)
}

func (m *MockAppService) ParseSkillContent(content string) (*types.Skill, error) {
	return m.ParseSkillContentFunc(content)
}

func (m *MockAppService) SaveSkill(path string, skill *types.Skill) error {
	return m.SaveSkillFunc(path, skill)
}

func (m *MockAppService) Sync(root string, opts syncengine.SyncOptions) (*runner.SyncReport, error) {
	return m.SyncFunc(root, opts)
}

func (m *MockAppService) RegisterProject(path string) error {
	return m.RegisterProjectFunc(path)
}

func (m *MockAppService) RegisterProjectInitial(path string) error {
	return m.RegisterProjectInitialFunc(path)
}

func (m *MockAppService) ListStoredSkills() ([]storage.StoredSkill, error) {
	return m.ListStoredSkillsFunc()
}

func (m *MockAppService) GetProjects() ([]storage.ProjectInfo, error) {
	return m.GetProjectsFunc()
}

func (m *MockAppService) SaveToStorage(skill *types.Skill, metadata storage.StoredMetadata) error {
	return m.SaveToStorageFunc(skill, metadata)
}

func (m *MockAppService) LoadFromStorage(id string) (string, error) {
	return m.LoadFromStorageFunc(id)
}

func (m *MockAppService) InstallCoreSkill(name string) error {
	if m.InstallCoreSkillFunc != nil {
		return m.InstallCoreSkillFunc(name)
	}
	return nil
}

func (m *MockAppService) RegisterOpenCodeTools() error {
	if m.RegisterOpenCodeToolsFunc != nil {
		return m.RegisterOpenCodeToolsFunc()
	}
	return nil
}

func (m *MockAppService) RegisterSkillManagerAgent() error {
	if m.RegisterSkillManagerAgentFunc != nil {
		return m.RegisterSkillManagerAgentFunc()
	}
	return nil
}

func (m *MockAppService) EnsureAgentsMD() error {
	if m.EnsureAgentsMDFunc != nil {
		return m.EnsureAgentsMDFunc()
	}
	return nil
}

func (m *MockAppService) RemoveSkill(name string, opts remove.Options) error {
	if m.RemoveSkillFunc != nil {
		return m.RemoveSkillFunc(name, opts)
	}
	return nil
}

func TestMockAppService_InterfaceCompliance(t *testing.T) {
	// Task 1.1 RED: This will fail to compile until AppService is defined.
	var _ AppService = (*MockAppService)(nil)
}

func TestInstallerModel_Update_Navigation(t *testing.T) {
	// Task 2.1 RED: Tests for InstallerModel navigation
	m := InstallerModel{
		Cursor:    0,
		Providers: []bool{true, false, true},
		Skills:    []bool{true, true},
	}

	// Move down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	newModel, _ := m.Update(msg)
	im := newModel.(InstallerModel)

	if im.Cursor != 1 {
		t.Errorf("expected Cursor 1, got %d", im.Cursor)
	}

	// Move up
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	newModel, _ = im.Update(msg)
	im = newModel.(InstallerModel)

	if im.Cursor != 0 {
		t.Errorf("expected Cursor 0, got %d", im.Cursor)
	}
}

func TestInstallerModel_Update_Toggle(t *testing.T) {
	// Task 2.1 RED: Tests for InstallerModel toggling
	m := InstallerModel{
		Cursor:    0,
		Providers: []bool{true, false, true},
	}

	// Toggle first provider
	msg := tea.KeyMsg{Type: tea.KeySpace}
	newModel, _ := m.Update(msg)
	im := newModel.(InstallerModel)

	if im.Providers[0] != false {
		t.Errorf("expected provider 0 to be false, got %v", im.Providers[0])
	}
}

func TestListModel_Update_Navigation(t *testing.T) {
	// Task 3.1 RED: Tests for ListModel navigation
	m := ListModel{
		list: list.New([]list.Item{
			item{skill: types.Skill{Name: "Skill 1"}},
			item{skill: types.Skill{Name: "Skill 2"}},
		}, list.NewDefaultDelegate(), 0, 0),
	}

	// Move down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	newModel, _ := m.Update(msg)
	lm := newModel.(ListModel)

	if lm.list.Index() != 1 {
		t.Errorf("expected Index 1, got %d", lm.list.Index())
	}
}

func TestListModel_Update_SkillsLoaded(t *testing.T) {
	// Task 3.1 RED: Test handling skillsLoadedMsg
	m := NewListModel(nil, ".")
	skills := []types.Skill{
		{Name: "New Skill", Path: "path/to/skill"},
	}

	msg := skillsLoadedMsg(skills)
	newModel, _ := m.Update(msg)
	lm := newModel.(ListModel)

	if len(lm.list.Items()) != 1 {
		t.Errorf("expected 1 item, got %d", len(lm.list.Items()))
	}
}

func TestModel_UpdateDelegation(t *testing.T) {
	// Task 4.1 RED: Test root Model.Update delegation
	backend := &MockAppService{}
	m := NewModel(backend)

	t.Run("Delegates to InstallerModel when ScreenInstaller", func(t *testing.T) {
		m.Screen = ScreenInstaller
		m.Installer.Cursor = 0

		// Move down in Installer
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
		newModel, _ := m.Update(msg)
		root := newModel.(Model)

		if root.Installer.Cursor != 1 {
			t.Errorf("expected Installer.Cursor 1, got %d", root.Installer.Cursor)
		}
	})

	t.Run("Delegates to ListModel when ScreenList", func(t *testing.T) {
		m.Screen = ScreenList
		m.List = NewListModel(backend, ".")
		m.List.list.SetItems([]list.Item{
			item{skill: types.Skill{Name: "Skill 1"}},
			item{skill: types.Skill{Name: "Skill 2"}},
		})
		m.List.list.Select(0)

		// Move down in List
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
		newModel, _ := m.Update(msg)
		root := newModel.(Model)

		if root.List.list.Index() != 1 {
			t.Errorf("expected List.list.Index 1, got %d", root.List.list.Index())
		}
	})
}

func TestModel_ViewDelegation(t *testing.T) {
	// Task 4.1 RED: Test root Model.View delegation
	backend := &MockAppService{}
	m := NewModel(backend)
	m.Width = 100
	m.Height = 30

	t.Run("Delegates to InstallerModel View", func(t *testing.T) {
		m.Screen = ScreenInstaller
		m.Installer.Cursor = 0
		view := m.View()
		if !strings.Contains(view, "SYNCK INSTALLER") {
			t.Errorf("expected Installer view content, got %q", view)
		}
	})

	t.Run("Delegates to ListModel View", func(t *testing.T) {
		m.Screen = ScreenList
		view := m.View()
		if !strings.Contains(view, "Search skills...") {
			t.Errorf("expected List view content (search placeholder), got %q", view)
		}
	})
}
