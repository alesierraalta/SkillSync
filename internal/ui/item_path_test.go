package ui

import (
	"strings"
	"testing"

	"skillsync/tui/internal/types"
)

func TestManageItemDescriptionShowsFullPath(t *testing.T) {
	it := item{skill: types.Skill{
		Name: "bash-defensive-patterns",
		Path: "/home/u/.claude/skills/bash-defensive-patterns/SKILL.md",
		Metadata: types.Metadata{
			Description: "Master defensive Bash programming techniques.",
		},
	}}

	desc := it.Description()

	if !strings.Contains(desc, "Path: /home/u/.claude/skills/bash-defensive-patterns/SKILL.md") {
		t.Errorf("expected full path in description, got:\n%s", desc)
	}
	if !strings.Contains(desc, "Master defensive Bash") {
		t.Errorf("expected description text preserved, got:\n%s", desc)
	}
}

func TestManageItemDescriptionNoPathWhenEmpty(t *testing.T) {
	it := item{skill: types.Skill{Name: "virtual", Path: ""}}
	desc := it.Description()
	if strings.Contains(desc, "Path:") {
		t.Errorf("did not expect a Path line when path is empty, got:\n%s", desc)
	}
}
