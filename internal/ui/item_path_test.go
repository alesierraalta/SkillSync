package ui

import (
	"path/filepath"
	"strings"
	"testing"

	"skillsync/tui/internal/types"
)

func TestManageItemDescriptionShowsAbsolutePath(t *testing.T) {
	it := item{skill: types.Skill{
		Name: "bash-defensive-patterns",
		Path: "/home/u/.claude/skills/bash-defensive-patterns/SKILL.md",
		Metadata: types.Metadata{
			Description: "Master defensive Bash programming techniques.",
		},
	}}

	desc := it.Description()

	if !strings.Contains(desc, "Path:") {
		t.Errorf("expected a Path line, got:\n%s", desc)
	}
	if !strings.Contains(desc, "(proyecto actual)") {
		t.Errorf("expected '(proyecto actual)' marker, got:\n%s", desc)
	}
	if !strings.Contains(desc, "Master defensive Bash") {
		t.Errorf("expected description text preserved, got:\n%s", desc)
	}
}

func TestManageItemDescriptionRelativePathBecomesAbsolute(t *testing.T) {
	it := item{skill: types.Skill{
		Name: "x",
		Path: filepath.Join(".agents", "skills", "x", "SKILL.md"),
	}}

	desc := it.Description()

	// Extract the path segment from the "Path: <p> (proyecto actual)" line.
	line := ""
	for _, l := range strings.Split(desc, "\n") {
		if strings.HasPrefix(l, "Path: ") {
			line = l
			break
		}
	}
	if line == "" {
		t.Fatalf("no Path line in:\n%s", desc)
	}
	p := strings.TrimSuffix(strings.TrimPrefix(line, "Path: "), " (proyecto actual)")
	if !filepath.IsAbs(p) {
		t.Errorf("expected absolute path, got %q", p)
	}
}

func TestManageItemDescriptionNoPathWhenEmpty(t *testing.T) {
	it := item{skill: types.Skill{Name: "virtual", Path: ""}}
	desc := it.Description()
	if strings.Contains(desc, "Path:") {
		t.Errorf("did not expect a Path line when path is empty, got:\n%s", desc)
	}
}
