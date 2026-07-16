package ui

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// listSkillFiles walks a skill directory and returns every file as a
// slash-separated path relative to dir, sorted alphabetically. This lets the
// browser surface supporting files in subfolders (references/, assets/, ...)
// alongside SKILL.md.
func listSkillFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(dir, path)
		if rerr != nil {
			return rerr
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// openSkillFileBrowser loads the file list for the selected skill and
// switches to the file browser screen.
func (m Model) openSkillFileBrowser() (tea.Model, tea.Cmd) {
	if m.selected == nil {
		return m, nil
	}
	dir := filepath.Dir(m.selected.Path)
	files, err := listSkillFiles(dir)
	if err != nil {
		m.List.viewport.SetContent(fmt.Sprintf("Error listing files: %v", err))
		return m, nil
	}
	m.skillDir = dir
	m.skillFiles = files
	m.skillFilesCursor = 0
	// Remember the screen that launched the content view so esc from the
	// browser can restore the original back-navigation target.
	m.skillFilesOrigin = m.PrevScreen
	m.Screen = ScreenSkillFiles
	return m, nil
}

func (m Model) handleSkillFilesKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.Screen = ScreenContentView
		m.PrevScreen = m.skillFilesOrigin
		return m, nil
	case "j", "down":
		if m.skillFilesCursor < len(m.skillFiles)-1 {
			m.skillFilesCursor++
		}
	case "k", "up":
		if m.skillFilesCursor > 0 {
			m.skillFilesCursor--
		}
	case "enter":
		if len(m.skillFiles) == 0 {
			return m, nil
		}
		rel := m.skillFiles[m.skillFilesCursor]
		path := filepath.Join(m.skillDir, filepath.FromSlash(rel))
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			m.List.viewport.SetContent(fmt.Sprintf("Error reading file: %v", err))
		} else {
			m.List.viewport.SetContent(string(contentBytes))
		}
		m.List.viewport.GotoTop()
		m.PrevScreen = ScreenSkillFiles
		m.Screen = ScreenContentView
		return m, nil
	}
	return m, nil
}

func (m Model) viewSkillFiles() string {
	if m.selected == nil {
		return "No skill selected"
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Files: %s", m.selected.Name)) + "\n\n")
	b.WriteString(hintStyle.Render("enter: view · j/k: navigate · esc: back") + "\n\n")

	start, end := windowBounds(m.skillFilesCursor, len(m.skillFiles), m.Height-8)
	for i := start; i < end; i++ {
		cursor := "  "
		if i == m.skillFilesCursor {
			cursor = "> "
		}
		b.WriteString(cursor + m.skillFiles[i] + "\n")
	}
	if len(m.skillFiles) == 0 {
		b.WriteString("(no files)\n")
	}
	return docStyle.Render(b.String())
}
