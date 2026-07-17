package ui

import (
	"strings"
	"testing"

	"skillsync/tui/internal/storage"
)

func TestSkillsLoadingClearsOnLoaded(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService(t.TempDir())))
	if !m.skillsLoading {
		t.Fatal("expected skillsLoading true initially")
	}

	next, _ := m.Update(skillsLoadedMsg(nil))
	nm := next.(Model)
	if nm.skillsLoading {
		t.Errorf("expected skillsLoading false after skillsLoadedMsg")
	}
}

func TestLoadingViewShowsMessage(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService(t.TempDir())))
	m.skillsLoading = true
	v := m.loadingView()
	if !strings.Contains(v, "Cargando skills") {
		t.Errorf("expected loading message, got:\n%s", v)
	}
}

func TestBeginLoadSkillsSetsFlag(t *testing.T) {
	m := NewModel(NewBackend(storage.NewService(t.TempDir())))
	m.skillsLoading = false
	cmd := m.beginLoadSkills()
	if !m.skillsLoading {
		t.Errorf("beginLoadSkills must set skillsLoading true")
	}
	if cmd == nil {
		t.Errorf("beginLoadSkills must return a command")
	}
}
