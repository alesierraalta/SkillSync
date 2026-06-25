package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
	"skillsync/tui/internal/types"
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
	if skill.ID == "" {
		t.Error("Expected skill.ID to be populated from path")
	}
	if filepath.ToSlash(path) != skill.ID {
		t.Errorf("Expected skill.ID %q, got %q", filepath.ToSlash(path), skill.ID)
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
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

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
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Round-trip 1
	skill, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse 1 failed: %v", err)
	}
	err = Save(path, skill)
	if err != nil {
		t.Fatalf("Save 1 failed: %v", err)
	}

	// Read back raw
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile 1 failed: %v", err)
	}
	interim := string(raw)

	// Round-trip 2
	skill2, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse 2 failed: %v", err)
	}
	err = Save(path, skill2)
	if err != nil {
		t.Fatalf("Save 2 failed: %v", err)
	}

	raw2, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile 2 failed: %v", err)
	}
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
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

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
	if filepath.ToSlash(path) != skill.ID {
		t.Errorf("Expected skill.ID %q, got %q", filepath.ToSlash(path), skill.ID)
	}
}

func TestParseFlexibleScope(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expScope string
	}{
		{
			name:     "Nested List",
			yaml:     "---\nmetadata:\n  scope: [ui, root]\n---",
			expScope: "ui, root",
		},
		{
			name:     "Nested String",
			yaml:     "---\nmetadata:\n  scope: api\n---",
			expScope: "api",
		},
		{
			name:     "Top Level List",
			yaml:     "---\nscope: [common, infra]\n---",
			expScope: "common, infra",
		},
		{
			name:     "Top Level String",
			yaml:     "---\nscope: root\n---",
			expScope: "root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			path := filepath.Join(tmp, "SKILL.md")
			if err := os.WriteFile(path, []byte(tt.yaml+"\n# Body"), 0644); err != nil {
				t.Fatal(err)
			}

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

func TestFormat(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "SKILL.md")
	content := `# Prefix
---
# Comment
scope: original
---
Body content`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	skill, err := Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	// Modify
	skill.Metadata.Scope = "updated"
	skill.RawBody = "Updated body"

	formatted, err := Format(skill)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(formatted, "# Prefix") {
		t.Error("Prefix lost")
	}
	if !strings.Contains(formatted, "# Comment") {
		t.Error("YAML comment lost")
	}
	if !strings.Contains(formatted, "scope: updated") {
		t.Error("Scope not updated in formatted string")
	}
	if !strings.Contains(formatted, "Updated body") {
		t.Error("Body not updated in formatted string")
	}
	if !strings.HasPrefix(formatted, "# Prefix") {
		t.Errorf("Format should start with prefix, got: %q", formatted)
	}
}

func TestParseFlexibleAutoInvoke(t *testing.T) {
	tests := []struct {
		name          string
		yaml          string
		skillName     string
		expAutoInvoke []string
	}{
		{
			name:          "Boolean True",
			yaml:          "---\nauto_invoke: true\n---",
			skillName:     "git-branches",
			expAutoInvoke: []string{"git-branches"},
		},
		{
			name:          "Boolean False",
			yaml:          "---\nauto_invoke: false\n---",
			skillName:     "git-branches",
			expAutoInvoke: []string{},
		},
		{
			name:          "String Value",
			yaml:          "---\nauto_invoke: run-tests\n---",
			skillName:     "tester",
			expAutoInvoke: []string{"run-tests"},
		},
		{
			name:          "Array Value",
			yaml:          "---\nauto_invoke:\n  - action A\n  - action B\n---",
			skillName:     "multi",
			expAutoInvoke: []string{"action A", "action B"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			dir := filepath.Join(tmp, tt.skillName)
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, "SKILL.md")
			if err := os.WriteFile(path, []byte(tt.yaml+"\n# Body"), 0644); err != nil {
				t.Fatal(err)
			}

			skill, err := Parse(path)
			if err != nil {
				t.Fatal(err)
			}

			if len(skill.Metadata.AutoInvoke) != len(tt.expAutoInvoke) {
				t.Fatalf("AutoInvoke length mismatch: got %d (%v), want %d (%v)",
					len(skill.Metadata.AutoInvoke), skill.Metadata.AutoInvoke,
					len(tt.expAutoInvoke), tt.expAutoInvoke)
			}
			for i, v := range skill.Metadata.AutoInvoke {
				if v != tt.expAutoInvoke[i] {
					t.Errorf("AutoInvoke[%d] = %q, want %q", i, v, tt.expAutoInvoke[i])
				}
			}
		})
	}
}

func TestFormatFromScratch(t *testing.T) {
	skill := &types.Skill{
		Name:    "my-new-skill",
		Prefix:  "# Some Prefix\n",
		RawBody: "This is the body.",
	}
	skill.Metadata.Description = "This is a description."
	skill.Metadata.Scope = "project"
	skill.Metadata.AutoInvoke = []string{"trigger-1", "trigger-2"}
	skill.Metadata.LocalOnly = true

	formatted, err := Format(skill)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(formatted, "name: my-new-skill") {
		t.Error("Missing 'name: my-new-skill'")
	}
	if !strings.Contains(formatted, "description: This is a description.") {
		t.Error("Missing 'description'")
	}
	if !strings.Contains(formatted, "scope: project") {
		t.Error("Missing 'scope: project'")
	}
	if !strings.Contains(formatted, "local_only: true") {
		t.Error("Missing 'local_only: true'")
	}
	if !strings.Contains(formatted, "trigger-1") || !strings.Contains(formatted, "trigger-2") {
		t.Error("Missing auto_invoke triggers")
	}
}

func TestFormatPreservesNestedMetadata(t *testing.T) {
	input := `name: gentleman-bubbletea
description: Bubbletea patterns
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.0"
  scope: [root]
  auto_invoke: "Working on TUI screens"
`
	tmp := t.TempDir()
	path := filepath.Join(tmp, "SKILL.md")
	content := "---\n" + input + "---\nBody"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	skill, err := Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	skill.Metadata.Scope = "ui, root"

	formatted, err := Format(skill)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(formatted, "author: gentleman-programming") {
		t.Error("custom field metadata.author was not preserved")
	}
	if !strings.Contains(formatted, "license: Apache-2.0") {
		t.Error("custom field license was not preserved")
	}

	var parsed struct {
		Scope    string `yaml:"scope"`
		Metadata struct {
			Scope      interface{} `yaml:"scope"`
			Author     string      `yaml:"author"`
			Version    string      `yaml:"version"`
			AutoInvoke string      `yaml:"auto_invoke"`
		} `yaml:"metadata"`
	}

	parts := strings.Split(formatted, "---")
	if len(parts) < 3 {
		t.Fatalf("expected at least 3 parts, got %d", len(parts))
	}
	err = yaml.Unmarshal([]byte(parts[1]), &parsed)
	if err != nil {
		t.Fatalf("failed to unmarshal formatted YAML: %v", err)
	}

	if parsed.Scope != "" {
		t.Errorf("Duplicate or incorrect root level scope found: %q", parsed.Scope)
	}
	normalizedScope := normalizeScope(parsed.Metadata.Scope)
	if normalizedScope != "ui, root" {
		t.Errorf("Expected nested metadata.scope to be 'ui, root', got %q", normalizedScope)
	}
	if parsed.Metadata.Author != "gentleman-programming" {
		t.Errorf("Expected nested metadata.author to be 'gentleman-programming', got %q", parsed.Metadata.Author)
	}
}

func TestParseContentWithNameFallbackName(t *testing.T) {
	content := "---\ndescription: Fallback test\n---\nBody"
	skill, err := ParseContentWithName(content, "fallback-name")
	if err != nil {
		t.Fatalf("ParseContentWithName failed: %v", err)
	}
	if skill.Name != "fallback-name" {
		t.Errorf("Expected Name to be 'fallback-name', got %q", skill.Name)
	}
}

func TestEmptyFieldsNotAppended(t *testing.T) {
	tmp := t.TempDir()

	// Root-level metadata without scope or auto_invoke originally present
	content := "---\nname: my-skill\ndescription: Test description\n---\nBody"
	path1 := filepath.Join(tmp, "SKILL1.md")
	if err := os.WriteFile(path1, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	skill, err := Parse(path1)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Metadata has empty scope and auto_invoke
	skill.Metadata.Scope = ""
	skill.Metadata.AutoInvoke = []string{}

	formatted, err := Format(skill)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if strings.Contains(formatted, "scope:") {
		t.Error("scope was appended to YAML even though it was empty and not originally present")
	}
	if strings.Contains(formatted, "auto_invoke:") {
		t.Error("auto_invoke was appended to YAML even though it was empty and not originally present")
	}

	// Nested metadata without scope or auto_invoke originally present
	nestedContent := `---
name: my-skill
metadata:
  author: tester
---
Body`
	path2 := filepath.Join(tmp, "SKILL2.md")
	if err := os.WriteFile(path2, []byte(nestedContent), 0644); err != nil {
		t.Fatal(err)
	}

	skillNested, err := Parse(path2)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	skillNested.Metadata.Scope = ""
	skillNested.Metadata.AutoInvoke = []string{}

	formattedNested, err := Format(skillNested)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if strings.Contains(formattedNested, "scope:") {
		t.Error("scope was appended to nested metadata even though it was empty and not originally present")
	}
	if strings.Contains(formattedNested, "auto_invoke:") {
		t.Error("auto_invoke was appended to nested metadata even though it was empty and not originally present")
	}
}

func TestSaveBackupStrategy(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "SKILL.md")

	// Create original file
	err := os.WriteFile(path, []byte("original content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	skill := &types.Skill{
		Name:    "my-skill",
		RawBody: "new body",
	}

	err = Save(path, skill)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Ensure original is updated
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "new body") {
		t.Errorf("Expected file to be updated, got: %q", string(raw))
	}

	// Ensure backup is cleaned up
	if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
		t.Error("Backup file (.bak) was not cleaned up on success")
	}
}

