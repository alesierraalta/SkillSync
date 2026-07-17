package syncengine

import (
	"os"
	"path/filepath"
	"testing"
)

func writeAsset(t *testing.T, root, skill, name string) string {
	t.Helper()
	dir := filepath.Join(root, ".agents", "skills", skill, "assets")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestCleanupPreservesSkillSyncAssets guards the regression where the legacy
// harness cleanup deleted skill-sync's own git-tracked assets (sync.sh,
// sync_test.sh, sync_test.ps1), which are legitimate embedded files.
func TestCleanupPreservesSkillSyncAssets(t *testing.T) {
	tmp := t.TempDir()
	kept := []string{
		writeAsset(t, tmp, "skill-sync", "sync.sh"),
		writeAsset(t, tmp, "skill-sync", "sync_test.sh"),
		writeAsset(t, tmp, "skill-sync", "sync_test.ps1"),
	}

	changes, err := cleanupLegacyHarnessArtifacts(tmp, false)
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	for _, p := range kept {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("skill-sync asset was deleted: %s", p)
		}
	}
	if len(changes) != 0 {
		t.Errorf("expected no deletions for skill-sync, got %d: %+v", len(changes), changes)
	}
}

// TestCleanupRemovesLegacyArtifactsInOtherSkills confirms the cleanup still
// purges the legacy harness copies from non-skill-sync skills.
func TestCleanupRemovesLegacyArtifactsInOtherSkills(t *testing.T) {
	tmp := t.TempDir()
	legacy := writeAsset(t, tmp, "some-other-skill", "sync.sh")

	changes, err := cleanupLegacyHarnessArtifacts(tmp, false)
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("legacy artifact in other skill was not removed: %s", legacy)
	}
	if len(changes) != 1 {
		t.Errorf("expected 1 deletion, got %d", len(changes))
	}
}

// TestCleanupDryRunKeepsFiles ensures dry-run reports but does not delete.
func TestCleanupDryRunKeepsFiles(t *testing.T) {
	tmp := t.TempDir()
	legacy := writeAsset(t, tmp, "some-other-skill", "sync_test.sh")

	changes, err := cleanupLegacyHarnessArtifacts(tmp, true)
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
	if _, err := os.Stat(legacy); err != nil {
		t.Errorf("dry-run must not delete: %s", legacy)
	}
	if len(changes) != 1 {
		t.Errorf("expected 1 reported change in dry-run, got %d", len(changes))
	}
}
