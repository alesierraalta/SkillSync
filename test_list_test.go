package root_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/muesli/reflow/wordwrap"
)

// testListItem is a minimal list.Item that mirrors the descriptor shape used
// by internal/ui items: a title, a (potentially long) description, and a path
// rendered in the multi-line delegate.
type testListItem struct {
	title, desc, path string
}

func (i testListItem) Title() string { return i.title }
func (i testListItem) Description() string {
	return fmt.Sprintf("Path: %s\n%s", i.path, wordwrap.String(i.desc, 70))
}
func (i testListItem) FilterValue() string { return i.title }

// TestBubbleteaList_RendersWordWrappedDescription is a structural smoke test
// for the bubbles list component with a multi-line delegate. It exercises
// both a short and an unusually long description (the M5Stack onboarding
// prompt) to confirm the word-wrapped Path/description delegate still
// produces a non-empty view that contains the expected fields. This guards
// against regressions where a delegate change would clip the path or fail to
// render wrapped long descriptions.
func TestBubbleteaList_RendersWordWrappedDescription(t *testing.T) {
	items := []list.Item{
		testListItem{title: "Test 1", desc: "Short", path: "/my/path"},
		testListItem{
			title: "m5-onboard",
			desc:  "End-to-end onboarding for a freshly-plugged-in M5Stack ESP32 device (Cardputer, Cardputer-Adv, Core, CoreS3, Stick) — detect on USB, flash UIFlow 2.0 and install the Claude Buddy MicroPython app",
			path:  "/some/very/long/path/here",
		},
	}

	d := list.NewDefaultDelegate()
	d.SetHeight(5) // Title + 4 lines of description
	l := list.New(items, d, 80, 20)
	view := l.View()

	if view == "" {
		t.Fatal("list.View() returned empty string")
	}
	for _, want := range []string{"Test 1", "m5-onboard", "/my/path", "/some/very/long/path/here"} {
		if !strings.Contains(view, want) {
			t.Errorf("list view missing %q", want)
		}
	}
}
