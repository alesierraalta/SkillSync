package coreskills

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestCoreSkillDrift(t *testing.T) {
	coreSkills := []string{"find-skills", "skill-creator", "skill-sync"}

	// Find project root
	root := "../.."
	if _, err := os.Stat(filepath.Join(root, ".agents", "skills")); err != nil {
		// Fallback for different test execution contexts
		root = "."
		for i := 0; i < 5; i++ {
			if _, err := os.Stat(filepath.Join(root, ".agents", "skills")); err == nil {
				break
			}
			root = filepath.Join("..", root)
		}
	}

	for _, skill := range coreSkills {
		t.Run(skill, func(t *testing.T) {
			embeddedSub, err := fs.Sub(EmbeddedSkills, "skills/"+skill)
			if err != nil {
				t.Fatalf("failed to get sub-fs for %s: %v", skill, err)
			}

			localPath := filepath.Join(root, ".agents", "skills", skill)
			if _, err := os.Stat(localPath); err != nil {
				t.Logf("skipping drift check for %s: local directory not found at %s", skill, localPath)
				return
			}

			// 1. Walk embedded and check against local
			err = fs.WalkDir(embeddedSub, ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if path == "." {
					return nil
				}

				localFile := filepath.Join(localPath, path)
				info, err := os.Stat(localFile)
				if err != nil {
					t.Errorf("missing local file: %s (at %s)", path, localFile)
					return nil
				}

				if d.IsDir() {
					if !info.IsDir() {
						t.Errorf("mismatch: %s is a directory in embedded but a file in local", path)
					}
					return nil
				}

				if info.IsDir() {
					t.Errorf("mismatch: %s is a file in embedded but a directory in local", path)
					return nil
				}

				// Compare content
				embFile, err := embeddedSub.Open(path)
				if err != nil {
					t.Errorf("failed to open embedded file %s: %v", path, err)
					return nil
				}
				defer embFile.Close()

				embData, _ := io.ReadAll(embFile)
				localData, err := os.ReadFile(localFile)
				if err != nil {
					t.Errorf("failed to read local file %s: %v", localFile, err)
					return nil
				}

				if !bytes.Equal(embData, localData) {
					t.Errorf("DRIFT DETECTED in %s/%s: embedded content does not match repository source", skill, path)
				}

				return nil
			})
			if err != nil {
				t.Fatalf("error walking embedded skills: %v", err)
			}

			// 2. Walk local and check against embedded (detect extra files)
			err = filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				rel, _ := filepath.Rel(localPath, path)
				if rel == "." {
					return nil
				}

				_, err = fs.Stat(embeddedSub, filepath.ToSlash(rel))
				if err != nil {
					t.Errorf("extra local file not found in embedded: %s/%s", skill, rel)
				}
				return nil
			})
			if err != nil {
				t.Fatalf("error walking local skills: %v", err)
			}
		})
	}
}
