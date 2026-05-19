package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestService_RegistryIsolation(t *testing.T) {
	// 1.1 RED: Test internal/storage/service.go for SKILLSYNC_HOME prioritization.
	t.Run("prioritizes SKILLSYNC_HOME for registry path", func(t *testing.T) {
		tmpHome := t.TempDir()
		t.Setenv("SKILLSYNC_HOME", tmpHome)

		s := NewService("")
		expected := filepath.Join(tmpHome, "projects.json")
		if s.registryPath() != expected {
			t.Errorf("expected registryPath %q, got %q", expected, s.registryPath())
		}
	})
}

func TestService_GetProjectsFiltering(t *testing.T) {
	// 1.3 RED: Test GetProjects in internal/storage/service.go for filtering missing paths.
	tmpDir := t.TempDir()
	s := &Service{RootPath: filepath.Join(tmpDir, "storage")}

	// Create a dummy project that exists
	existingDir := filepath.Join(tmpDir, "existing")
	if err := os.MkdirAll(existingDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Register one existing and one missing
	if err := s.RegisterProject(existingDir); err != nil {
		t.Fatal(err)
	}
	if err := s.RegisterProject(filepath.Join(tmpDir, "missing")); err != nil {
		t.Fatal(err)
	}

	projects, err := s.GetProjects()
	if err != nil {
		t.Fatal(err)
	}

	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	} else if projects[0].Path != existingDir {
		t.Errorf("expected path %q, got %q", existingDir, projects[0].Path)
	}
}

func TestService_PruneRegistry(t *testing.T) {
	// 1.5 RED: Test PruneRegistry to remove dead entries from the registry file.
	tmpDir := t.TempDir()
	s := &Service{RootPath: filepath.Join(tmpDir, "storage")}

	existingDir := filepath.Join(tmpDir, "existing")
	os.MkdirAll(existingDir, 0755)
	missingDir := filepath.Join(tmpDir, "missing")

	s.RegisterProject(existingDir)
	s.RegisterProject(missingDir)

	// Verify both are in the file initially
	registry, _ := s.loadRegistry()
	if len(registry.Projects) != 2 {
		t.Errorf("expected 2 projects in registry file, got %d", len(registry.Projects))
	}

	if err := s.PruneRegistry(); err != nil {
		t.Fatal(err)
	}

	// Verify only existing remains in the file
	registry, _ = s.loadRegistry()
	if len(registry.Projects) != 1 {
		t.Errorf("expected 1 project in registry file after prune, got %d", len(registry.Projects))
	}
	if registry.Projects[0].Path != existingDir {
		t.Errorf("expected remaining path %q, got %q", existingDir, registry.Projects[0].Path)
	}
}
