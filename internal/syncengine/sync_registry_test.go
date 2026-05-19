package syncengine

import (
	"os"
	"path/filepath"
	"testing"
	"skillsync/tui/internal/storage"
)

func TestSync_RegistersProject(t *testing.T) {
	tmpDir := t.TempDir()
	storageRoot := filepath.Join(tmpDir, "storage")
	s := storage.NewService(storageRoot)
	
	opts := SyncOptions{
		Storage: s,
	}
	
	os.MkdirAll(filepath.Join(tmpDir, ".agents", "skills"), 0755)

	_, err := Sync(tmpDir, opts)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	
	projects, err := s.GetProjects()
	if err != nil {
		t.Fatalf("GetProjects failed: %v", err)
	}
	
	if len(projects) == 0 {
		t.Error("expected project to be registered")
	}
}

func TestSync_FailureDoesNotRegister(t *testing.T) {
	tmpDir := t.TempDir()
	storageRoot := filepath.Join(tmpDir, "storage")
	s := storage.NewService(storageRoot)
	
	opts := SyncOptions{
		Storage: s,
	}
	
	// Create a file and use it as root
	rootFile := filepath.Join(tmpDir, "rootfile")
	os.WriteFile(rootFile, []byte("not a dir"), 0644)
	
	_, err := Sync(rootFile, opts)
	if err == nil {
		t.Fatal("expected Sync to fail for file as root")
	}
	
	projects, _ := s.GetProjects()
	if len(projects) != 0 {
		t.Errorf("expected 0 projects registered on failure, got %d", len(projects))
	}
}
