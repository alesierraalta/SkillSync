package ui

import (
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/syncengine"
	"skillsync/tui/internal/types"
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

func TestMockAppService_InterfaceCompliance(t *testing.T) {
	// Task 1.1 RED: This will fail to compile until AppService is defined.
	var _ AppService = (*MockAppService)(nil)
}
