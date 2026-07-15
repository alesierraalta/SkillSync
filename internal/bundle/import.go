package bundle

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"skillsync/tui/internal/opencode"
)

// validSkillName matches allowed skill name characters.
var validSkillName = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// SyncAfterImport is called after all skills are written to synchronize
// registries. It can be overridden in tests.
var SyncAfterImport = func(root string) error {
	_, err := opencode.SyncSkills(root, opencode.Options{})
	return err
}

// Import extracts skills from a .skillsync bundle into the project's
// provider directories. It validates every zip entry against path traversal
// attacks and only extracts entries belonging to manifest-declared skills.
func Import(bundlePath string, opts ImportOptions) ([]ImportResult, error) {
	if opts.ProjectRoot == "" {
		return nil, fmt.Errorf("project root is required")
	}

	r, err := zip.OpenReader(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("open bundle: %w", err)
	}
	defer r.Close()

	// Read and validate manifest
	m, err := readManifestFromZip(r)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	// Build allowed skill names set
	allowedNames := make(map[string]bool, len(m.Skills))
	for _, s := range m.Skills {
		allowedNames[s.Name] = true
	}

	// Determine target directories
	targets := opts.Targets
	if len(targets) == 0 {
		targets = activeProvidersIn(opts.ProjectRoot)
		if len(targets) == 0 {
			targets = []string{filepath.Join(opts.ProjectRoot, ".agents", "skills")}
		}
	}

	// Pre-check existing skills for duplicate detection
	type skillState struct {
		exists     bool
		needDelete bool
	}
	skillStates := make(map[string]*skillState, len(m.Skills))
	for _, s := range m.Skills {
		skillStates[s.Name] = &skillState{}
	}
	for _, target := range targets {
		for _, s := range m.Skills {
			skillDir := filepath.Join(target, s.Name)
			if _, err := os.Stat(skillDir); err == nil {
				skillStates[s.Name].exists = true
			}
		}
	}

	// Stream zip entries to staging directory for atomic install
	stagingRoot, err := os.MkdirTemp(opts.ProjectRoot, ".bundle-import-")
	if err != nil {
		return nil, fmt.Errorf("create staging dir: %w", err)
	}
	stagingCleanup := true
	defer func() {
		if stagingCleanup {
			if err := os.RemoveAll(stagingRoot); err != nil {
				fmt.Fprintf(os.Stderr, "warning: staging cleanup: %v\n", err)
			}
		}
	}()

	for _, zf := range r.File {
		// Skip non-skill entries (e.g. manifest.json at root)
		if zf.Name == "manifest.json" || strings.HasPrefix(zf.Name, "manifest") {
			continue
		}

		// Skip ZIP directory entries
		if strings.HasSuffix(zf.Name, "/") || strings.HasSuffix(zf.Name, "\\") {
			continue
		}

		// Validate path safety
		skillName, relPath, err := safeZipPath(zf.Name, allowedNames)
		if err != nil {
			return nil, fmt.Errorf("unsafe entry %q: %w", zf.Name, err)
		}

		// Pre-check UncompressedSize64 before opening to prevent zip-bomb
		if zf.UncompressedSize64 > MaxEntryBytes {
			return nil, fmt.Errorf("zip entry %q uncompressed size %d exceeds max %d",
				zf.Name, zf.UncompressedSize64, MaxEntryBytes)
		}

		// Open, read bounded content, write to staging
		rc, err := zf.Open()
		if err != nil {
			return nil, fmt.Errorf("open zip entry %q: %w", zf.Name, err)
		}
		content, err := io.ReadAll(io.LimitReader(rc, MaxEntryBytes+1))
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("read zip entry %q: %w", zf.Name, err)
		}
		if len(content) > MaxEntryBytes {
			return nil, fmt.Errorf("zip entry %q exceeds max size %d bytes", zf.Name, MaxEntryBytes)
		}

		// Write to staging directory (relPath = "mytool/SKILL.md", subPath = "SKILL.md")
		subPath := strings.TrimPrefix(relPath, skillName+"/")
		stagePath := filepath.Join(stagingRoot, skillName, subPath)
		// Verify path stays within staging (defense-in-depth)
		if _, err := filepath.Rel(stagingRoot, stagePath); err != nil {
			return nil, fmt.Errorf("staging path containment: %w", err)
		}
		if err := os.MkdirAll(filepath.Dir(stagePath), 0755); err != nil {
			return nil, fmt.Errorf("staging mkdir: %w", err)
		}
		if err := os.WriteFile(stagePath, content, 0644); err != nil {
			return nil, fmt.Errorf("staging write: %w", err)
		}
	}

	// Build results
	var results []ImportResult

	for _, s := range m.Skills {
		state := skillStates[s.Name]
		result := ImportResult{Skill: s.Name}

		switch {
		case state.exists && opts.OnDuplicate == DuplicateSkip:
			result.Status = "skipped"
		default:
			if state.exists {
				result.Status = "overwritten"
			} else {
				result.Status = "installed"
			}

			stageSkill := filepath.Join(stagingRoot, s.Name)
			if _, statErr := os.Stat(stageSkill); os.IsNotExist(statErr) {
				// No files in staging for this skill; nothing to write
				result.Status = "skipped"
				results = append(results, result)
				continue
			}

			// Copy from staging to each target
			var failed bool
			for _, target := range targets {
				finalDir := filepath.Join(target, s.Name)
				// For overwrite: remove existing
				if state.exists {
					if err := os.RemoveAll(finalDir); err != nil {
						result.Status = "failed"
						result.Error = fmt.Errorf("remove existing %s: %w", finalDir, err)
						failed = true
						break
					}
				}
				// Copy from staging (still exists for next target)
				if err := copyDir(stageSkill, finalDir); err != nil {
					result.Status = "failed"
					result.Error = fmt.Errorf("install to %s: %w", finalDir, err)
					failed = true
					break
				}
			}
			if failed {
				results = append(results, result)
				continue
			}
		}

		results = append(results, result)
	}

	// Cleanup staging on success
	stagingCleanup = false
	if err := os.RemoveAll(stagingRoot); err != nil {
		fmt.Fprintf(os.Stderr, "warning: staging cleanup: %v\n", err)
	}

	// Call sync only if at least one skill was actually written
	hasChanges := false
	for _, r := range results {
		if r.Status == "installed" || r.Status == "overwritten" {
			hasChanges = true
			break
		}
	}
	if hasChanges {
		if syncErr := SyncAfterImport(opts.ProjectRoot); syncErr != nil {
			// Non-fatal: report as a warning result entry
			results = append(results, ImportResult{
				Skill:  "*sync",
				Status: "warning",
				Error:  fmt.Errorf("post-import sync: %w", syncErr),
			})
		}
	}

	return results, nil
}

// copyDir recursively copies src directory contents to dst. Both must exist.
// Returns the first error encountered, if any.
func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("mkdir dst: %w", err)
	}
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, 0644)
	})
}

// safeZipPath validates a zip entry path and returns the skill name and
// relative path within the skill directory. It rejects:
// - Entries not under the "skills/" prefix
// - Paths with ".." traversal components
// - Absolute paths (leading "/" or Windows drive letters)
// - Entries whose skill name is not in the allowed set
func safeZipPath(entryName string, allowedNames map[string]bool) (skillName, relPath string, err error) {
	// Reject NUL bytes and raw backslashes before any normalization
	if strings.ContainsRune(entryName, 0) {
		return "", "", fmt.Errorf("invalid path: contains null byte")
	}
	if strings.Contains(entryName, "\\") {
		return "", "", fmt.Errorf("invalid path: backslash not allowed")
	}

	// Normalize to forward slashes for consistent validation
	name := filepath.ToSlash(entryName)

	// Reject absolute paths
	if strings.HasPrefix(name, "/") {
		return "", "", fmt.Errorf("absolute path not allowed")
	}
	// Reject Windows drive letters (e.g., "C:...") including drive-relative
	if len(name) >= 2 && name[1] == ':' && ((name[0] >= 'a' && name[0] <= 'z') || (name[0] >= 'A' && name[0] <= 'Z')) {
		return "", "", fmt.Errorf("drive-relative or absolute path not allowed")
	}

	// Clean the path and check for traversal
	cleaned := filepath.ToSlash(filepath.Clean(name))

	// Reject if cleaning reveals traversal
	if cleaned != name && strings.HasPrefix(cleaned, "..") {
		return "", "", fmt.Errorf("path traversal detected")
	}
	if strings.Contains(cleaned, "..") {
		return "", "", fmt.Errorf("path traversal detected")
	}

	// Must be under skills/ prefix
	if !strings.HasPrefix(cleaned, "skills/") {
		return "", "", fmt.Errorf("outside skills directory")
	}

	// Extract skill name (first path component after "skills/")
	rest := strings.TrimPrefix(cleaned, "skills/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "", "", fmt.Errorf("empty skill name in path")
	}
	skillName = parts[0]

	// Validate skill name charset before checking declaration
	if !validSkillName.MatchString(skillName) {
		return "", "", fmt.Errorf("invalid skill name %q", skillName)
	}

	// Verify skill name is declared in manifest
	if !allowedNames[skillName] {
		return "", "", fmt.Errorf("undeclared skill name %q", skillName)
	}

	relPath = rest
	return skillName, relPath, nil
}

// readManifestFromZip reads and validates the manifest from an already-open zip.
func readManifestFromZip(r *zip.ReadCloser) (*Manifest, error) {
	for _, f := range r.File {
		if f.Name != "manifest.json" {
			continue
		}
		return ReadManifestFromFile(f)
	}
	return nil, fmt.Errorf("manifest.json not found in bundle")
}

// ReadManifestFromFile reads the manifest from a single zip file entry.
func ReadManifestFromFile(f *zip.File) (*Manifest, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("open manifest: %w", err)
	}
	defer rc.Close()

	m, err := readManifestFromReader(rc)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func readManifestFromReader(rc io.ReadCloser) (*Manifest, error) {
	data, err := io.ReadAll(io.LimitReader(rc, MaxManifestBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	if len(data) > MaxManifestBytes {
		return nil, fmt.Errorf("manifest exceeds max size %d bytes", MaxManifestBytes)
	}

	return parseManifest(data)
}

// parseManifest unmarshals and validates manifest JSON.
func parseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := unmarshalStrict(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if m.Version != ManifestVersion {
		return nil, fmt.Errorf("unsupported manifest version %d (supported: %d)", m.Version, ManifestVersion)
	}
	return &m, nil
}

// unmarshalStrict unmarshals JSON, rejecting unknown fields.
func unmarshalStrict(data []byte, v interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// activeProvidersIn returns existing provider skill directories under root.
func activeProvidersIn(root string) []string {
	providers := []string{
		filepath.Join(root, ".agents", "skills"),
		filepath.Join(root, ".opencode", "skills"),
		filepath.Join(root, ".claude", "skills"),
		filepath.Join(root, ".gemini", "skills"),
		filepath.Join(root, ".cursor", "skills"),
		filepath.Join(root, ".copilot", "skills"),
		filepath.Join(root, ".qwen", "skills"),
	}
	var active []string
	for _, p := range providers {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			active = append(active, p)
		}
	}
	return active
}
