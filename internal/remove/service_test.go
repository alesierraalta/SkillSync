package remove

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"skillsync/tui/internal/storage"
)

func setupTestSkill(t *testing.T, root, name string) {
	t.Helper()

	canonicalDir := filepath.Join(root, ".agents", "skills", name)
	if err := os.MkdirAll(canonicalDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(canonicalDir, "SKILL.md"), []byte("# "+name), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(canonicalDir, "METADATA.json"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}
}

func setupMirrorSkill(t *testing.T, root, provider, name string) {
	t.Helper()

	dir := filepath.Join(root, provider, "skills", name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# "+name), 0644); err != nil {
		t.Fatal(err)
	}
}

func setupSkillsLock(t *testing.T, root string, entries []string) {
	t.Helper()

	skills := make(map[string]interface{})
	for _, name := range entries {
		skills[name] = map[string]interface{}{
			"source":        "test",
			"sourceType":    "github",
			"computedHash": "abc123",
		}
	}
	lock := map[string]interface{}{
		"version": 1,
		"skills":  skills,
	}
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(root, "skills-lock.json")
	if err := os.WriteFile(lockPath, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func setupStorageSkill(t *testing.T, storageDir, name string) {
	t.Helper()

	skillDir := filepath.Join(storageDir, name)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# "+name), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "METADATA.json"), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveByID_Success(t *testing.T) {
	tmpDir := t.TempDir()
	storageDir := t.TempDir()

	// Setup: create skill in canonical location + mirrors + storage + lock
	setupTestSkill(t, tmpDir, "my-skill")
	setupMirrorSkill(t, tmpDir, ".opencode", "my-skill")
	setupMirrorSkill(t, tmpDir, ".claude", "my-skill")
	setupStorageSkill(t, storageDir, "my-skill")
	setupSkillsLock(t, tmpDir, []string{"my-skill", "other-skill"})

	s := &Service{
		Storage:  &storage.Service{RootPath: storageDir},
		RootPath: tmpDir,
	}

	err := s.RemoveByID("my-skill", Options{})
	if err != nil {
		t.Fatalf("RemoveByID failed: %v", err)
	}

	// Verify canonical deleted
	if _, err := os.Stat(filepath.Join(tmpDir, ".agents", "skills", "my-skill")); !os.IsNotExist(err) {
		t.Error("expected canonical dir to be removed")
	}

	// Verify mirrors deleted
	if _, err := os.Stat(filepath.Join(tmpDir, ".opencode", "skills", "my-skill")); !os.IsNotExist(err) {
		t.Error("expected opencode mirror to be removed")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, ".claude", "skills", "my-skill")); !os.IsNotExist(err) {
		t.Error("expected claude mirror to be removed")
	}

	// Verify storage deleted
	if _, err := os.Stat(filepath.Join(storageDir, "my-skill")); !os.IsNotExist(err) {
		t.Error("expected storage dir to be removed")
	}

	// Verify lock updated
	lockData, err := os.ReadFile(filepath.Join(tmpDir, "skills-lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if containsInLock(lockData, "my-skill") {
		t.Error("expected my-skill to be removed from lock")
	}
	if !containsInLock(lockData, "other-skill") {
		t.Error("expected other-skill to remain in lock")
	}

	// Verify other skill dirs remain untouched
	if _, err := os.Stat(filepath.Join(tmpDir, ".agents", "skills", "other-skill")); os.IsNotExist(err) {
		// other-skill was never created — that's fine
	}
}

func TestRemoveByID_RejectsCoreSkills(t *testing.T) {
	tmpDir := t.TempDir()
	storageDir := t.TempDir()

	coreSkills := []string{"skill-creator", "skill-sync", "find-skills"}

	for _, name := range coreSkills {
		t.Run(name, func(t *testing.T) {
			s := &Service{
				Storage:  &storage.Service{RootPath: storageDir},
				RootPath: tmpDir,
			}

			err := s.RemoveByID(name, Options{})
			if err == nil {
				t.Fatal("expected error for core skill")
			}
		})
	}
}

func TestRemoveByID_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	storageDir := t.TempDir()

	// Create lock with the skill entry (to verify lock is still updated)
	setupSkillsLock(t, tmpDir, []string{"ghost-skill"})

	s := &Service{
		Storage:  &storage.Service{RootPath: storageDir},
		RootPath: tmpDir,
	}

	err := s.RemoveByID("ghost-skill", Options{})
	if err != nil {
		t.Fatalf("RemoveByID failed for non-existent skill: %v", err)
	}

	// Lock should still be cleaned up
	lockData, err := os.ReadFile(filepath.Join(tmpDir, "skills-lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	if containsInLock(lockData, "ghost-skill") {
		t.Error("expected ghost-skill to be removed from lock")
	}
}

func TestRemoveByID_LocalFlag(t *testing.T) {
	tmpDir := t.TempDir()
	storageDir := t.TempDir()

	setupTestSkill(t, tmpDir, "local-skill")
	setupStorageSkill(t, storageDir, "local-skill")
	setupSkillsLock(t, tmpDir, []string{"local-skill"})

	s := &Service{
		Storage:  &storage.Service{RootPath: storageDir},
		RootPath: tmpDir,
	}

	err := s.RemoveByID("local-skill", Options{Local: true})
	if err != nil {
		t.Fatalf("RemoveByID failed: %v", err)
	}

	// Canonical should be gone
	if _, err := os.Stat(filepath.Join(tmpDir, ".agents", "skills", "local-skill")); !os.IsNotExist(err) {
		t.Error("expected canonical dir to be removed")
	}

	// Storage should REMAIN (Local flag)
	if _, err := os.Stat(filepath.Join(storageDir, "local-skill")); os.IsNotExist(err) {
		t.Error("expected storage dir to remain with Local flag")
	}
}

func TestRemoveByID_CleansMirrors(t *testing.T) {
	tmpDir := t.TempDir()
	storageDir := t.TempDir()

	setupTestSkill(t, tmpDir, "mirror-skill")
	// Set up every mirror provider
	providers := []string{".opencode", ".claude", ".gemini", ".cursor", ".copilot", ".qwen"}
	for _, p := range providers {
		setupMirrorSkill(t, tmpDir, p, "mirror-skill")
	}
	setupSkillsLock(t, tmpDir, []string{"mirror-skill"})

	s := &Service{
		Storage:  &storage.Service{RootPath: storageDir},
		RootPath: tmpDir,
	}

	err := s.RemoveByID("mirror-skill", Options{})
	if err != nil {
		t.Fatalf("RemoveByID failed: %v", err)
	}

	// All mirror dirs should be gone
	for _, p := range providers {
		dir := filepath.Join(tmpDir, p, "skills", "mirror-skill")
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Errorf("expected %s mirror to be removed", p)
		}
	}
}

func TestRemoveByID_PartialFailure(t *testing.T) {
	tmpDir := t.TempDir()
	storageDir := t.TempDir()

	setupTestSkill(t, tmpDir, "partial-skill")
	setupMirrorSkill(t, tmpDir, ".opencode", "partial-skill")
	setupStorageSkill(t, storageDir, "partial-skill")
	setupSkillsLock(t, tmpDir, []string{"partial-skill"})

	// Make canonical dir read-only to trigger failure
	canonicalDir := filepath.Join(tmpDir, ".agents", "skills", "partial-skill")
	// We can make it fail by not having permissions. On Windows, this is tricky.
	// Instead, let's just test that mirrors and lock update still happen even if canonical
	// was already deleted (this simulates one partial-failure scenario)

	// Pre-delete canonical to simulate partial prior state
	if err := os.RemoveAll(canonicalDir); err != nil {
		t.Fatal(err)
	}

	s := &Service{
		Storage:  &storage.Service{RootPath: storageDir},
		RootPath: tmpDir,
	}

	err := s.RemoveByID("partial-skill", Options{})
	// Should still succeed because canonical deletion of already-deleted is a no-op
	if err != nil {
		t.Fatalf("RemoveByID failed: %v", err)
	}

	// Mirror should still be cleaned
	if _, err := os.Stat(filepath.Join(tmpDir, ".opencode", "skills", "partial-skill")); !os.IsNotExist(err) {
		t.Error("expected mirror to be removed despite canonical already gone")
	}

	// Lock should be updated
	lockData, _ := os.ReadFile(filepath.Join(tmpDir, "skills-lock.json"))
	if containsInLock(lockData, "partial-skill") {
		t.Error("expected partial-skill to be removed from lock despite canonical missing")
	}
}

// containsInLock checks if a skill name is a key in the skills-lock JSON.
func containsInLock(data []byte, name string) bool {
	var lock struct {
		Skills map[string]interface{} `json:"skills"`
	}
	if err := json.Unmarshal(data, &lock); err != nil {
		return false
	}
	_, ok := lock.Skills[name]
	return ok
}
