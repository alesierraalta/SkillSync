package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"skillsync/tui/internal/types"
)

// Service handles global skill storage operations
type Service struct {
	RootPath string
}

// NewService creates a new storage service with default path if empty
func NewService(root string) *Service {
	if root == "" {
		home, _ := os.UserHomeDir()
		root = filepath.Join(home, ".skillsync", "storage")
	}
	return &Service{RootPath: root}
}

// Save persists a skill and its metadata to global storage
func (s *Service) Save(skill *types.Skill, metadata StoredMetadata) error {
	skillDir := filepath.Join(s.RootPath, skill.Name)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage dir: %w", err)
	}

	// Save SKILL.md
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skill.RawBody), 0644); err != nil {
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
