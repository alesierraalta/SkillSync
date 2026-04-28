package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const skillWithComments = `---
# Scope of the skill
scope: project
# Should be invoked automatically
auto_invoke: true
# Human readable name
name: My Skill
---
# Content
`

func TestParseAndSave(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "SKILL.md")
	
	err := os.WriteFile(path, []byte(skillWithComments), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Parse
	skill, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if skill.Metadata.Scope != "project" {
		t.Errorf("Expected scope 'project', got '%s'", skill.Metadata.Scope)
	}

	// Modify
	skill.Metadata.Scope = "personal"
	
	// Save
	err = Save(path, skill)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Read back raw
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	
	content := string(raw)
	
	// Verify comments preserved
	if !strings.Contains(content, "# Scope of the skill") {
		t.Error("Comment '# Scope of the skill' lost")
	}
	if !strings.Contains(content, "# Should be invoked automatically") {
		t.Error("Comment '# Should be invoked automatically' lost")
	}
	
	// Verify modification
	if !strings.Contains(content, "scope: personal") {
		t.Error("Modification 'scope: personal' not found in file")
	}

	// Verify content after frontmatter preserved
	if !strings.Contains(content, "# Content") {
		t.Error("Markdown content lost")
	}

	// Verify round-trip doesn't duplicate frontmatter
	if strings.Count(content, "---") != 2 {
		t.Errorf("Expected 2 frontmatter delimiters, found %d", strings.Count(content, "---"))
	}
}

func TestParseTrimsWhitespace(t *testing.T) {
	content := "---" + "\n" +
		"name: Test" + "\n" +
		"---" + "\n\n" +
		"   # Body with whitespace   " + "\n\n"
	
	tmp := t.TempDir()
	path := filepath.Join(tmp, "SKILL.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	skill, err := Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	expectedBody := "# Body with whitespace"
	if skill.RawBody != expectedBody {
		t.Errorf("Expected raw body %q, got %q", expectedBody, skill.RawBody)
	}
}

func TestParseWithPrefix(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedPrefix string
		expectedBody   string
		expectMeta     bool
	}{
		{
			name: "With Header Prefix",
			content: `# My Skill
---
description: Test
---
# Body content`,
			expectedPrefix: "# My Skill\n",
			expectedBody:   "# Body content",
			expectMeta:     true,
		},
		{
			name: "No Prefix",
			content: `---
description: Test
---
# Body content`,
			expectedPrefix: "",
			expectedBody:   "# Body content",
			expectMeta:     true,
		},
		{
			name: "No Frontmatter",
			content: `# Just body
No delimiters here`,
			expectedPrefix: "",
			expectedBody:   "# Just body\nNo delimiters here",
			expectMeta:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			path := filepath.Join(tmp, "SKILL.md")
			os.WriteFile(path, []byte(tt.content), 0644)

			skill, err := Parse(path)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if skill.Prefix != tt.expectedPrefix {
				t.Errorf("Prefix = %q, want %q", skill.Prefix, tt.expectedPrefix)
			}

			if skill.RawBody != tt.expectedBody {
				t.Errorf("RawBody = %q, want %q", skill.RawBody, tt.expectedBody)
			}
		})
	}
}

func TestSaveWithPrefixIdempotency(t *testing.T) {
	content := `# Header
---
description: Test
---
# Body`
	
	tmp := t.TempDir()
	path := filepath.Join(tmp, "SKILL.md")
	os.WriteFile(path, []byte(content), 0644)

	// Round-trip 1
	skill, _ := Parse(path)
	err := Save(path, skill)
	if err != nil {
		t.Fatalf("Save 1 failed: %v", err)
	}
	
	// Read back raw
	raw, _ := os.ReadFile(path)
	interim := string(raw)

	// Round-trip 2
	skill2, _ := Parse(path)
	err = Save(path, skill2)
	if err != nil {
		t.Fatalf("Save 2 failed: %v", err)
	}

	raw2, _ := os.ReadFile(path)
	final := string(raw2)

	if final != interim {
		t.Errorf("Expected idempotency after first save, got:\n%q\nwant:\n%q", final, interim)
	}

	if !strings.HasPrefix(final, "# Header\n---") {
		t.Errorf("Prefix lost or corrupted, got: %q", final)
	}
}


func TestParseModernMetadata(t *testing.T) {
	content := `---
name: modern-skill
description: Modern description
metadata:
  scope: [root]
  auto_invoke:
    - Test action
---
# Body`
	
	tmp := t.TempDir()
	path := filepath.Join(tmp, "SKILL.md")
	os.WriteFile(path, []byte(content), 0644)

	skill, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if skill.Metadata.Description != "Modern description" {
		t.Errorf("Expected description 'Modern description', got %q", skill.Metadata.Description)
	}

	if skill.Metadata.Scope != "root" {
		t.Errorf("Expected scope 'root', got %q", skill.Metadata.Scope)
	}
}

func TestParseFlexibleScope(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expScope string
	}{
		{
			name: "Nested List",
			yaml: "---\nmetadata:\n  scope: [ui, root]\n---",
			expScope: "ui, root",
		},
		{
			name: "Nested String",
			yaml: "---\nmetadata:\n  scope: api\n---",
			expScope: "api",
		},
		{
			name: "Top Level List",
			yaml: "---\nscope: [common, infra]\n---",
			expScope: "common, infra",
		},
		{
			name: "Top Level String",
			yaml: "---\nscope: root\n---",
			expScope: "root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			path := filepath.Join(tmp, "SKILL.md")
			os.WriteFile(path, []byte(tt.yaml+"\n# Body"), 0644)

			skill, err := Parse(path)
			if err != nil {
				t.Fatal(err)
			}

			if skill.Metadata.Scope != tt.expScope {
				t.Errorf("Scope = %q, want %q", skill.Metadata.Scope, tt.expScope)
			}
		})
	}
}
