package root_test

import (
	"os"
	"strings"
	"testing"
)

// TestChangelog_2026_06_25_EntryExists verifies that CHANGELOG.md contains
// a 2026-06-25 entry with the required audit deliverables (AUDIT-10a, AUDIT-10b).
func TestChangelog_2026_06_25_EntryExists(t *testing.T) {
	raw, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	body := string(raw)

	// AUDIT-10a: Verify the date entry exists
	if !strings.Contains(body, "2026-06-25") {
		t.Error("CHANGELOG.md missing 2026-06-25 entry")
	}

	// AUDIT-10b: Verify the required items are mentioned
	requiredItems := []string{
		"audit", // audit pass / reports
		"slog", // structured logging
		"panic", // panic recovery
		"Makefile",
		"golangci", // golangci config
		"test_command",
	}
	for _, item := range requiredItems {
		if !strings.Contains(body, item) {
			t.Errorf("CHANGELOG.md 2026-06-25 entry missing required item %q", item)
		}
	}
}
