package ui

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"skillsync/tui/internal/agentdetect"
	"skillsync/tui/internal/bundle"
	"skillsync/tui/internal/coreskills"
	"skillsync/tui/internal/discovery"
	"skillsync/tui/internal/opencode"
	"skillsync/tui/internal/parser"
	"skillsync/tui/internal/remove"
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/syncengine"
	"skillsync/tui/internal/types"
	"skillsync/tui/internal/vault"
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
	CopyStorageExtras(id, dstDir string) error

	InstallCoreSkill(name string) error
	RegisterOpenCodeTools() error
	RegisterSkillManagerAgent() error
	EnsureAgentsMD(root string) error

	RemoveSkill(name string, opts remove.Options) error
	RemoveGlobalSkill(path string) error

	// ExportBundle writes the named vault skills to a .skillsync bundle at
	// destPath and returns the bundle path.
	ExportBundle(names []string, destPath string) (string, error)
	// ImportBundle installs the skills from a .skillsync bundle into the
	// project at projectRoot and returns a per-skill result summary.
	ImportBundle(bundlePath, projectRoot string) ([]bundle.ImportResult, error)

	// DetectAgentEcosystem returns a read-only inventory of installed AI agent tools.
	DetectAgentEcosystem() ([]agentdetect.AgentInfo, error)
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

// RegisterOpenCodeTools preserves the package-level API used by the CLI while
// routing the behavior through the AppService implementation.
func RegisterOpenCodeTools() error {
	return NewBackend(storage.NewService("")).RegisterOpenCodeTools()
}

// InstallCoreSkill preserves the package-level API used by the CLI while
// routing the behavior through the AppService implementation.
func InstallCoreSkill(name string) error {
	return NewBackend(storage.NewService("")).InstallCoreSkill(name)
}

// RegisterSkillManagerAgent preserves the package-level API used by the CLI
// while routing the behavior through the AppService implementation.
func RegisterSkillManagerAgent() error {
	return NewBackend(storage.NewService("")).RegisterSkillManagerAgent()
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

func (b *Backend) CopyStorageExtras(id, dstDir string) error {
	return b.storage.CopyExtras(id, dstDir)
}

func (b *Backend) InstallCoreSkill(name string) error {
	if name == "skill-sync" {
		if err := b.installCoreSharedLib(); err != nil {
			return err
		}
	}

	srcDir := "skills/" + name
	providers := b.activeProviders()
	if len(providers) == 0 {
		// Fallback: install to default .agents only
		providers = []string{".agents/skills"}
	}

	for _, destBase := range providers {
		destDir := filepath.Join(destBase, name)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return err
		}

		if err := fs.WalkDir(coreskills.EmbeddedSkills, srcDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(srcDir, path)
			if err != nil {
				return err
			}

			targetPath := filepath.Join(destDir, relPath)

			if d.IsDir() {
				return os.MkdirAll(targetPath, 0755)
			}

			data, err := coreskills.EmbeddedSkills.ReadFile(path)
			if err != nil {
				return err
			}

			if filepath.Ext(path) == ".sh" {
				data = b.ensureLF(data)
			}

			return os.WriteFile(targetPath, data, 0644)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (b *Backend) installCoreSharedLib() error {
	libDir := filepath.Join(".agents", "skills", "lib")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		return err
	}
	utilsPath := filepath.Join(libDir, "utils.sh")

	embedded, err := coreskills.EmbeddedSkills.ReadFile("skills/lib/utils.sh")
	if err != nil {
		return err
	}
	embedded = b.ensureLF(embedded)

	existing, err := os.ReadFile(utilsPath)
	if err != nil {
		return os.WriteFile(utilsPath, embedded, 0644)
	}

	if bytes.Equal(existing, embedded) {
		return nil
	}

	repaired := b.ensureLF(existing)
	if bytes.Equal(existing, repaired) {
		return nil
	}

	return os.WriteFile(utilsPath, repaired, 0644)
}

func (b *Backend) RemoveSkill(name string, opts remove.Options) error {
	svc := remove.Service{
		RootPath: ".",
		Storage:  b.storage,
	}
	if err := svc.RemoveByID(name, opts); err != nil {
		return err
	}

	// Post-delete regeneration is best-effort; errors are non-fatal
	if err := RegenerateAfterDelete("."); err != nil {
		// Log the error but don't revert the deletion
		_ = err
	}
	return nil
}

// ExportBundle writes the named vault skills to a .skillsync bundle at destPath.
// The vault is the same storage path the backend already manages.
func (b *Backend) ExportBundle(names []string, destPath string) (string, error) {
	if destPath != "" {
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return "", fmt.Errorf("create bundle dir: %w", err)
		}
	}
	v := vault.NewServiceWithRoot(b.storage.RootPath)
	return bundle.Export(v, names, destPath)
}

// ImportBundle installs the bundle's skills into the project at projectRoot,
// overwriting existing skills of the same name.
func (b *Backend) ImportBundle(bundlePath, projectRoot string) ([]bundle.ImportResult, error) {
	return bundle.Import(bundlePath, bundle.ImportOptions{
		ProjectRoot: projectRoot,
		OnDuplicate: bundle.DuplicateOverwrite,
	})
}

func (b *Backend) activeProviders() []string {
	providers := []string{
		".agents/skills",
		".opencode/skills",
		".claude/skills",
		".gemini/skills",
		".cursor/skills",
		".copilot/skills",
		".qwen/skills",
	}
	var active []string
	for _, p := range providers {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			active = append(active, p)
		}
	}
	return active
}

func (b *Backend) ensureLF(data []byte) []byte {
	return bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
}

func (b *Backend) RegisterOpenCodeTools() error {
	// Sync skills first
	if _, err := opencode.SyncSkills(".", opencode.Options{}); err != nil {
		// Non-fatal: continue with tool registration even if sync fails
	}

	// Parse mirrored skills
	skills, err := b.parseMirroredSkillsForEcosystem()
	if err != nil {
		return err
	}

	// Add 5 base commands as pseudo-skills
	baseSkills := b.getBaseCommandSkills()
	skills = append(baseSkills, skills...)

	if _, err := opencode.RegenerateTools(".", skills, false); err != nil {
		return err
	}

	_, err = opencode.RegenerateCommands(".", skills, false)
	return err
}

func (b *Backend) parseMirroredSkillsForEcosystem() ([]types.Skill, error) {
	skillsPath := filepath.Join(".", ".opencode", "skills")
	var skillPaths []string
	err := filepath.WalkDir(skillsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil
		}
		if info.IsDir() {
			skillFile := filepath.Join(path, "SKILL.md")
			if _, err := os.Stat(skillFile); err == nil {
				skillPaths = append(skillPaths, skillFile)
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	var skills []types.Skill
	for _, p := range skillPaths {
		skill, err := b.parseSkill(p)
		if err != nil {
			continue
		}
		skills = append(skills, *skill)
	}
	return skills, nil
}

func (b *Backend) parseSkill(path string) (*types.Skill, error) {
	return parser.Parse(path)
}

func (b *Backend) getBaseCommandSkills() []types.Skill {
	return []types.Skill{
		{Name: "skill", Metadata: types.Metadata{Description: "Entry point for skill management", AutoInvoke: []string{"skill"}}},
		{Name: "find", Metadata: types.Metadata{Description: "Search and list existing skills", AutoInvoke: []string{"find"}}},
		{Name: "create", Metadata: types.Metadata{Description: "Create a new agent skill from a prompt", AutoInvoke: []string{"create"}}},
		{Name: "sync", Metadata: types.Metadata{Description: "Synchronize skills and update AGENTS.md/OPENCODE.md", AutoInvoke: []string{"sync"}}},
		{Name: "fullskills", Metadata: types.Metadata{Description: "Complete skill workflow", AutoInvoke: []string{"fullskills"}}},
	}
}

func (b *Backend) RegisterSkillManagerAgent() error {
	// Sync skills first (non-fatal)
	opencode.SyncSkills(".", opencode.Options{})

	// Parse mirrored skills
	skills, err := b.parseMirroredSkillsForEcosystem()
	if err != nil {
		return err
	}

	// Add base command skills
	baseSkills := b.getBaseCommandSkills()
	skills = append(baseSkills, skills...)

	_, err = opencode.RegenerateAgent(".", skills, false)
	return err
}

// DetectAgentEcosystem delegates to agentdetect.Detect().
func (b *Backend) DetectAgentEcosystem() ([]agentdetect.AgentInfo, error) {
	return agentdetect.Detect()
}

func (b *Backend) EnsureAgentsMD(root string) error {
	if root == "" {
		root = "."
	}
	agentsFile := filepath.Join(root, "AGENTS.md")
	if _, err := os.Stat(agentsFile); os.IsNotExist(err) {
		content := []byte("# Agent Skills\n\nThis document lists the AI skills available in the project.\n\n## Available Skills\n\n| Skill | Description | Location |\n| ----- | ----------- | -------- |\n\n### Auto-invoke Skills\n\nWhen performing these actions, ALWAYS invoke the corresponding skill FIRST:\n\n| Action | Skill |\n| ------ | ----- |\n")
		return os.WriteFile(agentsFile, content, 0644)
	}
	return nil
}

func (b *Backend) RemoveGlobalSkill(path string) error {
	if filepath.Base(path) != "SKILL.md" {
		return fmt.Errorf("invalid skill path: %s", path)
	}
	dir := filepath.Dir(path)
	return os.RemoveAll(dir)
}
