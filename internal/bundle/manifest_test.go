package bundle

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManifestRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	bundlePath := filepath.Join(tmpDir, "test.skillsync")

	// Create a minimal bundle with manifest
	createdAt := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	skills := []ManifestSkill{
		{Name: "skill-a", OriginProvider: "claude", Description: "Skill A"},
		{Name: "skill-b", OriginProvider: "opencode"},
	}

	f, err := os.Create(bundlePath)
	if err != nil {
		t.Fatal(err)
	}

	zw := zip.NewWriter(f)
	if err := writeManifestJSON(zw, Manifest{
		Version:   1,
		CreatedAt: createdAt,
		CreatedBy: "synck",
		Skills:    skills,
	}); err != nil {
		zw.Close()
		f.Close()
		t.Fatal(err)
	}
	zw.Close()
	f.Close()

	// Read back
	got, err := ReadManifest(bundlePath)
	if err != nil {
		t.Fatalf("ReadManifest failed: %v", err)
	}

	if got.Version != 1 {
		t.Errorf("Version = %d, want 1", got.Version)
	}
	if !got.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, createdAt)
	}
	if got.CreatedBy != "synck" {
		t.Errorf("CreatedBy = %s, want synck", got.CreatedBy)
	}
	if len(got.Skills) != 2 {
		t.Fatalf("len(Skills) = %d, want 2", len(got.Skills))
	}
	if got.Skills[0].Name != "skill-a" {
		t.Errorf("Skills[0].Name = %s, want skill-a", got.Skills[0].Name)
	}
	if got.Skills[0].OriginProvider != "claude" {
		t.Errorf("Skills[0].OriginProvider = %s, want claude", got.Skills[0].OriginProvider)
	}
	if got.Skills[1].Name != "skill-b" {
		t.Errorf("Skills[1].Name = %s, want skill-b", got.Skills[1].Name)
	}
}

func TestReadManifest_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := ReadManifest(filepath.Join(tmpDir, "nonexistent.skillsync"))
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestReadManifest_InvalidZip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid.skillsync")
	if err := os.WriteFile(path, []byte("not a zip"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := ReadManifest(path)
	if err == nil {
		t.Error("expected error for invalid zip")
	}
}

func TestReadManifest_MissingManifest(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "empty.skillsync")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	// Write some other file but no manifest.json
	w, _ := zw.Create("skills/foo/SKILL.md")
	w.Write([]byte("name: foo"))
	zw.Close()
	f.Close()

	_, err = ReadManifest(path)
	if err == nil {
		t.Error("expected error for bundle without manifest.json")
	}
}

func TestReadManifest_VersionMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "badver.skillsync")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)

	w, _ := zw.Create("manifest.json")
	json.NewEncoder(w).Encode(Manifest{
		Version:   99,
		CreatedAt: time.Now(),
		CreatedBy: "synck",
	})
	zw.Close()
	f.Close()

	_, err = ReadManifest(path)
	if err == nil {
		t.Error("expected error for unsupported version")
	}
}
