package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"skillsync/tui/internal/storage"
	"strings"
	"testing"
)

func TestInstallerModel_StoredSkillsToggleTDD(t *testing.T) {
	m := NewInstallerModel(nil, ".")
	m.AllStored = []storage.StoredSkill{
		{Metadata: storage.StoredMetadata{SkillName: "skill-1"}},
		{Metadata: storage.StoredMetadata{SkillName: "skill-2"}},
	}
	m.StoredSkills = nil

	// Check OptionsView handling (around line 82), if m.StoredSkills is nil, initialize it with size len(m.AllStored).
	// If len(m.StoredSkills) <= idx, dynamically expand it before toggling StoredSkills[idx].
	m.Cursor = 9 // This should toggle index 0 of StoredSkills
	msg := tea.KeyMsg{Type: tea.KeySpace}
	newModel, _ := m.Update(msg)
	m = newModel.(InstallerModel)

	if m.StoredSkills == nil {
		t.Fatal("StoredSkills was not initialized")
	}
	if len(m.StoredSkills) != 2 {
		t.Fatalf("Expected StoredSkills to be initialized to len 2, got %d", len(m.StoredSkills))
	}
	if !m.StoredSkills[0] {
		t.Error("Expected StoredSkills[0] to be toggled to true")
	}

	// Now test out of bounds toggling (dynamically expand it)
	m.Cursor = 10 // toggles index 1
	newModel, _ = m.Update(msg)
	m = newModel.(InstallerModel)
	if !m.StoredSkills[1] {
		t.Error("Expected StoredSkills[1] to be toggled to true")
	}

	// Shrink StoredSkills to test dynamic expansion when len(m.StoredSkills) <= idx
	m.StoredSkills = m.StoredSkills[:1] // len = 1
	m.Cursor = 10 // idx = 1, out of bounds. Should expand
	newModel, _ = m.Update(msg)
	m = newModel.(InstallerModel)
	if len(m.StoredSkills) <= 1 {
		t.Fatalf("Expected StoredSkills to be expanded, got len %d", len(m.StoredSkills))
	}
	if !m.StoredSkills[1] {
		t.Error("Expected StoredSkills[1] to be toggled to true after expansion")
	}
}

func TestInstallerModel_PreviewViewTDD(t *testing.T) {
	m := NewInstallerModel(nil, ".")
	m.Skills = []bool{true, true, true}
	preview := m.PreviewView()
	if strings.Contains(preview, ".agent/skills/") {
		t.Error("PreviewView still contains .agent/skills/")
	}
	if !strings.Contains(preview, ".agents/skills/") {
		t.Error("PreviewView should contain .agents/skills/")
	}
}

func TestInstallerModel_LicenseDisclosureKeysTDD(t *testing.T) {
	m := NewInstallerModel(nil, ".")
	m.Screen = ScreenLicenseDisclosure
	m.Cursor = 0 // points to Providers[0]

	// Pressing space/enter on ScreenLicenseDisclosure should do nothing (not toggle Providers[0])
	origProvider := m.Providers[0]
	msgSpace := tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	newModel, cmd := m.Update(msgSpace)
	m = newModel.(InstallerModel)
	if m.Providers[0] != origProvider {
		t.Error("Expected space key on ScreenLicenseDisclosure to not toggle options")
	}
	if cmd != nil {
		t.Error("Expected space key on ScreenLicenseDisclosure to return nil cmd")
	}

	// Pressing y/Y on ScreenLicenseDisclosure should trigger sync
	m.Screen = ScreenLicenseDisclosure
	msgY := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	newModel, cmd = m.Update(msgY)
	m = newModel.(InstallerModel)
	if cmd == nil {
		t.Error("Expected 'y' key on ScreenLicenseDisclosure to return sync cmd")
	}

	// Pressing n/N on ScreenLicenseDisclosure should return to ScreenInstaller
	m.Screen = ScreenLicenseDisclosure
	msgN := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, cmd = m.Update(msgN)
	m = newModel.(InstallerModel)
	if m.Screen != ScreenInstaller {
		t.Errorf("Expected 'n' key on ScreenLicenseDisclosure to transition to ScreenInstaller, got %d", m.Screen)
	}
	if cmd != nil {
		t.Error("Expected 'n' key on ScreenLicenseDisclosure to return nil cmd")
	}
}

func TestInstallerModel_LicenseDisclosureNavigationTDD(t *testing.T) {
	m := NewInstallerModel(nil, ".")
	m.Screen = ScreenLicenseDisclosure
	m.Cursor = 5

	// Pressing down key should not change cursor
	msgDown := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := m.Update(msgDown)
	m = newModel.(InstallerModel)
	if m.Cursor != 5 {
		t.Errorf("Expected Cursor to remain 5 on ScreenLicenseDisclosure, got %d", m.Cursor)
	}

	// Pressing up key should not change cursor
	msgUp := tea.KeyMsg{Type: tea.KeyUp}
	newModel, _ = m.Update(msgUp)
	m = newModel.(InstallerModel)
	if m.Cursor != 5 {
		t.Errorf("Expected Cursor to remain 5 on ScreenLicenseDisclosure, got %d", m.Cursor)
	}
}

func TestInstallerModel_DynamicCardStyleTDD(t *testing.T) {
	m := NewInstallerModel(nil, ".")
	
	// Test normal border for small heights
	m.Height = 20
	styleSmall := m.getCardStyle()
	borderSmall := styleSmall.GetBorderStyle()
	if borderSmall.Top != lipgloss.NormalBorder().Top {
		t.Error("Expected normal border for height < 24")
	}

	// Test rounded border for large heights
	m.Height = 30
	styleLarge := m.getCardStyle()
	borderLarge := styleLarge.GetBorderStyle()
	if borderLarge.Top != lipgloss.RoundedBorder().Top {
		t.Error("Expected rounded border for height >= 24")
	}
}


