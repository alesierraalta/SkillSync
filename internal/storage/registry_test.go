package storage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestRegisterProject(t *testing.T) {
	tmpDir := t.TempDir()
	root := filepath.Join(tmpDir, "storage")
	s := NewService(root)

	projectPath := t.TempDir()

	t.Run("creates registry if missing", func(t *testing.T) {
		err := s.RegisterProject(projectPath)
		if err != nil {
			t.Fatalf("RegisterProject failed: %v", err)
		}

		projects, err := s.GetProjects()
		if err != nil {
			t.Fatalf("GetProjects failed: %v", err)
		}

		if len(projects) != 1 {
			t.Errorf("expected 1 project, got %d", len(projects))
		}
	})

	t.Run("upsert logic - updates existing entry", func(t *testing.T) {
		// Register same project again
		err := s.RegisterProject(projectPath)
		if err != nil {
			t.Fatalf("RegisterProject failed: %v", err)
		}

		projects, err := s.GetProjects()
		if err != nil {
			t.Fatalf("GetProjects failed: %v", err)
		}

		if len(projects) != 1 {
			t.Errorf("expected 1 project after re-registration, got %d", len(projects))
		}
	})

	t.Run("sorting - last synced descending", func(t *testing.T) {
		p2 := t.TempDir()
		err := s.RegisterProject(p2)
		if err != nil {
			t.Fatalf("RegisterProject failed: %v", err)
		}

		p3 := t.TempDir()
		err = s.RegisterProject(p3)
		if err != nil {
			t.Fatalf("RegisterProject failed: %v", err)
		}

		projects, err := s.GetProjects()
		if err != nil {
			t.Fatalf("GetProjects failed: %v", err)
		}

		if len(projects) != 3 {
			t.Errorf("expected 3 projects, got %d", len(projects))
		}

		// p3 should be first (most recent)
		if projects[0].Path != p3 {
			t.Errorf("expected first project to be %s, got %s", p3, projects[0].Path)
		}
		// p2 should be second
		if projects[1].Path != p2 {
			t.Errorf("expected second project to be %s, got %s", p2, projects[1].Path)
		}
		// projectPath should be last
		if projects[2].Path != projectPath {
			t.Errorf("expected third project to be %s, got %s", projectPath, projects[2].Path)
		}
	})
}

func TestRegisterProjectInitial(t *testing.T) {
	tmpDir := t.TempDir()
	root := filepath.Join(tmpDir, "storage")
	s := NewService(root)

	projectPath := t.TempDir()

	t.Run("registers with zero time", func(t *testing.T) {
		err := s.RegisterProjectInitial(projectPath)
		if err != nil {
			t.Fatalf("RegisterProjectInitial failed: %v", err)
		}

		projects, err := s.GetProjects()
		if err != nil {
			t.Fatalf("GetProjects failed: %v", err)
		}

		if len(projects) != 1 {
			t.Errorf("expected 1 project, got %d", len(projects))
		}

		if !projects[0].LastSynced.IsZero() {
			t.Errorf("expected LastSynced to be zero, got %v", projects[0].LastSynced)
		}
	})

	t.Run("is idempotent - does not update existing zero time", func(t *testing.T) {
		err := s.RegisterProjectInitial(projectPath)
		if err != nil {
			t.Fatalf("RegisterProjectInitial failed: %v", err)
		}

		projects, err := s.GetProjects()
		if err != nil {
			t.Fatalf("GetProjects failed: %v", err)
		}

		if len(projects) != 1 {
			t.Errorf("expected 1 project, got %d", len(projects))
		}
		if !projects[0].LastSynced.IsZero() {
			t.Errorf("expected LastSynced to remain zero, got %v", projects[0].LastSynced)
		}
	})

	t.Run("does not overwrite existing non-zero time", func(t *testing.T) {
		p2 := t.TempDir()
		// First register normally (non-zero time)
		err := s.RegisterProject(p2)
		if err != nil {
			t.Fatal(err)
		}

		projects, _ := s.GetProjects()
		var originalTime time.Time
		for _, p := range projects {
			if p.Path == p2 {
				originalTime = p.LastSynced
				break
			}
		}

		if originalTime.IsZero() {
			t.Fatal("original time should not be zero")
		}

		// Now try to register initially
		err = s.RegisterProjectInitial(p2)
		if err != nil {
			t.Fatal(err)
		}

		projects, _ = s.GetProjects()
		var newTime time.Time
		for _, p := range projects {
			if p.Path == p2 {
				newTime = p.LastSynced
				break
			}
		}

		if !newTime.Equal(originalTime) {
			t.Errorf("expected time to be preserved, got %v instead of %v", newTime, originalTime)
		}
	})
}

func TestGetProjectsFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	root := filepath.Join(tmpDir, "storage")
	s := NewService(root)

	existingPath := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "does-not-exist")

	s.RegisterProject(existingPath)
	s.RegisterProject(nonExistentPath)

	t.Run("GetProjects filters non-existent paths", func(t *testing.T) {
		projects, err := s.GetProjects()
		if err != nil {
			t.Fatal(err)
		}

		for _, p := range projects {
			if p.Path == nonExistentPath {
				t.Errorf("expected %q to be filtered out", nonExistentPath)
			}
		}

		found := false
		for _, p := range projects {
			if p.Path == existingPath {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %q to be present", existingPath)
		}
	})
}

func TestPruneRegistry(t *testing.T) {
	tmpDir := t.TempDir()
	root := filepath.Join(tmpDir, "storage")
	s := NewService(root)

	existingPath := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "gone")

	s.RegisterProject(existingPath)
	s.RegisterProject(nonExistentPath)

	t.Run("PruneRegistry removes non-existent paths from file", func(t *testing.T) {
		err := s.PruneRegistry()
		if err != nil {
			t.Fatal(err)
		}

		// Verify via loadRegistry (bypass GetProjects filtering)
		registry, err := s.loadRegistry()
		if err != nil {
			t.Fatal(err)
		}

		if len(registry.Projects) != 1 {
			t.Errorf("expected 1 project in registry file, got %d", len(registry.Projects))
		}
		if registry.Projects[0].Path != existingPath {
			t.Errorf("expected %s to remain, got %s", existingPath, registry.Projects[0].Path)
		}
	})
}
