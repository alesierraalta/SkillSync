package ui

import (
	"skillsync/tui/internal/discovery"
	"skillsync/tui/internal/parser"
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/syncengine"
	"skillsync/tui/internal/types"
)

// AppService abstracts core business logic for the UI.
type AppService interface {
	DiscoverSkills(rootPath string) ([]string, error)
	ScanProjects(roots []string, depth int) ([]string, error)
	ParseSkill(path string) (*types.Skill, error)
	ParseSkillContent(content string) (*types.Skill, error)
	SaveSkill(path string, skill *types.Skill) error

	Sync(root string, opts syncengine.SyncOptions) (*runner.SyncReport, error)

	RegisterProject(path string) error
	RegisterProjectInitial(path string) error
	ListStoredSkills() ([]storage.StoredSkill, error)
	GetProjects() ([]storage.ProjectInfo, error)
	SaveToStorage(skill *types.Skill, metadata storage.StoredMetadata) error
	LoadFromStorage(id string) (string, error)
}

// Backend implements AppService using concrete implementations.
type Backend struct {
	storage *storage.Service
}

func NewBackend(storageService *storage.Service) *Backend {
	return &Backend{
		storage: storageService,
	}
}

func (b *Backend) DiscoverSkills(rootPath string) ([]string, error) {
	return discovery.DiscoverSkills(rootPath)
}

func (b *Backend) ScanProjects(roots []string, depth int) ([]string, error) {
	return discovery.ScanProjects(roots, depth)
}

func (b *Backend) ParseSkill(path string) (*types.Skill, error) {
	return parser.Parse(path)
}

func (b *Backend) ParseSkillContent(content string) (*types.Skill, error) {
	return parser.ParseContent(content)
}

func (b *Backend) SaveSkill(path string, skill *types.Skill) error {
	return parser.Save(path, skill)
}

func (b *Backend) Sync(root string, opts syncengine.SyncOptions) (*runner.SyncReport, error) {
	opts.Storage = b.storage
	return syncengine.Sync(root, opts)
}

func (b *Backend) RegisterProject(path string) error {
	return b.storage.RegisterProject(path)
}

func (b *Backend) RegisterProjectInitial(path string) error {
	return b.storage.RegisterProjectInitial(path)
}

func (b *Backend) ListStoredSkills() ([]storage.StoredSkill, error) {
	return b.storage.List()
}

func (b *Backend) GetProjects() ([]storage.ProjectInfo, error) {
	return b.storage.GetProjects()
}

func (b *Backend) SaveToStorage(skill *types.Skill, metadata storage.StoredMetadata) error {
	return b.storage.Save(skill, metadata)
}

func (b *Backend) LoadFromStorage(id string) (string, error) {
	return b.storage.Load(id)
}
