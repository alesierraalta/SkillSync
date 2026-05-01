package opencode

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Options controls sync behavior.
type Options struct {
	// Prune deletes skills in .opencode/skills/ that have no corresponding
	// source in .agents/skills/ when true.
	Prune bool
	// DryRun reports planned changes without writing anything when true.
	DryRun bool
}

// SyncSkills mirrors .agents/skills/ → .opencode/skills/.
// It copies each SKILL.md file, preserves symlinks/junctions as symlinks,
// skips writing when destination content matches source (content-addressed),
// and optionally prunes orphans.
func SyncSkills(root string, opts Options) error {
	srcRoot := filepath.Join(root, ".agents", "skills")
	dstRoot := filepath.Join(root, ".opencode", "skills")

	// Collect all source skill paths
	var srcPaths []string
	err := filepath.WalkDir(srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		lstatInfo, err := os.Lstat(path)
		if err != nil {
			return nil
		}
		// If it's a symlink, check the target
		if lstatInfo.Mode()&os.ModeSymlink != 0 {
			// Symlink: check if target is a directory with SKILL.md
			evalPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				return nil
			}
			info, err := os.Stat(evalPath)
			if err != nil {
				return nil
			}
			if info.IsDir() {
				skillFile := filepath.Join(path, "SKILL.md")
				if _, err := os.Stat(skillFile); err == nil {
					srcPaths = append(srcPaths, skillFile)
					return filepath.SkipDir
				}
			}
			// Symlink to a file - check if path itself is a SKILL.md
			if filepath.Base(path) == "SKILL.md" {
				srcPaths = append(srcPaths, path)
			}
			return nil
		}
		// Regular directory
		if d.IsDir() {
			skillFile := filepath.Join(path, "SKILL.md")
			if _, err := os.Stat(skillFile); err == nil {
				srcPaths = append(srcPaths, skillFile)
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("walk .agents/skills: %w", err)
	}

	// Collect existing destination skills (for prune)
	var dstOrphans map[string]bool
	if opts.Prune {
		dstOrphans = make(map[string]bool)
		_ = filepath.WalkDir(dstRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			// Skip the root itself; only consider direct children
			if path == dstRoot {
				return nil
			}
			info, err := os.Stat(path)
			if err != nil || !info.IsDir() {
				return nil
			}
			// This is a directory - check if it's a skill directory
			skillFile := filepath.Join(path, "SKILL.md")
			if _, err := os.Stat(skillFile); err == nil {
				name := filepath.Base(path)
				dstOrphans[name] = true
				return filepath.SkipDir
			}
			return nil
		})
	}

	// Mirror each source skill
	for _, srcPath := range srcPaths {
		rel, err := filepath.Rel(srcRoot, srcPath)
		if err != nil {
			continue
		}
		dstPath := filepath.Join(dstRoot, rel)

		// Determine if source is a symlink
		srcInfo, err := os.Lstat(srcPath)
		if err != nil {
			return fmt.Errorf("lstat source %s: %w", srcPath, err)
		}

		if opts.DryRun {
			fmt.Printf("[dry-run] would mirror: %s → %s\n", srcPath, dstPath)
			// Mark as not orphaned
			if opts.Prune {
				name := filepath.Base(filepath.Dir(dstPath))
				delete(dstOrphans, name)
			}
			continue
		}

		// Ensure destination directory exists
		dstDir := filepath.Dir(dstPath)
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dstDir, err)
		}

		if srcInfo.Mode()&os.ModeSymlink != 0 {
			// Symlink: recreate as symlink
			linkTarget, err := os.Readlink(srcPath)
			if err != nil {
				return fmt.Errorf("readlink %s: %w", srcPath, err)
			}
			// Remove existing file/symlink if present
			os.Remove(dstPath)
			if err := os.Symlink(linkTarget, dstPath); err != nil {
				return fmt.Errorf("symlink %s → %s: %w", dstPath, linkTarget, err)
			}
		} else {
			// Regular file: content-addressed copy
			srcContent, err := os.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("read source %s: %w", srcPath, err)
			}

			// Skip if dest already has identical content
			dstContent, err := os.ReadFile(dstPath)
			if err == nil && string(dstContent) == string(srcContent) {
				// Content unchanged, skip write
			} else {
				if err := os.WriteFile(dstPath, srcContent, 0644); err != nil {
					return fmt.Errorf("write %s: %w", dstPath, err)
				}
			}
		}

		// Mark as not orphaned
		if opts.Prune {
			name := filepath.Base(filepath.Dir(dstPath))
			delete(dstOrphans, name)
		}
	}

	// Prune orphans
	if opts.Prune && !opts.DryRun {
		for name := range dstOrphans {
			orphanPath := filepath.Join(dstRoot, name)
			if err := os.RemoveAll(orphanPath); err != nil {
				return fmt.Errorf("prune %s: %w", orphanPath, err)
			}
		}
	} else if opts.Prune && opts.DryRun {
		for name := range dstOrphans {
			fmt.Printf("[dry-run] would prune: %s\n", name)
		}
	}

	return nil
}

// MirrorSkills returns the list of skill names that were mirrored.
func MirrorSkills(root string) ([]string, error) {
	srcRoot := filepath.Join(root, ".agents", "skills")
	var names []string
	err := filepath.WalkDir(srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil
		}
		if info.IsDir() {
			skillFile := filepath.Join(path, "SKILL.md")
			if _, err := os.Stat(skillFile); err == nil {
				names = append(names, filepath.Base(path))
				return filepath.SkipDir
			}
		}
		return nil
	})
	return names, err
}

// IsSymlink returns true if the given path is a symlink or Windows junction.
func IsSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// SkipReason describes why a file was skipped during sync.
type SkipReason int

const (
	SkipNone SkipReason = iota
	SkipUnchanged
	SkipSymlink
)

// ShouldSkip returns SkipUnchanged if dstContent == srcContent, SkipSymlink
// if the source is a symlink, or SkipNone if the file should be written.
func ShouldSkip(srcPath, dstPath string) (SkipReason, error) {
	srcInfo, err := os.Lstat(srcPath)
	if err != nil {
		return SkipNone, err
	}
	if srcInfo.Mode()&os.ModeSymlink != 0 {
		return SkipSymlink, nil
	}
	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		return SkipNone, err
	}
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		return SkipNone, nil
	}
	if string(dstContent) == string(srcContent) {
		return SkipUnchanged, nil
	}
	return SkipNone, nil
}

// skillNameFromPath extracts the skill name from a path like
// /root/.agents/skills/foo/SKILL.md → "foo".
func skillNameFromPath(path string) string {
	dir := filepath.Dir(path)
	name := filepath.Base(dir)
	// If under .agents/skills/, go up one more level
	if strings.HasSuffix(dir, filepath.Join(".agents", "skills")) {
		return filepath.Base(dir)
	}
	return name
}
