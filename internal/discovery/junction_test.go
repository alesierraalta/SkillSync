package discovery

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// makeLink creates a directory junction (Windows, no admin needed) or a
// symlink (Unix) at linkPath pointing to target. Returns false if the
// platform refuses (e.g. symlink without privilege), so the test can skip.
func makeLink(t *testing.T, target, linkPath string) bool {
	t.Helper()
	if runtime.GOOS == "windows" {
		// Directory junctions do not require elevation, unlike symlinks.
		cmd := exec.Command("cmd", "/c", "mklink", "/J", linkPath, target)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Logf("mklink /J failed: %v (%s)", err, out)
			return false
		}
		return true
	}
	if err := os.Symlink(target, linkPath); err != nil {
		t.Logf("symlink failed: %v", err)
		return false
	}
	return true
}

// TestDiscoverSkillsJunctionDoesNotTruncateSiblings reproduces the bug where a
// junction/symlink skill dir sorted before a real skill dir caused WalkDir to
// skip all remaining siblings (returning SkipDir on a non-directory DirEntry),
// so real skills after it were silently lost.
func TestDiscoverSkillsJunctionDoesNotTruncateSiblings(t *testing.T) {
	tmp := t.TempDir()
	skillsDir := filepath.Join(tmp, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Real skill that sorts AFTER the junction ("m..." > "b...").
	realAfter := filepath.Join(skillsDir, "zzz-real-skill")
	if err := os.MkdirAll(realAfter, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realAfter, "SKILL.md"), []byte("# real"), 0644); err != nil {
		t.Fatal(err)
	}

	// Junction skill that sorts BEFORE the real one.
	linkTarget := filepath.Join(tmp, "link-target")
	if err := os.MkdirAll(linkTarget, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(linkTarget, "SKILL.md"), []byte("# linked"), 0644); err != nil {
		t.Fatal(err)
	}
	linkPath := filepath.Join(skillsDir, "aaa-junction-skill")
	if !makeLink(t, linkTarget, linkPath) {
		t.Skip("cannot create junction/symlink on this platform without privilege")
	}

	found, err := DiscoverSkills(tmp)
	if err != nil {
		t.Fatalf("DiscoverSkills failed: %v", err)
	}

	var foundReal, foundJunction bool
	for _, f := range found {
		fs := strings.ReplaceAll(f, string(filepath.Separator), "/")
		if strings.Contains(fs, "/zzz-real-skill/SKILL.md") {
			foundReal = true
		}
		if strings.Contains(fs, "/aaa-junction-skill/SKILL.md") {
			foundJunction = true
		}
	}

	if !foundJunction {
		t.Error("junction skill (sorted first) not found")
	}
	if !foundReal {
		t.Error("real skill sorted AFTER a junction was truncated — SkipDir-on-symlink bug")
	}
}
