package remove

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skillsync/tui/internal/storage"
)

// Options controls the behavior of RemoveByID.
type Options struct {
	Local bool // skip global storage deletion
	Force bool // skip confirmation
}

// Service orchestrates the full removal of a skill from a project.
type Service struct {
	Storage  *storage.Service // global storage (may be nil)
	RootPath string           // project root
}

// IsCoreSkill returns true if the skill is protected and cannot be removed.
func IsCoreSkill(name string) bool {
	return coreSkills[name]
}

// coreSkills are skills that cannot be deleted.
var coreSkills = map[string]bool{
	"skill-creator": true,
	"skill-sync":    true,
	"find-skills":   true,
}

// mirrorProviders are the provider directories to scan for skill mirrors.
var mirrorProviders = []string{
	".opencode",
	".claude",
	".gemini",
	".cursor",
	".copilot",
	".qwen",
}

// MultiError collects multiple errors and implements the error interface.
type MultiError struct {
	Errors []error
}

func (m *MultiError) Error() string {
	msgs := make([]string, len(m.Errors))
	for i, e := range m.Errors {
		msgs[i] = e.Error()
	}
	return strings.Join(msgs, "; ")
}

// RemoveByID removes a skill by name from the project.
// It performs best-effort cleanup across all locations: canonical, mirrors, storage, and lock.
func (s *Service) RemoveByID(name string, opts Options) error {
	// Step 1: Validate — reject core skills
	if coreSkills[name] {
		return fmt.Errorf("cannot delete core skill %q", name)
	}

	var errs []error

	// Step 2: Delete canonical — <RootPath>/.agents/skills/<name>/
	canonicalDir := filepath.Join(s.RootPath, ".agents", "skills", name)
	if _, statErr := os.Stat(canonicalDir); statErr == nil {
		if err := os.RemoveAll(canonicalDir); err != nil {
			errs = append(errs, fmt.Errorf("canonical: %w", err))
		}
	}

	// Step 3: Delete mirrors — walk all provider directories
	for _, prov := range mirrorProviders {
		mirrorDir := filepath.Join(s.RootPath, prov, "skills", name)
		if _, err := os.Stat(mirrorDir); err == nil {
			if err := os.RemoveAll(mirrorDir); err != nil {
				errs = append(errs, fmt.Errorf("mirror %s: %w", prov, err))
			}
		}
	}

	// Step 4: Delete from global storage (unless opts.Local)
	if !opts.Local && s.Storage != nil {
		if err := s.Storage.Delete(name); err != nil {
			// "Not found" in storage is expected — skill may not have been synced
			if !errors.Is(err, storage.ErrSkillNotFound) {
				errs = append(errs, fmt.Errorf("storage: %w", err))
			}
		}
	}

	// Step 5: Update skills-lock.json — remove entry for name
	if err := removeFromLock(s.RootPath, name); err != nil {
		errs = append(errs, fmt.Errorf("lock: %w", err))
	}

	// Steps 6-7: AGENTS.md + OpenCode regeneration handled at caller level
	// (Backend.RemoveSkill and runRemove call ui.RegenerateAfterDelete)

	if len(errs) > 0 {
		return &MultiError{Errors: errs}
	}
	return nil
}

// skillsLock represents the structure of skills-lock.json
type skillsLock struct {
	Version int                    `json:"version"`
	Skills  map[string]interface{} `json:"skills"`
}

// removeFromLock removes a skill entry from skills-lock.json.
// Uses atomic write (temp + rename). No-ops if lock file doesn't exist or skill not present.
func removeFromLock(root, name string) error {
	lockPath := filepath.Join(root, "skills-lock.json")

	data, err := os.ReadFile(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No lock file = nothing to update
		}
		return fmt.Errorf("read lock: %w", err)
	}

	var lock skillsLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return fmt.Errorf("parse lock: %w", err)
	}

	if lock.Skills == nil {
		return nil
	}

	if _, exists := lock.Skills[name]; !exists {
		return nil // Not in lock = nothing to do
	}

	delete(lock.Skills, name)

	// Atomic write: temp file + rename
	tmpPath := lockPath + ".tmp"
	newData, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal lock: %w", err)
	}
	if err := os.WriteFile(tmpPath, newData, 0644); err != nil {
		return fmt.Errorf("write lock tmp: %w", err)
	}
	if err := os.Rename(tmpPath, lockPath); err != nil {
		return fmt.Errorf("rename lock: %w", err)
	}

	return nil
}
