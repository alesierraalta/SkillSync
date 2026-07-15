package ui

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"skillsync/tui/internal/storage"
)

func storageModelWith(backend AppService, names ...string) Model {
	m := NewModel(backend)
	m.Screen = ScreenStorage
	items := make([]list.Item, 0, len(names))
	for _, n := range names {
		items = append(items, storageItem{stored: storage.StoredSkill{
			Metadata: storage.StoredMetadata{SkillName: n},
		}})
	}
	m.storageList.SetItems(items)
	m.storageList.Select(0)
	return m
}

func keyRune(r rune) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}} }

func TestStorageSpaceTogglesSelection(t *testing.T) {
	m := storageModelWith(&MockAppService{}, "foo", "bar")

	nm, _ := m.handleStorageKeys(keyRune(' '))
	m = nm.(Model)
	if !m.vaultSelected["foo"] {
		t.Fatal("space should select the highlighted vault skill")
	}

	nm, _ = m.handleStorageKeys(keyRune(' '))
	if nm.(Model).vaultSelected["foo"] {
		t.Fatal("space again should deselect")
	}
}

func TestStorageExportSelected(t *testing.T) {
	var gotNames []string
	mock := &MockAppService{ExportBundleFunc: func(names []string, dest string) (string, error) {
		gotNames = append([]string{}, names...)
		return dest, nil
	}}
	m := storageModelWith(mock, "foo", "bar")
	m.vaultSelected["foo"] = true

	_, cmd := m.handleStorageKeys(keyRune('e'))
	if cmd == nil {
		t.Fatal("'e' with a selection should return an export cmd")
	}
	cmd() // execute the export

	if len(gotNames) != 1 || gotNames[0] != "foo" {
		t.Fatalf("ExportBundle called with %v, want [foo]", gotNames)
	}
}

func TestStorageExportNothingSelected(t *testing.T) {
	m := storageModelWith(&MockAppService{}, "foo")

	nm, cmd := m.handleStorageKeys(keyRune('e'))
	if cmd != nil {
		t.Fatal("'e' with no selection should not export")
	}
	if nm.(Model).StatusMsg == "" {
		t.Fatal("'e' with no selection should set a status message")
	}
}
