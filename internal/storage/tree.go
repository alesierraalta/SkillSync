package storage

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// reservedFiles are storage-managed files at the root of a stored skill that
// are never treated as part of the skill's own file tree.
var reservedFiles = map[string]bool{
	"SKILL.md":      true,
	"METADATA.json": true,
}

// copyTree copies every file under srcDir into dstDir, preserving relative
// paths. Top-level reserved files (SKILL.md, METADATA.json) are skipped —
// SKILL.md is written separately in its normalized form.
func copyTree(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(srcDir, path)
		if rerr != nil {
			return rerr
		}
		if reservedFiles[filepath.ToSlash(rel)] {
			return nil
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

// CopyExtras copies a stored skill's supporting files (references/, assets/,
// anything beyond SKILL.md and METADATA.json) into dstDir.
func (s *Service) CopyExtras(skillID, dstDir string) error {
	srcDir := filepath.Join(s.RootPath, skillID)
	if _, err := os.Stat(srcDir); err != nil {
		return fmt.Errorf("stored skill %q: %w", skillID, err)
	}
	return copyTree(srcDir, dstDir)
}
