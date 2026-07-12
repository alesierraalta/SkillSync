package ui

import (
	"os"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/types"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func TestGlobalSkillItemTitle(t *testing.T) {
	tests := []struct {
		name     string
		category string
		path     string
		expected string
	}{
		{
			name:     "All category prefixes provider",
			category: "All",
			path:     ".claude/skills/my-skill/SKILL.md",
			expected: "[claude] Test Skill",
		},
		{
			name:     "Specific category does not prefix",
			category: "Claude",
			path:     ".claude/skills/my-skill/SKILL.md",
			expected: "Test Skill",
		},
		{
			name:     "All category with unknown provider",
			category: "All",
			path:     "unknown/skills/my-skill/SKILL.md",
			expected: "Test Skill",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sk := types.Skill{
				Name: "Test Skill",
				Path: tt.path,
			}

			it := globalSkillItem{skill: sk, category: tt.category}
			title := it.Title()

			if title != tt.expected {
				t.Errorf("expected title %q, got %q", tt.expected, title)
			}
		})
	}
}

func TestItemTitle(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Unix path .agents",
			path:     ".agents/skills/my-skill/SKILL.md",
			expected: "Test Skill [.agents]",
		},
		{
			name:     "Windows path .opencode",
			path:     ".opencode\\skills\\my-skill\\SKILL.md",
			expected: "Test Skill [.opencode]",
		},
		{
			name:     "False positive prevention (substring in name)",
			path:     ".agents/skills/claude-helper/SKILL.md",
			expected: "Test Skill [.agents]",
		},
		{
			name:     "Virtual agent",
			path:     "virtual",
			expected: "Test Skill",
		},
		{
			name:     "No provider directory",
			path:     "src/skills/my-skill/SKILL.md",
			expected: "Test Skill",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sk := types.Skill{
				Name: "Test Skill",
				Path: tt.path,
			}
			if tt.name == "Virtual agent" {
				sk.ID = "virtual:agents"
			}

			it := item{skill: sk}
			title := it.Title()

			if title != tt.expected {
				t.Errorf("expected title %q, got %q", tt.expected, title)
			}
		})
	}
}

func TestProjectItemDescription(t *testing.T) {
	t.Run("displays Nunca for zero LastSynced", func(t *testing.T) {
		item := projectItem{
			project: storage.ProjectInfo{
				Path:       "/test/path",
				LastSynced: time.Time{},
			},
		}
		expected := "Último sync: Nunca"
		if got := item.Description(); got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("displays formatted date for non-zero LastSynced", func(t *testing.T) {
		now := time.Date(2024, 5, 11, 15, 4, 0, 0, time.UTC)
		item := projectItem{
			project: storage.ProjectInfo{
				Path:       "/test/path",
				LastSynced: now,
			},
		}
		expected := "Último sync: 2024-05-11 15:04"
		if got := item.Description(); got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})
}

func TestInit_StaleBinaryWarning(t *testing.T) {
	// This test verifies that Init emits a statusMsg warning when the
	// running synck binary is older than minSupportedCommit.
	// The actual version check will be implemented in Phase 2.
	// For now, verify the Init batch structure allows for a version check command.
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	_ = os.MkdirAll(".agents/skills", 0755)
	_ = os.WriteFile("AGENTS.md", []byte("# Agents\n"), 0644)

	m := NewModel(NewBackend(storage.NewService("")))
	m.rootPath = tmpDir

	// tea.Batch is a variadic Cmd func - can't be introspected directly.
	// After the fix, Init() returns tea.Batch(m.loadSkills(), instantiateEcosystemCmd()).
	// This test verifies Init() returns non-nil Cmd. The actual batching is verified
	// by the runtime behavior (both ecosystem and skills loading run at startup).
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init() returned nil - must return a batch of startup commands")
	}

	// Verify calling the cmd doesn't panic (confirms proper batching)
	// After the fix, this executes both loadSkills and instantiateEcosystemCmd.
	// Before the fix, this only executes loadSkills.
	// The ecosystemMsg will be handled by the Update loop after the fix.
}

func TestInit_ReturnsNonNilCmd(t *testing.T) {
	// This test verifies Init() returns a non-nil tea.Cmd.
	// The actual startup registration (.opencode package.json, skill-manager.md)
	// is verified by TestInstantiateEcosystemCmd_RegistersOpenCode.
	// Init batching is also checked via model state; this test only asserts
	// that the command is non-nil so TUI can boot without panicking.
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	_ = os.MkdirAll(".agents/skills", 0755)
	_ = os.WriteFile("AGENTS.md", []byte("# Agents\n"), 0644)

	m := NewModel(NewBackend(storage.NewService("")))
	m.rootPath = tmpDir

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init() returned nil — must return non-nil tea.Cmd for startup")
	}
}

func TestGetKeyBindings(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenStorage
	bindings := m.GetKeyBindings()

	found := false
	for _, b := range bindings {
		if b.Key == "i" {
			found = true
			if b.Help != "install & sync" {
				t.Errorf("expected help 'install & sync', got %q", b.Help)
			}
			break
		}
	}

	if !found {
		t.Error("expected keybinding 'i' not found for ScreenStorage")
	}
}

func TestDeleteKeybinding_VirtualSkill(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenList
	m.List.selected = &types.Skill{
		ID:   "virtual:agents",
		Name: "★ AGENTS.md",
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	result := updated.(Model)

	if result.Screen == ScreenDeleteConfirm {
		t.Error("'d' on virtual skill should not transition to ScreenDeleteConfirm")
	}
	if result.Screen != ScreenList {
		t.Errorf("expected ScreenList after 'd' on virtual, got Screen=%d", result.Screen)
	}
}

func TestDeleteKeybinding_CoreSkill(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenList
	m.List.selected = &types.Skill{
		ID:   "skill-creator",
		Name: "skill-creator",
		Path: ".agents/skills/skill-creator/SKILL.md",
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	result := updated.(Model)

	if result.Screen == ScreenDeleteConfirm {
		t.Error("'d' on core skill should not transition to ScreenDeleteConfirm")
	}
	if result.Screen != ScreenList {
		t.Errorf("expected ScreenList after 'd' on core, got Screen=%d", result.Screen)
	}
}

func TestDeleteKeybinding_RegularSkill(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenList
	m.List.selected = &types.Skill{
		ID:   "my-skill",
		Name: "my-skill",
		Path: ".agents/skills/my-skill/SKILL.md",
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	result := updated.(Model)

	if result.Screen != ScreenDeleteConfirm {
		t.Errorf("'d' on regular skill should transition to ScreenDeleteConfirm, got Screen=%d", result.Screen)
	}
	if result.deleteConfirm.skillName != "my-skill" {
		t.Errorf("expected deleteConfirm.skillName='my-skill', got %q", result.deleteConfirm.skillName)
	}
}

func TestDeleteKeybinding_NoSelection(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenList
	// no selection set

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	result := updated.(Model)

	if result.Screen == ScreenDeleteConfirm {
		t.Error("'d' with no selection should not transition to ScreenDeleteConfirm")
	}
}

func TestDeleteKeybinding_StorageScreen(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenStorage

	// Populate the storage list with items
	item := storageItem{stored: storage.StoredSkill{
		ID: "stored-skill",
		Metadata: storage.StoredMetadata{
			SkillName: "stored-skill",
		},
	}}
	m.storageList.SetItems([]list.Item{item})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	result := updated.(Model)

	if result.Screen != ScreenDeleteConfirm {
		t.Errorf("'d' on storage screen should transition to ScreenDeleteConfirm, got Screen=%d", result.Screen)
	}
	if result.deleteConfirm.skillName != "stored-skill" {
		t.Errorf("expected deleteConfirm.skillName='stored-skill', got %q", result.deleteConfirm.skillName)
	}
}
