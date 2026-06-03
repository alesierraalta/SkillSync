package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"testing"
)

func TestInstallerModel_AutoskillsToggle(t *testing.T) {
	m := NewInstallerModel(nil, ".")

	// Initial state: SmartScan should be false
	if m.Autoskills {
		t.Errorf("expected Autoskills to be false initially")
	}

	// Move cursor to SmartScan (index 8, since 0-4 Providers, 5-7 Skills)
	// We'll update the model in the next step to add this.
	// For TDD, we assume the cursor index 8 will be SmartScan.
	m.Cursor = 8

	// Toggle
	msg := tea.KeyMsg{Type: tea.KeySpace}
	newModel, _ := m.Update(msg)
	m = newModel.(InstallerModel)

	if !m.Autoskills {
		t.Errorf("expected Autoskills to be true after toggle")
	}

	// Toggle back
	newModel, _ = m.Update(msg)
	m = newModel.(InstallerModel)

	if m.Autoskills {
		t.Errorf("expected Autoskills to be false after toggle back")
	}
}

func TestInstallerModel_LicenseDisclosure(t *testing.T) {
	m := NewInstallerModel(nil, ".")
	m.Autoskills = true

	// Press Execute Install (cursor will be at 10 + storageOffset)
	// We'll simulate the state transition.

	m.Cursor = 9 + len(m.AllStored) + 1 // [ Execute Install ]

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(msg)
	m = newModel.(InstallerModel)

	if m.Screen != ScreenLicenseDisclosure {
		t.Errorf("expected screen to be ScreenLicenseDisclosure, got %v", m.Screen)
	}

	// Confirm license
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Error("expected syncRequestMsg command after license confirmation")
	}
}
