package discovery

import (
	"io/fs"
	"path/filepath"
)

// DiscoverSkills finds all SKILL.md files in the root path recursively.
func DiscoverSkills(root string) ([]string, error) {
	var skills []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Base(path) == "SKILL.md" {
			skills = append(skills, path)
		}
		return nil
	})
	return skills, err
}
