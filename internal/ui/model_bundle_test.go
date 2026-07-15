package ui

import (
	"skillsync/tui/internal/storage"
	"testing"
)

func TestScreenBundleImportDistinct(t *testing.T) {
	seen := map[Screen]bool{}
	for _, s := range []Screen{
		ScreenGlobalSkillsCats, ScreenGlobalSkillsList, ScreenBundleImport,
	} {
		if seen[s] {
			t.Fatalf("duplicate screen constant value %d", s)
		}
		seen[s] = true
	}
}

func TestNewModelInitsVaultSelection(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService(t.TempDir())))
	if m.vaultSelected == nil {
		t.Error("vaultSelected map should be initialized")
	}
	if m.selectMode {
		t.Error("selectMode should default to false")
	}
}
