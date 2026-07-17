// Package providersync mirrors the canonical .agents/skills tree into an
// arbitrary provider directory (e.g. .claude/skills, .gemini/skills), copying
// each skill's full file tree (SKILL.md plus references/assets/subfolders).
//
// OpenCode has its own richer sync (tools/agents/commands regeneration) in
// internal/opencode; this package covers the plain "copy the skills" mirror
// used for every other provider selected in the sync screen.
package providersync

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Mirror copies every skill under <root>/.agents/skills into
// <root>/<providerDir>/skills, preserving each skill's subfolders. It returns
// the number of skills mirrored. A missing source dir yields (0, nil).
func Mirror(root, providerDir string) (int, error) {
	srcRoot := filepath.Join(root, ".agents", "skills")
	dstRoot := filepath.Join(root, providerDir, "skills")

	entries, err := os.ReadDir(srcRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		srcDir := filepath.Join(srcRoot, name)
		// Follow junctions/symlinks and require a SKILL.md to qualify.
		if info, err := os.Stat(srcDir); err != nil || !info.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(srcDir, "SKILL.md")); err != nil {
			continue
		}
		dstDir := filepath.Join(dstRoot, name)
		if err := copyTree(srcDir, dstDir); err != nil {
			return count, fmt.Errorf("mirror %s: %w", name, err)
		}
		count++
	}
	return count, nil
}

// copyTree copies every file under srcDir into dstDir, preserving relative
// paths and creating directories as needed.
func copyTree(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, serr := os.Stat(path) // follow symlinks/junctions
		if serr != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(srcDir, path)
		if rerr != nil {
			return rerr
		}
		dst := filepath.Join(dstDir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
		return copyFile(path, dst)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
