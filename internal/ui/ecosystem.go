package ui

import (
	"io/fs"
	"os"
	"path/filepath"
	"skillsync/tui/internal/coreskills"
	"skillsync/tui/internal/opencode"
	"skillsync/tui/internal/parser"
	"skillsync/tui/internal/types"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ecosystemMsg is the result of ecosystem initialization commands.
// The instantiation process (installing core skills, creating AGENTS.md,
// and registering OpenCode tools) is idempotent — safe to call on every startup.
//
// Idempotency guarantees:
//   - InstallCoreSkill: writes skill files; if already present, overwrites with same content (no-op effectively)
//   - ensureAgentsMD: only creates AGENTS.md if it doesn't exist
//   - registerOpenCodeTools: read-merge-write on package.json; existing tools are preserved, new tools are merged
//   - registerSkillManagerAgent: writes agent file; safe to overwrite with identical content
type ecosystemMsg struct {
	err error
}

func runInstallerCmd(m Model) tea.Cmd {
	return func() tea.Msg {
		// Start with 10%
		return installerProgressMsg{percent: 0.1, task: "Creando directorios de proveedores..."}
	}
}

var coreSkills = map[string]string{
	"skill-creator": "Creates new agent skills following the Agent Skills spec.",
	"skill-sync":    "Syncs skill metadata to AGENTS.md Auto-invoke sections.",
	"find-skills":   "Helps find and install skills from the global or local registry.",
}

func nextInstallerStep(m Model, currentPercent float64) tea.Cmd {
	return func() tea.Msg {
		// Simulation delay
		time.Sleep(500 * time.Millisecond)

		switch currentPercent {
		case 0.1:
			// 1. Create Directories
			providers := []string{".claude/skills/", ".gemini/skills/", ".codex/skills/", ".github/", ".opencode/skills/"}
			for i, p := range m.installerProviders {
				if p {
					_ = os.MkdirAll(providers[i], 0755)
				}
			}
			return installerProgressMsg{percent: 0.3, task: "Instalando Core Skills..."}

		case 0.3:
			// 2. Install Core Skills
			skills := []string{"skill-creator", "skill-sync", "find-skills"}
			for i, enabled := range m.installerSkills {
				if enabled {
					sk := skills[i]
					if err := InstallCoreSkill(sk); err != nil {
						return installerFinishedMsg{err: err}
					}
				}
			}
			return installerProgressMsg{percent: 0.5, task: "Instalando skills desde almacenamiento..."}

		case 0.5:
			// 2.5 Install Stored Skills
			for i, enabled := range m.installerStoredSkills {
				if enabled && i < len(m.storedSkills) {
					stored := m.storedSkills[i]
					content, err := m.storageService.Load(stored.ID)
					if err != nil {
						return installerFinishedMsg{err: err}
					}
					destDir := filepath.Join(".agents", "skills", stored.Metadata.SkillName)
					if err := os.MkdirAll(destDir, 0755); err != nil {
						return installerFinishedMsg{err: err}
					}
					if err := os.WriteFile(filepath.Join(destDir, "SKILL.md"), []byte(content), 0644); err != nil {
						return installerFinishedMsg{err: err}
					}
				}
			}
			return installerProgressMsg{percent: 0.6, task: "Sincronizando configuraciones..."}

		case 0.6:
			// 3. Sync Configs
			agentsFile := "AGENTS.md"
			content, _ := os.ReadFile(agentsFile)
			if len(content) == 0 {
				content = []byte("# Agent Skills\n")
				_ = os.WriteFile(agentsFile, content, 0644)
			}
			
			if m.installerProviders[0] { _ = os.WriteFile("CLAUDE.md", content, 0644) }
			if m.installerProviders[1] { _ = os.WriteFile("GEMINI.md", content, 0644) }
			if m.installerProviders[2] { _ = os.WriteFile("codex.md", content, 0644) }
			if m.installerProviders[3] { 
				_ = os.MkdirAll(".github", 0755)
				_ = os.WriteFile(".github/copilot-instructions.md", content, 0644) 
			}
			if m.installerProviders[4] { 
				_ = os.WriteFile("OPENCODE.md", content, 0644) 
				_ = RegisterOpenCodeTools()
				_ = RegisterSkillManagerAgent()
			}

			if m.installerGlobal {
				// Simular global aliases
				_ = os.WriteFile("GLOBAL_ALIASES.txt", []byte("alias k='skillsync'\nalias ks='skillsync sync'"), 0644)
			}

			return installerProgressMsg{percent: 1.0, task: "Instalación completada"}

		case 1.0:
			return installerFinishedMsg{err: nil}
		}

		return nil
	}
}

func instantiateEcosystemCmd() tea.Cmd {
	return func() tea.Msg {
		skills := []string{"skill-creator", "skill-sync", "find-skills"}

		for _, sk := range skills {
			if err := InstallCoreSkill(sk); err != nil {
				return ecosystemMsg{err: err}
			}
		}

		if err := ensureAgentsMD(); err != nil {
			return ecosystemMsg{err: err}
		}

		// Register OpenCode tools and agents when .opencode/ directory exists
		if _, err := os.Stat(".opencode"); err == nil {
			if err := RegisterOpenCodeTools(); err != nil {
				return ecosystemMsg{err: err}
			}
			if err := RegisterSkillManagerAgent(); err != nil {
				return ecosystemMsg{err: err}
			}
		}

		return ecosystemMsg{err: nil}
	}
}

// activeProviders returns the set of provider skill directories that exist.
// Matches the provider set used by discovery.DiscoverSkills:
// .claude, .opencode, .agents, .gemini, .cursor, .copilot, .qwen
// Skill content lives under <provider>/skills/ for each root.
func activeProviders() []string {
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

func InstallCoreSkill(name string) error {
	if name == "skill-sync" {
		if err := installCoreSharedLib(); err != nil {
			return err
		}
	}

	srcDir := "skills/" + name
	providers := activeProviders()
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

			return os.WriteFile(targetPath, data, 0644)
		}); err != nil {
			return err
		}
	}
	return nil
}

func installCoreSharedLib() error {
	libDir := filepath.Join(".agents", "skills", "lib")
	if err := os.MkdirAll(libDir, 0755); err != nil {
		return err
	}
	utilsPath := filepath.Join(libDir, "utils.sh")
	if _, err := os.Stat(utilsPath); os.IsNotExist(err) {
		data, err := coreskills.EmbeddedSkills.ReadFile("skills/lib/utils.sh")
		if err != nil {
			return err
		}
		return os.WriteFile(utilsPath, data, 0644)
	}
	return nil
}

func ensureAgentsMD() error {
	agentsFile := "AGENTS.md"
	if _, err := os.Stat(agentsFile); os.IsNotExist(err) {
		content := []byte("# Agent Skills\n\nThis document lists the AI skills available in the project.\n\n## Available Skills\n\n| Skill | Description | Location |\n| ----- | ----------- | -------- |\n")
		return os.WriteFile(agentsFile, content, 0644)
	}
	return nil
}

// RegisterOpenCodeTools registers skill-management commands in .opencode/package.json.
// It syncs skills from .agents/skills/ to .opencode/skills/, then regenerates tools
// from the mirrored skills plus the 5 base commands.
func RegisterOpenCodeTools() error {
	// Sync skills first
	if _, err := opencode.SyncSkills(".", opencode.Options{}); err != nil {
		// Non-fatal: continue with tool registration even if sync fails
	}

	// Parse mirrored skills
	skills, err := parseMirroredSkillsForEcosystem()
	if err != nil {
		return err
	}

	// Add 5 base commands as pseudo-skills
	baseSkills := getBaseCommandSkills()
	skills = append(baseSkills, skills...)

	if _, err := opencode.RegenerateTools(".", skills, false); err != nil {
		return err
	}

	_, err = opencode.RegenerateCommands(".", skills, false)
	return err
}

// parseMirroredSkillsForEcosystem parses skills from .opencode/skills/.
func parseMirroredSkillsForEcosystem() ([]types.Skill, error) {
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
		skill, err := parseSkill(p)
		if err != nil {
			continue
		}
		skills = append(skills, *skill)
	}
	return skills, nil
}

// parseSkill parses a single skill file. Uses parser.Parse internally.
func parseSkill(path string) (*types.Skill, error) {
	return parser.Parse(path)
}

// getBaseCommandSkills returns the 5 base command pseudo-skills.
func getBaseCommandSkills() []types.Skill {
	return []types.Skill{
		{Name: "skill", Metadata: types.Metadata{Description: "Entry point for skill management", AutoInvoke: true}},
		{Name: "find", Metadata: types.Metadata{Description: "Search and list existing skills", AutoInvoke: true}},
		{Name: "create", Metadata: types.Metadata{Description: "Create a new agent skill from a prompt", AutoInvoke: true}},
		{Name: "sync", Metadata: types.Metadata{Description: "Synchronize skills and update AGENTS.md/OPENCODE.md", AutoInvoke: true}},
		{Name: "fullskills", Metadata: types.Metadata{Description: "Complete skill workflow", AutoInvoke: true}},
	}
}

// RegisterSkillManagerAgent creates .opencode/agents/skill-manager.md.
// It delegates to opencode.RegenerateAgent after discovering mirrored skills.
func RegisterSkillManagerAgent() error {
	// Sync skills first (non-fatal)
	opencode.SyncSkills(".", opencode.Options{})

	// Parse mirrored skills
	skills, err := parseMirroredSkillsForEcosystem()
	if err != nil {
		return err
	}

	// Add base command skills
	baseSkills := getBaseCommandSkills()
	skills = append(baseSkills, skills...)

	_, err = opencode.RegenerateAgent(".", skills, false)
	return err
}
