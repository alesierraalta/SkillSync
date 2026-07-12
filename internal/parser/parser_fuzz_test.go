package parser_test

import (
	"testing"

	"skillsync/tui/internal/parser"
)

// FuzzParseContent fuzzes the ParseContent function with varied SKILL.md content.
// Run with: go test -fuzz=FuzzParseContent -fuzztime=1s ./internal/parser/...
func FuzzParseContent(f *testing.F) {
	// Seed corpus — representative inputs covering major code paths.
	f.Add("---\nname: skill\ndescription: desc\n---\n# Body\n")
	f.Add("# No frontmatter\n\nJust a body.")
	f.Add("---\n---\n")
	f.Add("")

	f.Fuzz(func(t *testing.T, content string) {
		// Must not panic; errors are acceptable.
		_, _ = parser.ParseContent(content)
	})
}
