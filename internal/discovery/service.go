package discovery

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DiscoverSkills finds all SKILL.md files in known provider directories in the root path.
func DiscoverSkills(root string) ([]string, error) {
	var skills []string
	providers := []string{".claude", ".opencode", ".agents", ".gemini", ".cursor", ".copilot", ".qwen"}

	for _, provider := range providers {
		providerPath := filepath.Join(root, provider)
		if _, err := os.Stat(providerPath); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(providerPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// Use os.Stat to follow symlinks/junctions; this handles Windows junctions
			// where d.IsDir() would return false but the target is a directory
			info, err := os.Stat(path)
			if err != nil {
				return nil
			}
			if info.IsDir() {
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

// ScanProjects scans the provided roots for project markers.
// Depth 1 means only the root itself and its immediate children.
func ScanProjects(roots []string, maxDepth int) ([]string, error) {
	var projects []string
	excluded := map[string]bool{
		"node_modules": true,
		".git":         true,
		"vendor":       true,
		"dist":         true,
		"build":        true,
	}

	for _, root := range roots {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			continue
		}

		err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			rel, err := filepath.Rel(absRoot, path)
			if err != nil {
				return nil
			}

			// Calculate depth. filepath.Rel(".") returns "." which has 1 part.
			// "p1" returns "p1" (1 part). "nested/p4" returns "nested/p4" (2 parts).
			depth := 1
			if rel != "." {
				depth = len(strings.Split(rel, string(filepath.Separator)))
			}

			if depth > maxDepth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if d.IsDir() {
				name := d.Name()
				if excluded[name] {
					return filepath.SkipDir
				}
				// Skip hidden dirs unless they are markers
				if strings.HasPrefix(name, ".") && name != "." && name != ".." {
					if name != ".agents" && name != ".opencode" {
						return filepath.SkipDir
					}
				}

				// Check if current dir is a project
				if isProject(path) {
					projects = append(projects, path)
					return filepath.SkipDir
				}
			}

			return nil
		})
	}

	return projects, nil
}

func isProject(path string) bool {
	markers := []string{".agents", ".opencode", "AGENTS.md"}
	for _, m := range markers {
		if _, err := os.Stat(filepath.Join(path, m)); err == nil {
			return true
		}
	}
	return false
}
