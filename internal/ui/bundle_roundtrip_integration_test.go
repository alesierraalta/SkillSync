package ui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"skillsync/tui/internal/bundle"
	"skillsync/tui/internal/storage"
)

// TestBundleRoundTrip_ThroughUI exercises the real backend end-to-end via the
// UI handlers: select a vault skill, export it to a real .skillsync file, then
// import that bundle into a project and verify the skill files land on disk.
func TestBundleRoundTrip_ThroughUI(t *testing.T) {
	// no-op registry sync during import
	origSync := bundle.SyncAfterImport
	bundle.SyncAfterImport = func(string) error { return nil }
	t.Cleanup(func() { bundle.SyncAfterImport = origSync })

	vaultRoot := t.TempDir()
	writeVaultSkill(t, vaultRoot, "foo")
	backend := NewBackend(storage.NewService(vaultRoot))

	// --- Export via ScreenStorage ---
	m := storageModelWith(backend, "foo")
	m, _ = applyKey(t, m, m.handleStorageKeys, keyRune(' ')) // select foo
	if !m.vaultSelected["foo"] {
		t.Fatal("foo should be selected")
	}
	_, cmd := m.handleStorageKeys(keyRune('e'))
	if cmd == nil {
		t.Fatal("export should produce a cmd")
	}
	exp, ok := cmd().(bundleExportedMsg)
	if !ok || exp.err != nil {
		t.Fatalf("export failed: %+v", exp)
	}
	if _, err := os.Stat(exp.path); err != nil {
		t.Fatalf("exported bundle missing on disk: %v", err)
	}

	// --- Import into a fresh project via ScreenBundleImport ---
	projectRoot := t.TempDir()
	m.rootPath = projectRoot
	m.Screen = ScreenBundleImport
	m.bundleImportIn.SetValue(exp.path)
	_, cmd = m.handleBundleImportKeys(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("import should produce a cmd")
	}
	imp, ok := cmd().(bundleImportedMsg)
	if !ok || imp.err != nil {
		t.Fatalf("import failed: %+v", imp)
	}
	if len(imp.results) == 0 {
		t.Fatal("import returned no results")
	}

	// The skill must now exist somewhere under the project root.
	found := false
	_ = filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() && info.Name() == "SKILL.md" &&
			filepath.Base(filepath.Dir(path)) == "foo" {
			found = true
		}
		return nil
	})
	if !found {
		t.Fatalf("imported skill 'foo' not found under project root %s", projectRoot)
	}
}

// applyKey runs a key handler and returns the resulting Model.
func applyKey(t *testing.T, m Model, h func(tea.KeyMsg) (tea.Model, tea.Cmd), msg tea.KeyMsg) (Model, tea.Cmd) {
	t.Helper()
	nm, cmd := h(msg)
	return nm.(Model), cmd
}
