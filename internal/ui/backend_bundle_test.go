package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"skillsync/tui/internal/bundle"
	"skillsync/tui/internal/storage"
)

// writeVaultSkill creates a vault skill (SKILL.md + METADATA.json) under root.
func writeVaultSkill(t *testing.T, root, name string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("name: "+name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	meta := storage.StoredMetadata{SkillName: name, Description: "test " + name, SavedAt: time.Now()}
	data, _ := json.Marshal(meta)
	if err := os.WriteFile(filepath.Join(dir, "METADATA.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBackendExportBundle(t *testing.T) {
	root := t.TempDir()
	writeVaultSkill(t, root, "foo")
	b := NewBackend(storage.NewService(root))

	dest := filepath.Join(t.TempDir(), "out.skillsync")
	path, err := b.ExportBundle([]string{"foo"}, dest)
	if err != nil {
		t.Fatalf("ExportBundle: %v", err)
	}
	if path != dest {
		t.Errorf("path = %q, want %q", path, dest)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Errorf("bundle not created: %v", err)
	}
}

func TestBackendExportBundle_UnknownSkill(t *testing.T) {
	b := NewBackend(storage.NewService(t.TempDir()))
	if _, err := b.ExportBundle([]string{"nope"}, filepath.Join(t.TempDir(), "x.skillsync")); err == nil {
		t.Fatal("expected error for unknown skill")
	}
}

func TestBackendImportBundle_RoundTrip(t *testing.T) {
	orig := bundle.SyncAfterImport
	bundle.SyncAfterImport = func(string) error { return nil }
	t.Cleanup(func() { bundle.SyncAfterImport = orig })

	root := t.TempDir()
	writeVaultSkill(t, root, "foo")
	b := NewBackend(storage.NewService(root))

	dest := filepath.Join(t.TempDir(), "out.skillsync")
	if _, err := b.ExportBundle([]string{"foo"}, dest); err != nil {
		t.Fatalf("export: %v", err)
	}

	results, err := b.ImportBundle(dest, t.TempDir())
	if err != nil {
		t.Fatalf("ImportBundle: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one import result")
	}
}

func TestBackendImportBundle_BadPath(t *testing.T) {
	b := NewBackend(storage.NewService(t.TempDir()))
	if _, err := b.ImportBundle(filepath.Join(t.TempDir(), "missing.skillsync"), t.TempDir()); err == nil {
		t.Fatal("expected error for missing bundle")
	}
}
