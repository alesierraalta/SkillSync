package ui

import (
	"os"
	"path/filepath"
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
					dir := filepath.Join(".agents", "skills", sk)
					_ = os.MkdirAll(dir, 0755)
					skillFile := filepath.Join(dir, "SKILL.md")
					if _, err := os.Stat(skillFile); os.IsNotExist(err) {
						desc := coreSkills[sk]
						if desc == "" {
							desc = "Core skill for SkillSync."
						}
						_ = os.WriteFile(skillFile, []byte("# "+sk+"\n\n"+desc), 0644)
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
		skillsDir := filepath.Join(".agents", "skills")
		dirs := []string{
			filepath.Join(skillsDir, "skill-creator"),
			filepath.Join(skillsDir, "skill-sync"),
			filepath.Join(skillsDir, "find-skills"),
		}

		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return ecosystemMsg{err: err}
			}

			skillFile := filepath.Join(dir, "SKILL.md")
			if _, err := os.Stat(skillFile); os.IsNotExist(err) {
				sk := filepath.Base(dir)
				desc := coreSkills[sk]
				if desc == "" {
					desc = "Core skill for SkillSync."
				}
				content := []byte("# Skill " + sk + "\n\n" + desc)
				if err := os.WriteFile(skillFile, content, 0644); err != nil {
					return ecosystemMsg{err: err}
				}
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
