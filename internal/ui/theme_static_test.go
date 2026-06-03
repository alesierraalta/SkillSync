package ui

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestNoRawColorLiteralsOutsideTheme enforces the design-token layer boundary:
// no file in internal/ui (except theme.go and test files) may contain a raw
// lipgloss.Color("NNN") literal. All color values must flow from DefaultTheme.
// Spec: THEME-2.
func TestNoRawColorLiteralsOutsideTheme(t *testing.T) {
	re := regexp.MustCompile(`lipgloss\.Color\("[0-9]+"\)`)
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		n := e.Name()
		if !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") || n == "theme.go" {
			continue
		}
		src, err := os.ReadFile(n)
		if err != nil {
			t.Fatal(err)
		}
		if loc := re.FindIndex(src); loc != nil {
			t.Errorf("%s contains raw color literal %q — move to theme.go/DefaultTheme", n, src[loc[0]:loc[1]])
		}
	}
}

// TestDefaultThemeTokens verifies that all 9 DefaultTheme fields are accessible
// and that the non-empty ones are non-empty.
// Spec: THEME-1.
func TestDefaultThemeTokens(t *testing.T) {
	tokens := []struct {
		name  string
		value string
		empty bool
	}{
		{"Primary", string(DefaultTheme.Primary), false},
		{"Secondary", string(DefaultTheme.Secondary), false},
		{"Accent", string(DefaultTheme.Accent), false},
		{"Success", string(DefaultTheme.Success), false},
		{"Warning", string(DefaultTheme.Warning), false},
		{"Error", string(DefaultTheme.Error), false},
		{"Muted", string(DefaultTheme.Muted), false},
		{"Text", string(DefaultTheme.Text), true},
		{"SyncOutput", string(DefaultTheme.SyncOutput), false},
	}
	for _, tok := range tokens {
		if tok.empty {
			if tok.value != "" {
				t.Errorf("DefaultTheme.%s: expected empty string (terminal default), got %q", tok.name, tok.value)
			}
		} else {
			if tok.value == "" {
				t.Errorf("DefaultTheme.%s: expected non-empty color value, got empty string", tok.name)
			}
		}
	}
}
