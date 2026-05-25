package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"skillsync/tui/internal/parser"
	"skillsync/tui/internal/types"
	"time"
)

// Service handles global skill storage operations
type Service struct {
	RootPath string
}

// NewService creates a new storage service with default path if empty
func NewService(root string) *Service {
	if root == "" {
		if envHome := os.Getenv("SKILLSYNC_HOME"); envHome != "" {
			root = filepath.Join(envHome, "storage")
		} else {
			home, _ := os.UserHomeDir()
			root = filepath.Join(home, ".skillsync", "storage")
		}
	}
	return &Service{RootPath: root}
}

// Delete removes a skill's directory and all its contents from global storage.
// Returns ErrSkillNotFound if the directory does not exist.
// Returns an error if the skill name is empty or a filesystem error occurs.
func (s *Service) Delete(name string) error {
	if name == "" {
		return fmt.Errorf("skill name cannot be empty")
	}
	skillDir := filepath.Join(s.RootPath, name)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return fmt.Errorf("skill %q not found: %w", name, ErrSkillNotFound)
	}
	return os.RemoveAll(skillDir)
}

// Save persists a skill and its metadata to global storage
func (s *Service) Save(skill *types.Skill, metadata StoredMetadata) error {
	skillDir := filepath.Join(s.RootPath, skill.Name)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage dir: %w", err)
	}

	// Save SKILL.md
	skillPath := filepath.Join(skillDir, "SKILL.md")
	content, err := parser.Format(skill)
	if err != nil {
		return fmt.Errorf("failed to format skill: %w", err)
	}
	if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to save SKILL.md: %w", err)
	}

	// Save METADATA.json
	metaPath := filepath.Join(skillDir, "METADATA.json")
	metaBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	if err := os.WriteFile(metaPath, metaBytes, 0644); err != nil {
		return fmt.Errorf("failed to save METADATA.json: %w", err)
	}

	return nil
}

// List returns all skills stored in global storage
func (s *Service) List() ([]StoredSkill, error) {
	if _, err := os.Stat(s.RootPath); os.IsNotExist(err) {
		return []StoredSkill{}, nil
	}

	entries, err := os.ReadDir(s.RootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage root: %w", err)
	}

	var stored []StoredSkill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillID := entry.Name()
		metaPath := filepath.Join(s.RootPath, skillID, "METADATA.json")
		
		metaBytes, err := os.ReadFile(metaPath)
		if err != nil {
			continue // Skip if unreadable or missing
		}

		var metadata StoredMetadata
		if err := json.Unmarshal(metaBytes, &metadata); err != nil {
			continue // Skip corrupt
		}

		stored = append(stored, StoredSkill{
			ID:       skillID,
			Metadata: metadata,
		})
	}

	return stored, nil
}

// Load retrieves a skill's content from storage
func (s *Service) Load(skillID string) (string, error) {
	skillPath := filepath.Join(s.RootPath, skillID, "SKILL.md")
	content, err := os.ReadFile(skillPath)
	if err != nil {
		return "", fmt.Errorf("failed to load skill content: %w", err)
	}
	return string(content), nil
}

func (s *Service) registryPath() string {
	return filepath.Join(filepath.Dir(s.RootPath), "projects.json")
}

// RegisterProject adds or updates a project in the registry
func (s *Service) RegisterProject(path string) error {
	return s.registerProject(path, time.Now(), true)
}

// RegisterProjectInitial registers a project without setting a sync timestamp
func (s *Service) RegisterProjectInitial(path string) error {
	return s.registerProject(path, time.Time{}, false)
}

func (s *Service) registerProject(path string, lastSynced time.Time, overwrite bool) error {
	registry, err := s.loadRegistry()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	found := false
	for i, p := range registry.Projects {
		if p.Path == path {
			if overwrite {
				registry.Projects[i].LastSynced = lastSynced
			}
			found = true
			break
		}
	}

	if !found {
		registry.Projects = append(registry.Projects, ProjectInfo{
			Path:       path,
			LastSynced: lastSynced,
		})
	}

	return s.saveRegistry(registry)
}

// GetProjects returns all registered projects sorted by last synced descending
func (s *Service) GetProjects() ([]ProjectInfo, error) {
	registry, err := s.loadRegistry()
	if err != nil {
		if os.IsNotExist(err) {
			return []ProjectInfo{}, nil
		}
		return nil, fmt.Errorf("failed to load registry: %w", err)
	}

	var existing []ProjectInfo
	for _, p := range registry.Projects {
		if _, err := os.Stat(p.Path); err == nil {
			existing = append(existing, p)
		}
	}

	// Sort descending by LastSynced
	for i := 0; i < len(existing); i++ {
		for j := i + 1; j < len(existing); j++ {
			if existing[j].LastSynced.After(existing[i].LastSynced) {
				existing[i], existing[j] = existing[j], existing[i]
			}
		}
	}

	return existing, nil
}

func (s *Service) PruneRegistry() error {
	registry, err := s.loadRegistry()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to load registry: %w", err)
	}

	var existing []ProjectInfo
	for _, p := range registry.Projects {
		if _, err := os.Stat(p.Path); err == nil {
			existing = append(existing, p)
		}
	}

	if len(existing) == len(registry.Projects) {
		return nil // Nothing to prune
	}

	registry.Projects = existing
	return s.saveRegistry(registry)
}

func (s *Service) loadRegistry() (ProjectRegistry, error) {
	var registry ProjectRegistry
	data, err := os.ReadFile(s.registryPath())
	if err != nil {
		return registry, err
	}
	err = json.Unmarshal(data, &registry)
	return registry, err
}

func (s *Service) saveRegistry(registry ProjectRegistry) error {
	dir := filepath.Dir(s.registryPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	tmpFile := s.registryPath() + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary registry file: %w", err)
	}

	if err := os.Rename(tmpFile, s.registryPath()); err != nil {
		return fmt.Errorf("failed to rename temporary registry file: %w", err)
	}

	return nil
}
