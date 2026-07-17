package ui

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCategoryRootLabelOpenCodeShowsBothRoots(t *testing.T) {
	home := filepath.Join("C:", "Users", "test")
	label := categoryRootLabel("OpenCode", home)

	if !strings.Contains(label, filepath.Join(home, ".opencode", "skills")) {
		t.Errorf("expected .opencode/skills in label, got %q", label)
	}
	if !strings.Contains(label, filepath.Join(home, ".config", "opencode", "skills")) {
		t.Errorf("expected .config/opencode/skills in label, got %q", label)
	}
}

func TestCategoryRootLabelSingleRoot(t *testing.T) {
	home := filepath.Join("C:", "Users", "test")
	label := categoryRootLabel("Claude", home)

	want := filepath.Join(home, ".claude", "skills")
	if label != want {
		t.Errorf("expected %q, got %q", want, label)
	}
}

func TestCategoryRootLabelAll(t *testing.T) {
	home := filepath.Join("C:", "Users", "test")
	label := categoryRootLabel("All", home)

	want := filepath.Join(home, ".*", "skills")
	if label != want {
		t.Errorf("expected %q, got %q", want, label)
	}
}
