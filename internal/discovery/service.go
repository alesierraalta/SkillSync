package discovery

import (
	"io/fs"
	"os"
	"path/filepath"
)

// DiscoverSkills finds all SKILL.md files in known provider directories in the root path.
func DiscoverSkills(root string) ([]string, error) {
	var skills []string
	providers := []string{".claude", ".opencode", ".agents", ".gemini", ".cursor", ".copilot"}

	for _, provider := range providers {
		providerPath := filepath.Join(root, provider)
		if _, err := os.Stat(providerPath); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(providerPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				skillFile := filepath.Join(path, "SKILL.md")
				if _, err := os.Stat(skillFile); err == nil {
					skills = append(skills, skillFile)
					return filepath.SkipDir
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return skills, nil
}


