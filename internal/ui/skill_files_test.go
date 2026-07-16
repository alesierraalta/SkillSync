package ui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/types"
)

// writeSkillTree creates a skill dir with SKILL.md plus nested reference files.
func writeSkillTree(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"SKILL.md":               "# My Skill",
		"references/palette.md":  "palette content",
		"assets/sync.sh":         "#!/bin/sh",
		"references/deep/notes.md": "deep notes",
	}
	for rel, content := range files {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestListSkillFiles(t *testing.T) {
	dir := writeSkillTree(t)

	files, err := listSkillFiles(dir)
	if err != nil {
		t.Fatalf("listSkillFiles failed: %v", err)
	}

	want := []string{
		"SKILL.md",
		"assets/sync.sh",
		"references/deep/notes.md",
		"references/palette.md",
	}
	if len(files) != len(want) {
		t.Fatalf("expected %d files, got %d: %v", len(want), len(files), files)
	}
	for i, w := range want {
		if files[i] != w {
			t.Errorf("files[%d] = %q, want %q", i, files[i], w)
		}
	}
}

func TestContentViewFKeyOpensFileBrowser(t *testing.T) {
	dir := writeSkillTree(t)
	skillPath := filepath.Join(dir, "SKILL.md")

	m := NewModel(NewBackend(storage.NewService(t.TempDir())))
	m.selected = &types.Skill{Name: "my-skill", Path: skillPath}
	m.Screen = ScreenContentView

	next, _ := m.handleContentViewKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	nm := next.(Model)

	if nm.Screen != ScreenSkillFiles {
		t.Fatalf("expected ScreenSkillFiles, got %v", nm.Screen)
	}
	if len(nm.skillFiles) != 4 {
		t.Fatalf("expected 4 files, got %d: %v", len(nm.skillFiles), nm.skillFiles)
	}
}

func TestSkillFilesEnterViewsFile(t *testing.T) {
	dir := writeSkillTree(t)
	skillPath := filepath.Join(dir, "SKILL.md")

	m := NewModel(NewBackend(storage.NewService(t.TempDir())))
	m.selected = &types.Skill{Name: "my-skill", Path: skillPath}
	m.Screen = ScreenContentView
	next, _ := m.handleContentViewKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	nm := next.(Model)

	// Move cursor to "references/palette.md" (index 3 in sorted order)
	nm.skillFilesCursor = 3
	next2, _ := nm.handleSkillFilesKeys(tea.KeyMsg{Type: tea.KeyEnter})
	nm2 := next2.(Model)

	if nm2.Screen != ScreenContentView {
		t.Fatalf("expected ScreenContentView, got %v", nm2.Screen)
	}
	if nm2.PrevScreen != ScreenSkillFiles {
		t.Fatalf("expected PrevScreen ScreenSkillFiles, got %v", nm2.PrevScreen)
	}
}

func TestSkillFilesEscReturnsToContent(t *testing.T) {
	dir := writeSkillTree(t)
	skillPath := filepath.Join(dir, "SKILL.md")

	m := NewModel(NewBackend(storage.NewService(t.TempDir())))
	m.selected = &types.Skill{Name: "my-skill", Path: skillPath}
	m.Screen = ScreenContentView
	next, _ := m.handleContentViewKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	nm := next.(Model)

	next2, _ := nm.handleSkillFilesKeys(tea.KeyMsg{Type: tea.KeyEsc})
	nm2 := next2.(Model)

	if nm2.Screen != ScreenContentView {
		t.Fatalf("expected ScreenContentView after esc, got %v", nm2.Screen)
	}
}
