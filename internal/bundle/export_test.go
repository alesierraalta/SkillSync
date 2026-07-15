package bundle

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/vault"
)

// setupTestStorage creates a temporary vault with one or more skills.
// Each skill gets a SKILL.md and METADATA.json.
func setupTestStorage(t *testing.T, skillNames ...string) *vault.Service {
	t.Helper()
	root := t.TempDir()
	svc := vault.NewServiceWithRoot(root)

	for _, name := range skillNames {
		skillDir := filepath.Join(root, name)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("name: "+name+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
		meta := storage.StoredMetadata{
			SkillName:     name,
			Description:   "Test skill " + name,
			OriginProject: "test-project",
			OriginPath:    filepath.Join("some", ".claude", "skills", name),
			SavedAt:       time.Now(),
		}
		data, _ := json.Marshal(meta)
		if err := os.WriteFile(filepath.Join(skillDir, "METADATA.json"), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return svc
}

func TestExportSingleSkill(t *testing.T) {
	svc := setupTestStorage(t, "foo")
	output := filepath.Join(t.TempDir(), "out.skillsync")

	path, err := Export(svc, []string{"foo"}, output)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if path != output {
		t.Errorf("Export returned path %q, want %q", path, output)
	}

	// Verify the zip exists and contains expected entries
	r, err := zip.OpenReader(output)
	if err != nil {
		t.Fatalf("open bundle: %v", err)
	}
	defer r.Close()

	var foundManifest, foundSkill, foundMeta bool
	for _, f := range r.File {
		switch f.Name {
		case "manifest.json":
			foundManifest = true
		case "skills/foo/SKILL.md":
			foundSkill = true
		case "skills/foo/METADATA.json":
			foundMeta = true
		}
	}
	if !foundManifest {
		t.Error("bundle missing manifest.json")
	}
	if !foundSkill {
		t.Error("bundle missing skills/foo/SKILL.md")
	}
	if !foundMeta {
		t.Error("bundle missing skills/foo/metadata.json")
	}

	// Verify manifest content
	m, err := ReadManifest(output)
	if err != nil {
		t.Fatalf("ReadManifest failed: %v", err)
	}
	if m.Version != 1 {
		t.Errorf("manifest version = %d, want 1", m.Version)
	}
	if m.CreatedBy != "synck" {
		t.Errorf("manifest CreatedBy = %s, want synck", m.CreatedBy)
	}
	if len(m.Skills) != 1 {
		t.Fatalf("manifest has %d skills, want 1", len(m.Skills))
	}
	if m.Skills[0].Name != "foo" {
		t.Errorf("manifest skill name = %s, want foo", m.Skills[0].Name)
	}
	if m.Skills[0].OriginProvider != "claude" {
		t.Errorf("manifest origin_provider = %s, want claude (from OriginPath)", m.Skills[0].OriginProvider)
	}
}

func TestExportMultipleSkills(t *testing.T) {
	svc := setupTestStorage(t, "alpha", "beta")
	output := filepath.Join(t.TempDir(), "multi.skillsync")

	_, err := Export(svc, []string{"alpha", "beta"}, output)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	m, err := ReadManifest(output)
	if err != nil {
		t.Fatalf("ReadManifest failed: %v", err)
	}
	if len(m.Skills) != 2 {
		t.Fatalf("manifest has %d skills, want 2", len(m.Skills))
	}
	names := make(map[string]bool)
	for _, s := range m.Skills {
		names[s.Name] = true
	}
	if !names["alpha"] {
		t.Error("manifest missing alpha")
	}
	if !names["beta"] {
		t.Error("manifest missing beta")
	}
}

func TestExportMissingSkillFailsFast(t *testing.T) {
	svc := setupTestStorage(t, "foo")
	output := filepath.Join(t.TempDir(), "fail.skillsync")

	// Export a skill that exists AND one that doesn't
	_, err := Export(svc, []string{"foo", "nonexistent"}, output)
	if err == nil {
		t.Fatal("expected error for nonexistent skill")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention the missing skill name, got: %v", err)
	}

	// Verify no bundle was created (fail-fast)
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Error("bundle was created despite missing skill (expected atomic fail-fast)")
	}
}

func TestExportDefaultOutputPath(t *testing.T) {
	svc := setupTestStorage(t, "myskill")

	// chdir to a temp dir so the default output path writes there, not in the package dir
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	// Empty output path should default
	path, err := Export(svc, []string{"myskill"}, "")
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if path == "" {
		t.Fatal("Export returned empty path")
	}
	if !strings.HasSuffix(path, "myskill.skillsync") {
		t.Errorf("default path should end with myskill.skillsync, got: %s", path)
	}
	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("bundle not found at default path: %s", path)
	}
}

func TestExportCustomOutputPath(t *testing.T) {
	svc := setupTestStorage(t, "bar")
	output := filepath.Join(t.TempDir(), "custom", "path", "backup.skillsync")

	path, err := Export(svc, []string{"bar"}, output)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if path != output {
		t.Errorf("Export returned %q, want %q", path, output)
	}
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Errorf("bundle not found at custom path: %s", output)
	}
}

func TestExportImportRoundTrip(t *testing.T) {
	// Regression test: Export followed by Import must succeed even when
	// Export creates ZIP directory entries (subdirectories with trailing /).
	// Import must safely skip directory-only entries.
	mockSync(t)

	// Create storage with a skill that has a subdirectory
	root := t.TempDir()
	svc := vault.NewServiceWithRoot(root)

	skillDir := filepath.Join(root, "mytool")
	if err := os.MkdirAll(filepath.Join(skillDir, "assets"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("name: mytool\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "assets", "helper.sh"), []byte("echo hi\n"), 0644); err != nil {
		t.Fatal(err)
	}
	meta := storage.StoredMetadata{
		SkillName: "mytool",
		SavedAt:   time.Now(),
	}
	data, _ := json.Marshal(meta)
	os.WriteFile(filepath.Join(skillDir, "METADATA.json"), data, 0644)

	// Export to bundle
	bundlePath := filepath.Join(t.TempDir(), "roundtrip.skillsync")
	_, err := Export(svc, []string{"mytool"}, bundlePath)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Import to a fresh project
	projRoot := t.TempDir()
	targetDir := filepath.Join(projRoot, ".agents", "skills")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}

	results, err := Import(bundlePath, ImportOptions{
		ProjectRoot: projRoot,
		OnDuplicate: DuplicateOverwrite,
	})
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Status != "installed" {
		t.Errorf("result Status = %s, want installed", results[0].Status)
	}
	if results[0].Error != nil {
		t.Errorf("result Error = %v", results[0].Error)
	}

	// Verify files were installed correctly
	skillPath := filepath.Join(targetDir, "mytool")
	checkFile := func(relPath, wantContent string) {
		fullPath := filepath.Join(skillPath, relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("read %s: %v", relPath, err)
			return
		}
		if string(content) != wantContent {
			t.Errorf("%s content = %q, want %q", relPath, string(content), wantContent)
		}
	}
	checkFile("SKILL.md", "name: mytool\n")
	checkFile(filepath.Join("assets", "helper.sh"), "echo hi\n")
}

func TestExportWholeDirectory(t *testing.T) {
	// Verify that a skill with extra files (assets) includes them
	root := t.TempDir()
	svc := vault.NewServiceWithRoot(root)

	skillDir := filepath.Join(root, "with-assets")
	if err := os.MkdirAll(filepath.Join(skillDir, "assets"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("name: with-assets\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "assets", "helper.sh"), []byte("echo hi\n"), 0644); err != nil {
		t.Fatal(err)
	}
	meta := storage.StoredMetadata{
		SkillName: "with-assets",
		SavedAt:   time.Now(),
	}
	data, _ := json.Marshal(meta)
	os.WriteFile(filepath.Join(skillDir, "METADATA.json"), data, 0644)

	output := filepath.Join(t.TempDir(), "assets-test.skillsync")
	_, err := Export(svc, []string{"with-assets"}, output)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	r, err := zip.OpenReader(output)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	var foundAsset bool
	for _, f := range r.File {
		if f.Name == "skills/with-assets/assets/helper.sh" {
			foundAsset = true
		}
	}
	if !foundAsset {
		t.Error("bundle missing skills/with-assets/assets/helper.sh")
	}
}

func TestExportMissingSkill_ReturnsNonZero(t *testing.T) {
	svc := setupTestStorage(t, "existing")
	output := filepath.Join(t.TempDir(), "out.skillsync")
	_, err := Export(svc, []string{"existing", "nonexistent"}, output)
	if err == nil {
		t.Fatal("expected error for missing skill")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error must mention missing skill name, got: %v", err)
	}
}

func TestExportVaultWired(t *testing.T) {
	// Verify Export accepts a vault service and reads skills from it
	svc := setupTestStorage(t, "vaultskill")
	output := filepath.Join(t.TempDir(), "vault.skillsync")
	path, err := Export(svc, []string{"vaultskill"}, output)
	if err != nil {
		t.Fatalf("Export via vault failed: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("bundle not created via vault-backed Export")
	}
}


