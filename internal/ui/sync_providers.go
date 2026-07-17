package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/syncengine"
)

// syncProvider pairs a display label with its on-disk provider directory.
type syncProvider struct {
	Label string
	Dir   string
}

// syncProviderList is the canonical set of providers the sync screen can target.
// OpenCode is first and gets the full tools/agents/commands regeneration; the
// rest receive a plain skill-tree mirror (see providersync.Mirror).
var syncProviderList = []syncProvider{
	{"OpenCode", ".opencode"},
	{"Claude", ".claude"},
	{"Gemini", ".gemini"},
	{"Cursor", ".cursor"},
	{"Copilot", ".copilot"},
	{"Qwen", ".qwen"},
}

// openSyncProviders shows the provider selection screen, pre-checking OpenCode
// plus any provider whose directory already exists in the project.
func (m Model) openSyncProviders() (tea.Model, tea.Cmd) {
	sel := make([]bool, len(syncProviderList))
	for i, p := range syncProviderList {
		if p.Dir == ".opencode" {
			sel[i] = true
			continue
		}
		if info, err := os.Stat(filepath.Join(m.rootPath, p.Dir)); err == nil && info.IsDir() {
			sel[i] = true
		}
	}
	m.syncProviderSel = sel
	m.syncProviderCursor = 0
	// Entering the provider screen begins a new sync flow; reset stale
	// result state so the syncing view never shows the previous run.
	m.SyncFailed = false
	m.SyncFinished = false
	m.PrevScreen = m.Screen
	m.Screen = ScreenSyncProviders
	return m, nil
}

func (m Model) handleSyncProvidersKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.Screen = m.PrevScreen
		return m, nil
	case "j", "down":
		if m.syncProviderCursor < len(syncProviderList)-1 {
			m.syncProviderCursor++
		}
	case "k", "up":
		if m.syncProviderCursor > 0 {
			m.syncProviderCursor--
		}
	case " ", "x":
		if m.syncProviderCursor >= 0 && m.syncProviderCursor < len(m.syncProviderSel) {
			m.syncProviderSel[m.syncProviderCursor] = !m.syncProviderSel[m.syncProviderCursor]
		}
	case "enter":
		// Require at least one provider selected.
		any := false
		for _, on := range m.syncProviderSel {
			if on {
				any = true
				break
			}
		}
		if !any {
			m.StatusMsg = "Select at least one provider (space to toggle)"
			return m, nil
		}
		cmd := m.syncToSelectedProvidersCmd()
		m.pendingStored = nil // consumed by the command closure
		m.Screen = ScreenSyncing
		m.SyncFailed = false
		m.SyncFinished = false
		return m, cmd
	}
	return m, nil
}

// selectedProviderDirs returns the provider directories currently toggled on.
func (m Model) selectedProviderDirs() []string {
	var dirs []string
	for i, on := range m.syncProviderSel {
		if on && i < len(syncProviderList) {
			dirs = append(dirs, syncProviderList[i].Dir)
		}
	}
	return dirs
}

func (m Model) syncToSelectedProvidersCmd() tea.Cmd {
	root := m.rootPath
	providers := m.selectedProviderDirs()
	pending := m.pendingStored
	return func() tea.Msg {
		_ = m.backend.InstallCoreSkill("skill-sync") // non-fatal
		_ = m.backend.EnsureAgentsMD(root)           // non-fatal

		// Storage "i" flow: materialize the vault skill into .agents/skills
		// before mirroring, so the provider mirror picks it up.
		if pending != nil {
			if err := m.installStoredToProject(*pending, root); err != nil {
				return syncReportMsg{report: &runner.SyncReport{}, err: err}
			}
		}

		_ = m.backend.SyncToProviders(root, providers)

		report, err := m.backend.Sync(root, syncengine.SyncOptions{})
		return syncReportMsg{report: report, err: err}
	}
}

// installStoredToProject writes a vault skill (SKILL.md plus references/assets)
// into <root>/.agents/skills/<name>. Shared by the storage "i" install path.
func (m Model) installStoredToProject(stored storage.StoredSkill, root string) error {
	content, err := m.backend.LoadFromStorage(stored.ID)
	if err != nil {
		return fmt.Errorf("load skill from storage: %w", err)
	}
	skill, err := m.backend.ParseSkillContent(content)
	if err != nil {
		return fmt.Errorf("malformed skill: %w", err)
	}
	name := skill.Name
	if name == "" {
		name = stored.Metadata.SkillName
	}
	skillDir := filepath.Join(root, ".agents", "skills", name)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("create skill dir: %w", err)
	}
	if err := m.backend.CopyStorageExtras(stored.ID, skillDir); err != nil {
		return fmt.Errorf("copy skill files: %w", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644); err != nil {
		return fmt.Errorf("write SKILL.md: %w", err)
	}
	return nil
}

func (m Model) viewSyncProviders() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Sync — select target providers") + "\n\n")
	b.WriteString(hintStyle.Render("space: toggle · enter: sync · esc: cancel") + "\n\n")
	for i, p := range syncProviderList {
		cursor := "  "
		if i == m.syncProviderCursor {
			cursor = "> "
		}
		marker := "[ ]"
		if i < len(m.syncProviderSel) && m.syncProviderSel[i] {
			marker = "[x]"
		}
		b.WriteString(cursor + marker + " " + p.Label + "\n")
	}
	return docStyle.Render(b.String())
}
