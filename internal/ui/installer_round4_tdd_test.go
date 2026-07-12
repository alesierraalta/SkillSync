package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"testing"
)

func TestInstallerModel_LicenseDisclosureConfirmationScreenResetTDD(t *testing.T) {
	m := NewInstallerModel(nil, ".")
	m.Screen = ScreenLicenseDisclosure

	msgY := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	newModel, cmd := m.Update(msgY)
	resModel := newModel.(InstallerModel)

	if resModel.Screen != ScreenInstaller {
		t.Errorf("Expected screen to reset to ScreenInstaller on 'y' press, got %v", resModel.Screen)
	}
	if cmd == nil {
		t.Fatal("Expected non-nil cmd returning syncRequestMsg")
	}
	msg := cmd()
	if _, ok := msg.(syncRequestMsg); !ok {
		t.Errorf("Expected msg of type syncRequestMsg, got %T", msg)
	}
}

func TestInstallerModel_ExecuteInstallSpaceKeyTDD(t *testing.T) {
	m := NewInstallerModel(nil, ".")
	m.Screen = ScreenInstaller
	storageOffset := len(m.AllStored)
	m.Cursor = 10 + storageOffset

	// Test "space" key
	msgSpace := tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	newModel, cmd := m.Update(msgSpace)
	m = newModel.(InstallerModel)
	if cmd == nil {
		t.Fatal("Expected space key to trigger Execute Install")
	}
	msg := cmd()
	if _, ok := msg.(syncRequestMsg); !ok {
		t.Errorf("Expected msg of type syncRequestMsg, got %T", msg)
	}

	// Test " " key
	msgSpaceRune := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	_, cmdRune := m.Update(msgSpaceRune)
	if cmdRune == nil {
		t.Fatal("Expected ' ' key to trigger Execute Install")
	}
}
