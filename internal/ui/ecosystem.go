package ui

import (
	"io/fs"
	"os"
	"path/filepath"
	"skillsync/tui/internal/coreskills"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

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
			// 3. Sync Configs (Dummy sync)
			agentsFile := "AGENTS.md"
			content, _ := os.ReadFile(agentsFile)
			if len(content) == 0 {
				content = []byte("# Agent Skills\n")
			}
			
			if m.installerProviders[0] { _ = os.WriteFile("CLAUDE.md", content, 0644) }
			if m.installerProviders[1] { _ = os.WriteFile("GEMINI.md", content, 0644) }
			if m.installerProviders[4] { _ = os.WriteFile("OPENCODE.md", content, 0644) }

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

		agentsFile := "AGENTS.md"
		if _, err := os.Stat(agentsFile); os.IsNotExist(err) {
			content := []byte("# Agent Skills\n\n| Skill | Description | Location |\n|---|---|---|\n")
			if err := os.WriteFile(agentsFile, content, 0644); err != nil {
				return ecosystemMsg{err: err}
			}
		}

		return ecosystemMsg{err: nil}
	}
}

func InstallCoreSkill(name string) error {
	destDir := filepath.Join(".agents", "skills", name)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	srcDir := "skills/" + name
	return fs.WalkDir(coreskills.EmbeddedSkills, srcDir, func(path string, d fs.DirEntry, err error) error {
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
	})
}
