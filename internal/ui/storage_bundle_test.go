package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"skillsync/tui/internal/bundle"
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

func TestBundleImportEnterTriggersImport(t *testing.T) {
	var gotPath, gotRoot string
	mock := &MockAppService{ImportBundleFunc: func(bundlePath, projectRoot string) ([]bundle.ImportResult, error) {
		gotPath = bundlePath
		gotRoot = projectRoot
		return []bundle.ImportResult{{Skill: "foo", Status: "installed"}}, nil
	}}
	m := NewModel(mock)
	m.Screen = ScreenBundleImport
	m.rootPath = "/proj"
	m.bundleImportIn.SetValue("/tmp/x.skillsync")

	_, cmd := m.handleBundleImportKeys(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should trigger an import cmd")
	}
	if _, ok := cmd().(bundleImportedMsg); !ok {
		t.Fatal("import cmd should return bundleImportedMsg")
	}
	if gotPath != "/tmp/x.skillsync" || gotRoot != "/proj" {
		t.Fatalf("ImportBundle called path=%q root=%q", gotPath, gotRoot)
	}
}

func TestBundleImportEscReturnsToStorage(t *testing.T) {
	m := NewModel(&MockAppService{})
	m.Screen = ScreenBundleImport
	nm, _ := m.handleBundleImportKeys(tea.KeyMsg{Type: tea.KeyEsc})
	if nm.(Model).Screen != ScreenStorage {
		t.Fatalf("esc should return to ScreenStorage, got %v", nm.(Model).Screen)
	}
}

func TestStorageViewShowsSelectionMarkers(t *testing.T) {
	m := storageModelWith(&MockAppService{}, "foo", "bar")
	m.storedSkills = []storage.StoredSkill{
		{Metadata: storage.StoredMetadata{SkillName: "foo"}},
		{Metadata: storage.StoredMetadata{SkillName: "bar"}},
	}
	m.selectMode = true
	m.vaultSelected["foo"] = true
	m.Width, m.Height = 80, 24

	out := m.storageView()
	if !strings.Contains(out, "[x]") {
		t.Error("select mode should render [x] for selected skills")
	}
	if !strings.Contains(out, "[ ]") {
		t.Error("select mode should render [ ] for unselected skills")
	}
}

func TestBundleImportViewRenders(t *testing.T) {
	m := NewModel(&MockAppService{})
	m.Screen = ScreenBundleImport
	m.Width, m.Height = 80, 24
	if strings.TrimSpace(m.bundleImportView()) == "" {
		t.Error("bundle import view should render content")
	}
}

func TestSummarizeImportReportsPartialFailure(t *testing.T) {
	got := summarizeImport([]importResultLine{
		{skill: "a", status: "installed"},
		{skill: "b", status: "failed"},
		{skill: "c", status: "warning"},
	})
	if !strings.Contains(got, "Imported 1") {
		t.Errorf("want installed count 1, got %q", got)
	}
	if !strings.Contains(got, "1 failed") {
		t.Errorf("partial failure must be surfaced, got %q", got)
	}
	if !strings.Contains(got, "warning") {
		t.Errorf("warnings must be surfaced, got %q", got)
	}
}

func TestImportedMsgSurfacesFailureInStatus(t *testing.T) {
	m := NewModel(&MockAppService{})
	nm, _ := m.Update(bundleImportedMsg{results: []importResultLine{
		{skill: "a", status: "installed"},
		{skill: "b", status: "failed"},
	}})
	if !strings.Contains(nm.(Model).StatusMsg, "failed") {
		t.Errorf("import status should report the failure, got %q", nm.(Model).StatusMsg)
	}
}

func TestWindowBounds(t *testing.T) {
	if s, e := windowBounds(0, 3, 0); s != 0 || e != 3 {
		t.Errorf("maxRows<=0 should show all: got [%d,%d)", s, e)
	}
	if s, e := windowBounds(9, 10, 3); s != 7 || e != 10 {
		t.Errorf("cursor near end should keep it visible: got [%d,%d)", s, e)
	}
	if s, e := windowBounds(0, 10, 3); s != 0 || e != 3 {
		t.Errorf("cursor at start: got [%d,%d)", s, e)
	}
}
