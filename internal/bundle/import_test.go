package bundle

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// createTestBundle builds a .skillsync zip at bundlePath with the given manifest
// and file entries. Each entry has content matching its name.
func createTestBundle(t *testing.T, bundlePath string, m Manifest, entries map[string]string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(bundlePath), 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)

	// Write manifest
	if err := writeManifestJSON(zw, m); err != nil {
		zw.Close()
		t.Fatal(err)
	}

	// Write entries
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			zw.Close()
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			zw.Close()
			t.Fatal(err)
		}
	}
	zw.Close()
}

// mockSync disables SyncAfterImport for tests and restores it on cleanup.
func mockSync(t *testing.T) {
	t.Helper()
	old := SyncAfterImport
	SyncAfterImport = func(root string) error { return nil }
	t.Cleanup(func() { SyncAfterImport = old })
}

func TestImportBasic(t *testing.T) {
	mockSync(t)
	projRoot := t.TempDir()
	targetDir := filepath.Join(projRoot, ".agents", "skills")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}

	bundlePath := filepath.Join(t.TempDir(), "test.skillsync")
	createTestBundle(t, bundlePath, Manifest{
		Version:   1,
		CreatedAt: time.Now(),
		CreatedBy: "synck",
		Skills: []ManifestSkill{
			{Name: "test-skill", OriginProvider: "claude"},
		},
	}, map[string]string{
		"skills/test-skill/SKILL.md":        "name: test-skill\n",
		"skills/test-skill/METADATA.json":   `{"skill_name":"test-skill"}`,
	})

	results, err := Import(bundlePath, ImportOptions{
		ProjectRoot: projRoot,
	})
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Skill != "test-skill" {
		t.Errorf("result Skill = %s, want test-skill", results[0].Skill)
	}
	if results[0].Status != "installed" {
		t.Errorf("result Status = %s, want installed", results[0].Status)
	}
	if results[0].Error != nil {
		t.Errorf("result Error = %v", results[0].Error)
	}

	// Verify file was written
	skillPath := filepath.Join(targetDir, "test-skill", "SKILL.md")
	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}
	if string(content) != "name: test-skill\n" {
		t.Errorf("SKILL.md content = %q, want %q", string(content), "name: test-skill\n")
	}
}

func TestImportDefaultSkip(t *testing.T) {
	mockSync(t)
	projRoot := t.TempDir()
	targetDir := filepath.Join(projRoot, ".agents", "skills")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Pre-create an existing skill
	existDir := filepath.Join(targetDir, "existing-skill")
	os.MkdirAll(existDir, 0755)
	os.WriteFile(filepath.Join(existDir, "SKILL.md"), []byte("original content\n"), 0644)

	bundlePath := filepath.Join(t.TempDir(), "test.skillsync")
	createTestBundle(t, bundlePath, Manifest{
		Version:   1,
		CreatedAt: time.Now(),
		CreatedBy: "synck",
		Skills: []ManifestSkill{
			{Name: "existing-skill"},
		},
	}, map[string]string{
		"skills/existing-skill/SKILL.md": "new content\n",
	})

	results, err := Import(bundlePath, ImportOptions{
		ProjectRoot: projRoot,
	})
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Status != "skipped" {
		t.Errorf("Status = %s, want skipped", results[0].Status)
	}

	// Verify original content was preserved
	content, _ := os.ReadFile(filepath.Join(targetDir, "existing-skill", "SKILL.md"))
	if string(content) != "original content\n" {
		t.Errorf("content = %q, want original", string(content))
	}
}

func TestImportOverwrite(t *testing.T) {
	mockSync(t)
	projRoot := t.TempDir()
	targetDir := filepath.Join(projRoot, ".agents", "skills")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Pre-create an existing skill
	os.MkdirAll(filepath.Join(targetDir, "overwrite-me"), 0755)
	os.WriteFile(filepath.Join(targetDir, "overwrite-me", "SKILL.md"), []byte("old content\n"), 0644)

	bundlePath := filepath.Join(t.TempDir(), "test.skillsync")
	createTestBundle(t, bundlePath, Manifest{
		Version:   1,
		CreatedAt: time.Now(),
		CreatedBy: "synck",
		Skills: []ManifestSkill{
			{Name: "overwrite-me"},
		},
	}, map[string]string{
		"skills/overwrite-me/SKILL.md": "new content\n",
	})

	results, err := Import(bundlePath, ImportOptions{
		ProjectRoot: projRoot,
		OnDuplicate: DuplicateOverwrite,
	})
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	if results[0].Status != "overwritten" {
		t.Errorf("Status = %s, want overwritten", results[0].Status)
	}

	content, _ := os.ReadFile(filepath.Join(targetDir, "overwrite-me", "SKILL.md"))
	if string(content) != "new content\n" {
		t.Errorf("content = %q, want new content", string(content))
	}
}

func TestImportTraversalRejection(t *testing.T) {
	mockSync(t)
	projRoot := t.TempDir()
	targetDir := filepath.Join(projRoot, ".agents", "skills")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		entry    string
		wantErr  string
	}{
		{
			name:    "parent directory traversal",
			entry:   "skills/foo/../../../etc/passwd",
			wantErr: "traversal",
		},
		{
			name:    "absolute path Unix",
			entry:   "/etc/passwd",
			wantErr: "absolute",
		},
		{
			name:    "undeclared skill name",
			entry:   "skills/undeclared/SKILL.md",
			wantErr: "undeclared",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundlePath := filepath.Join(t.TempDir(), "bad.skillsync")
			createTestBundle(t, bundlePath, Manifest{
				Version:   1,
				CreatedAt: time.Now(),
				CreatedBy: "synck",
				Skills: []ManifestSkill{
					{Name: "foo"},
				},
			}, map[string]string{
				tt.entry: "evil content",
			})

			_, err := Import(bundlePath, ImportOptions{
				ProjectRoot: projRoot,
			})
			if err == nil {
				t.Fatal("expected error for malicious entry")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestImportTargetSelection(t *testing.T) {
	mockSync(t)
	projRoot := t.TempDir()

	// Create multiple provider directories
	for _, dir := range []string{".agents/skills", ".claude/skills", ".opencode/skills"} {
		os.MkdirAll(filepath.Join(projRoot, dir), 0755)
	}

	bundlePath := filepath.Join(t.TempDir(), "multi.skillsync")
	createTestBundle(t, bundlePath, Manifest{
		Version:   1,
		CreatedAt: time.Now(),
		CreatedBy: "synck",
		Skills: []ManifestSkill{
			{Name: "foo"},
		},
	}, map[string]string{
		"skills/foo/SKILL.md": "name: foo\n",
	})

	// Import to two providers explicitly (not .opencode)
	results, err := Import(bundlePath, ImportOptions{
		ProjectRoot: projRoot,
		Targets: []string{
			filepath.Join(projRoot, ".agents", "skills"),
			filepath.Join(projRoot, ".claude", "skills"),
		},
	})
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	// Verify foo was installed in both target providers
	for _, dir := range []string{".agents/skills", ".claude/skills"} {
		skillFile := filepath.Join(projRoot, dir, "foo", "SKILL.md")
		if _, err := os.Stat(skillFile); os.IsNotExist(err) {
			t.Errorf("skill not found in %s", dir)
		}
	}
	// Verify it was NOT installed in .opencode/skills (not in targets)
	notInstalled := filepath.Join(projRoot, ".opencode", "skills", "foo", "SKILL.md")
	if _, err := os.Stat(notInstalled); !os.IsNotExist(err) {
		t.Errorf("skill should not be in .opencode/skills (not in targets)")
	}
}

func TestImportEmptyBundle(t *testing.T) {
	mockSync(t)
	projRoot := t.TempDir()
	os.MkdirAll(filepath.Join(projRoot, ".agents", "skills"), 0755)

	bundlePath := filepath.Join(t.TempDir(), "empty.skillsync")
	createTestBundle(t, bundlePath, Manifest{
		Version:   1,
		CreatedAt: time.Now(),
		CreatedBy: "synck",
	}, nil)

	results, err := Import(bundlePath, ImportOptions{
		ProjectRoot: projRoot,
	})
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}

func TestImportOverwriteRollback(t *testing.T) {
	mockSync(t)
	projRoot := t.TempDir()
	targetDir := filepath.Join(projRoot, ".agents", "skills")
	os.MkdirAll(targetDir, 0755)
	orig := "original\n"
	skillDir := filepath.Join(targetDir, "test-skill")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(orig), 0644)

	bundlePath := filepath.Join(t.TempDir(), "rollback.skillsync")
	createTestBundle(t, bundlePath, Manifest{
		Version: 1, CreatedAt: time.Now(), CreatedBy: "synck",
		Skills: []ManifestSkill{{Name: "test-skill"}},
	}, map[string]string{
		"skills/test-skill/SKILL.md":    "replacement\n",
		"skills/test-skill/bad|file.md": "evil\n",
	})

	_, err := Import(bundlePath, ImportOptions{ProjectRoot: projRoot, OnDuplicate: DuplicateOverwrite})
	// On Windows, the | char in filename causes staging write failure.
	// Import returns an error; original files must be untouched.
	if err == nil {
		// On non-Windows, | is valid — import may succeed
		got, _ := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
		if string(got) != "replacement\n" {
			t.Errorf("got %q, want replacement", string(got))
		}
		return
	}
	// Write failed during staging — original must be preserved
	got, _ := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if string(got) != orig {
		t.Errorf("rollback: got %q, want %q", string(got), orig)
	}
}

func TestImport_OversizedManifestRejected(t *testing.T) {
	// Create a manifest larger than MaxManifestBytes (1 MiB)
	// by adding dummy skills with long descriptions.
	var skills []ManifestSkill
	for i := 0; i < 5000; i++ {
		skills = append(skills, ManifestSkill{
			Name:        fmt.Sprintf("skill-%d", i),
			Description: strings.Repeat("x", 300),
		})
	}
	m := Manifest{
		Version:   1,
		CreatedAt: time.Now(),
		CreatedBy: "synck",
		Skills:    skills,
	}
	bundlePath := filepath.Join(t.TempDir(), "bigmanifest.skillsync")
	createTestBundle(t, bundlePath, m, nil)

	_, err := Import(bundlePath, ImportOptions{ProjectRoot: t.TempDir()})
	if err == nil {
		t.Fatal("expected error for oversized manifest")
	}
	if !strings.Contains(err.Error(), "exceeds max size") {
		t.Errorf("error = %q, want 'exceeds max size'", err.Error())
	}
}

func TestImport_EmptyBundleRejected(t *testing.T) {
	// Create a valid zip with NO manifest
	bundlePath := filepath.Join(t.TempDir(), "nomanifest.skillsync")
	f, _ := os.Create(bundlePath)
	zw := zip.NewWriter(f)
	zw.Create("somefile.txt") // non-manifest entry
	zw.Close()
	f.Close()

	_, err := Import(bundlePath, ImportOptions{ProjectRoot: t.TempDir()})
	if err == nil {
		t.Fatal("expected error for bundle without manifest")
	}
}

func TestSafeZipPath_EdgeCases(t *testing.T) {
	allowed := map[string]bool{"valid": true}
	for _, tt := range []struct{ name, entry, want string }{
		{"null byte", "skills/valid/foo\x00bar.md", "null byte"},
		{"backslash", "skills\\valid\\SKILL.md", "backslash"},
		{"drive-relative", "C:skills/valid/SKILL.md", "drive-relative"},
		{"invalid charset", "skills/💥/SKILL.md", "invalid skill name"},
	} {
		t.Run("reject_"+tt.name, func(t *testing.T) {
			_, _, err := safeZipPath(tt.entry, allowed)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
	for _, tt := range []struct{ name, entry string }{
		{"simple", "skills/valid/SKILL.md"},
		{"dot-component", "skills/valid/./SKILL.md"},
	} {
		t.Run("accept_"+tt.name, func(t *testing.T) {
			sn, _, err := safeZipPath(tt.entry, allowed)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sn != "valid" {
				t.Errorf("got %q, want valid", sn)
			}
		})
	}
}
