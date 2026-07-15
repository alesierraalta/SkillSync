package vault

import (
	"os"
	"path/filepath"
	"testing"

	"skillsync/tui/internal/storage"
)

// setupStorageWithSkills creates a storage directory with the given skill
// directories, each containing SKILL.md and METADATA.json.
func setupStorageWithSkills(t *testing.T, root string, names []string) {
	t.Helper()
	for _, name := range names {
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# "+name), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "METADATA.json"), []byte(`{"skill_name":"`+name+`"}`), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestList_ReturnsSkillsFromStorage(t *testing.T) {
	tmpDir := t.TempDir()
	setupStorageWithSkills(t, tmpDir, []string{"foo", "bar"})

	s := &Service{store: &storage.Service{RootPath: tmpDir}}
	skills, err := s.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}

	names := make(map[string]bool)
	for _, sk := range skills {
		names[sk.Metadata.SkillName] = true
	}
	if !names["foo"] {
		t.Error("expected 'foo' in list")
	}
	if !names["bar"] {
		t.Error("expected 'bar' in list")
	}
}

func TestList_EmptyStorageReturnsEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	s := &Service{store: &storage.Service{RootPath: tmpDir}}
	skills, err := s.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(skills))
	}
}

func TestRemove_BlocksCoreSkills(t *testing.T) {
	tmpDir := t.TempDir()

	s := &Service{store: &storage.Service{RootPath: tmpDir}}
	coreNames := []string{"skill-creator", "skill-sync", "find-skills"}

	for _, name := range coreNames {
		t.Run(name, func(t *testing.T) {
			err := s.Remove(name)
			if err == nil {
				t.Fatal("expected error for core skill removal")
			}
		})
	}
}

func TestRemove_DeletesNonCoreSkill(t *testing.T) {
	tmpDir := t.TempDir()
	setupStorageWithSkills(t, tmpDir, []string{"my-skill"})

	s := &Service{store: &storage.Service{RootPath: tmpDir}}
	err := s.Remove("my-skill")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify directory is gone
	if _, err := os.Stat(filepath.Join(tmpDir, "my-skill")); !os.IsNotExist(err) {
		t.Error("expected skill directory to be removed")
	}
}

func TestRemove_NonExistentReturnsError(t *testing.T) {
	tmpDir := t.TempDir()

	s := &Service{store: &storage.Service{RootPath: tmpDir}}
	err := s.Remove("nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent skill")
	}
}

func TestRootPath_ReturnsStorageRoot(t *testing.T) {
	tmpDir := t.TempDir()

	s := &Service{store: &storage.Service{RootPath: tmpDir}}
	if got := s.RootPath(); got != tmpDir {
		t.Errorf("expected RootPath %q, got %q", tmpDir, got)
	}
}

func TestNewService_CreatesStorageService(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("SKILLSYNC_HOME", tmpDir)

	s := NewService()
	if s == nil {
		t.Fatal("expected non-nil Service")
	}
	if s.store == nil {
		t.Fatal("expected non-nil underlying store")
	}
	expected := filepath.Join(tmpDir, "storage")
	if s.store.RootPath != expected {
		t.Errorf("expected store.RootPath %q, got %q", expected, s.store.RootPath)
	}
}

func TestSelection_IsTransientMap(t *testing.T) {
	sel := make(Selection)
	if len(sel) != 0 {
		t.Error("expected empty selection")
	}
	sel["foo"] = true
	sel["bar"] = true
	if !sel["foo"] {
		t.Error("expected foo to be selected")
	}
	if !sel["bar"] {
		t.Error("expected bar to be selected")
	}
	if sel["baz"] {
		t.Error("expected baz to NOT be selected")
	}
	// Remove from selection
	delete(sel, "foo")
	if sel["foo"] {
		t.Error("expected foo to be removed from selection")
	}
	if len(sel) != 1 {
		t.Errorf("expected 1 item in selection, got %d", len(sel))
	}
}

func TestList_ReturnsStoredSkillIDs(t *testing.T) {
	tmpDir := t.TempDir()
	setupStorageWithSkills(t, tmpDir, []string{"alpha", "beta"})

	s := &Service{store: &storage.Service{RootPath: tmpDir}}
	skills, err := s.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	ids := make(map[string]bool)
	for _, sk := range skills {
		ids[sk.ID] = true
	}
	if !ids["alpha"] {
		t.Error("expected ID 'alpha' in list")
	}
	if !ids["beta"] {
		t.Error("expected ID 'beta' in list")
	}
}
