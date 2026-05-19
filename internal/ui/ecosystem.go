package ui

import (
	"os"
	"path/filepath"
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
			for i, p := range m.Installer.Providers {
				if p {
					_ = os.MkdirAll(providers[i], 0755)
				}
			}
			return installerProgressMsg{percent: 0.3, task: "Instalando Core Skills..."}

		case 0.3:
			// 2. Install Core Skills
			skills := []string{"skill-creator", "skill-sync", "find-skills"}
			for i, enabled := range m.Installer.Skills {
				if enabled {
					sk := skills[i]
					if err := m.backend.InstallCoreSkill(sk); err != nil {
						return installerFinishedMsg{err: err}
					}
				}
			}
			return installerProgressMsg{percent: 0.5, task: "Instalando skills desde almacenamiento..."}

		case 0.5:
			// 2.5 Install Stored Skills
			for i, enabled := range m.Installer.StoredSkills {
				if enabled && i < len(m.storedSkills) {
					stored := m.storedSkills[i]
					content, err := m.backend.LoadFromStorage(stored.ID)
					if err != nil {
						// Log and skip instead of failing the whole ecosystem
						continue
					}

					// Verify skill before writing
					if _, err := m.backend.ParseSkillContent(content); err != nil {
						// Malformed YAML: bypass to keep ecosystem healthy
						continue
					}

					destDir := filepath.Join(".agents", "skills", stored.Metadata.SkillName)
					if err := os.MkdirAll(destDir, 0755); err != nil {
						continue
					}
					if err := os.WriteFile(filepath.Join(destDir, "SKILL.md"), []byte(content), 0644); err != nil {
						continue
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

			if m.Installer.Providers[0] {
				_ = os.WriteFile("CLAUDE.md", content, 0644)
			}
			if m.Installer.Providers[1] {
				_ = os.WriteFile("GEMINI.md", content, 0644)
			}
			if m.Installer.Providers[2] {
				_ = os.WriteFile("codex.md", content, 0644)
			}
			if m.Installer.Providers[3] {
				_ = os.MkdirAll(".github", 0755)
				_ = os.WriteFile(".github/copilot-instructions.md", content, 0644)
			}
			if m.Installer.Providers[4] {
				_ = os.WriteFile("OPENCODE.md", content, 0644)
				_ = m.backend.RegisterOpenCodeTools()
				_ = m.backend.RegisterSkillManagerAgent()
			}

			if m.Installer.Global {
				// Simular global aliases
				_ = os.WriteFile("GLOBAL_ALIASES.txt", []byte("alias k='skillsync'\nalias ks='skillsync sync'"), 0644)
			}

			if m.backend != nil {
				absRoot, _ := filepath.Abs(m.rootPath)
				_ = m.backend.RegisterProjectInitial(absRoot)
			}

			return installerProgressMsg{percent: 1.0, task: "Instalación completada"}

		case 1.0:
			return installerFinishedMsg{err: nil}
		}

		return nil
	}
}

func instantiateEcosystemCmd(backend AppService, rootPath string) tea.Cmd {
	return func() tea.Msg {
		skills := []string{"skill-creator", "skill-sync", "find-skills"}

		for _, sk := range skills {
			if err := backend.InstallCoreSkill(sk); err != nil {
				return ecosystemMsg{err: err}
			}
		}

		if err := backend.EnsureAgentsMD(); err != nil {
			return ecosystemMsg{err: err}
		}

		// Register OpenCode tools and agents when .opencode/ directory exists
		if _, err := os.Stat(".opencode"); err == nil {
			if err := backend.RegisterOpenCodeTools(); err != nil {
				return ecosystemMsg{err: err}
			}
			if err := backend.RegisterSkillManagerAgent(); err != nil {
				return ecosystemMsg{err: err}
			}
		}

		if backend != nil {
			_ = backend.RegisterProjectInitial(rootPath)
		}

		return ecosystemMsg{err: nil}
	}
}
