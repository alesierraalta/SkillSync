// Package vault reframes ~/.skillsync/storage/ as the persistent editable vault.
// It wraps *storage.Service to add core-skill protection, a transient
// multi-select model (Selection), and a stable identity as the single source
// of truth for export and import operations.
//
// Storage IS the vault — the same path, no migration, no parallel store.
// Save-to-storage from Global Skills IS save-to-vault.
package vault

import (
	"skillsync/tui/internal/remove"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/types"
)

// Service wraps *storage.Service with core-skill protection and a stable
// vault identity. All vault read/write operations route through this type.
type Service struct {
	store *storage.Service
}

// NewService creates a vault Service wrapping the default storage path.
// Equivalent to storage.NewService("") — same path, no migration.
func NewService() *Service {
	return &Service{store: storage.NewService("")}
}

// NewServiceWithRoot creates a vault Service wrapping a specific storage root.
// Used in tests and when an explicit storage path is known.
func NewServiceWithRoot(root string) *Service {
	return &Service{store: storage.NewService(root)}
}

// Save persists a skill and its metadata to the vault (storage).
func (s *Service) Save(skill *types.Skill, metadata storage.StoredMetadata) error {
	return s.store.Save(skill, metadata)
}

// List returns all skills currently in the vault (storage).
func (s *Service) List() ([]storage.StoredSkill, error) {
	return s.store.List()
}

// Remove deletes a skill from the vault by name. Returns an error if the
// skill is a core skill (protected). Delegates to storage.Delete for
// non-core skills.
func (s *Service) Remove(name string) error {
	if remove.IsCoreSkill(name) {
		return &ErrCoreSkill{Name: name}
	}
	return s.store.Delete(name)
}

// RootPath returns the filesystem path to the vault root directory.
func (s *Service) RootPath() string {
	return s.store.RootPath
}

// ErrCoreSkill is returned when attempting to delete a protected core skill.
type ErrCoreSkill struct {
	Name string
}

func (e *ErrCoreSkill) Error() string {
	return "cannot remove core skill \"" + e.Name + "\""
}

// Selection is a transient multi-select set of vault skill names.
// It is not persisted — each TUI session starts with an empty selection.
type Selection map[string]bool
