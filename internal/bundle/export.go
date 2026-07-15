package bundle

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/vault"
)

// Export creates a .skillsync zip bundle containing the specified skills.
// It performs atomic fail-fast: if any skill cannot be read, no bundle is written.
// If outputPath is empty, it defaults to "{cwd}/{first-skill}.skillsync".
// Returns the path to the created bundle.
func Export(svc *vault.Service, names []string, outputPath string) (string, error) {
	if len(names) == 0 {
		return "", fmt.Errorf("at least one skill name is required")
	}

	// Resolve output path
	if outputPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working directory: %w", err)
		}
		if len(names) == 1 {
			outputPath = filepath.Join(cwd, names[0]+".skillsync")
		} else {
			outputPath = filepath.Join(cwd, "skillsync-bundle.skillsync")
		}
	}

	// Collect skills metadata and verify all exist before writing
	now := time.Now().UTC()
	manifest := Manifest{
		Version:   ManifestVersion,
		CreatedAt: now,
		CreatedBy: "synck",
	}

	// Build a skill lookup map from storage
	storedList, err := svc.List()
	if err != nil {
		return "", fmt.Errorf("list stored skills: %w", err)
	}
	storedByName := make(map[string]storage.StoredSkill, len(storedList))
	for _, s := range storedList {
		storedByName[s.ID] = s
	}

	// Verify each requested skill exists and build manifest entries
	for _, name := range names {
		stored, ok := storedByName[name]
		if !ok {
			return "", fmt.Errorf("skill %q: %w", name, ErrSkillNotFound)
		}
		originProvider := deriveOriginProvider(stored.Metadata.OriginPath)
		manifest.Skills = append(manifest.Skills, ManifestSkill{
			Name:           name,
			OriginProvider: originProvider,
			Description:    stored.Metadata.Description,
		})
	}

	// Sort skills for deterministic output
	sort.Slice(manifest.Skills, func(i, j int) bool {
		return manifest.Skills[i].Name < manifest.Skills[j].Name
	})

	// Pre-walk skills to estimate total size and file count before writing
	var totalSize int64
	var fileCount int
	for _, ms := range manifest.Skills {
		skillDir := filepath.Join(svc.RootPath(), ms.Name)
		err := filepath.WalkDir(skillDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			totalSize += info.Size()
			fileCount++
			if totalSize > MaxTotalBytes {
				return fmt.Errorf("total export size %d exceeds maximum %d bytes", totalSize, MaxTotalBytes)
			}
			if fileCount > MaxFileCount {
				return fmt.Errorf("export file count %d exceeds maximum %d", fileCount, MaxFileCount)
			}
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("pre-check skill %q: %w", ms.Name, err)
		}
	}

	// Write to a temp file, then rename atomically
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	tmpPath := outputPath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}

	// Ensure cleanup on failure
	cleanup := true
	defer func() {
		if cleanup {
			f.Close()
			os.Remove(tmpPath)
		}
	}()

	zw := zip.NewWriter(f)

	// Write manifest
	if err := writeManifestJSON(zw, manifest); err != nil {
		return "", err
	}

	// Write each skill's files
	for _, ms := range manifest.Skills {
		skillDir := filepath.Join(svc.RootPath(), ms.Name)

		// Walk entire skill directory to include assets, references, etc.
		err := filepath.WalkDir(skillDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(svc.RootPath(), path)
			if err != nil {
				return err
			}
			// Normalize to forward slashes for zip
			zipPath := filepath.ToSlash(relPath)

			if d.IsDir() {
				// Add directory entry (not strictly required but good practice)
				_, err := zw.Create("skills/" + zipPath + "/")
				return err
			}

			w, err := zw.Create("skills/" + zipPath)
			if err != nil {
				return fmt.Errorf("create zip entry %s: %w", zipPath, err)
			}

			src, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("open %s: %w", path, err)
			}
			_, err = io.Copy(w, src)
			src.Close()
			if err != nil {
				return fmt.Errorf("copy %s: %w", path, err)
			}
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("export skill %q: %w", ms.Name, err)
		}
	}

	// Close the zip writer
	if err := zw.Close(); err != nil {
		return "", fmt.Errorf("close zip: %w", err)
	}

	// Close the temp file
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("close temp file: %w", err)
	}

	// Rename atomically
	if err := os.Rename(tmpPath, outputPath); err != nil {
		return "", fmt.Errorf("rename to output: %w", err)
	}

	cleanup = false
	return outputPath, nil
}
