package sandbox

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"skillsync/tui/internal/types"
	"strings"
	"testing"
)

func TestSandbox(t *testing.T) {
	// 1. Create sandbox
	sb, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer sb.Cleanup()

	// 2. Create fixtures
	fixtures := []Fixture{
		{
			Path:    ".agents/skills/test-skill/SKILL.md",
			Content: "# Test Skill",
		},
		{
			Path:    ".agents/skills/parent/SKILL.md",
			Content: "# Parent",
		},
		{
			Path:    ".agents/skills/parent/nested/SKILL.md",
			Content: "# Nested (should be skipped)",
		},
	}
	if err := sb.CreateFixtures(fixtures); err != nil {
		t.Fatalf("CreateFixtures() failed: %v", err)
	}

	// 3. Discovery
	found, err := sb.RunDiscovery()
	if err != nil {
		t.Fatalf("RunDiscovery() failed: %v", err)
	}
	if len(found) != 2 {
		t.Errorf("Expected 2 skills (SkipDir logic), found %d", len(found))
	}

	// 4. Parse
	skills, err := sb.ParseSkills(found)
	if err != nil {
		t.Fatalf("ParseSkills() failed: %v", err)
	}
	if len(skills) != 2 {
		t.Errorf("Expected 2 skills parsed, found %d", len(skills))
	}

	// 5. Normalization verification
	for _, sk := range skills {
		if strings.Contains(sk.ID, "\\") {
			t.Errorf("Skill.ID should not contain backslashes: %s", sk.ID)
		}
	}

	// Check for expected names
	names := map[string]bool{"test-skill": false, "parent": false}
	for _, sk := range skills {
		if _, ok := names[sk.Name]; ok {
			names[sk.Name] = true
		}
	}
	for name, found := range names {
		if !found {
			t.Errorf("Expected skill %s not found", name)
		}
	}
}

func TestSimulateInstall(t *testing.T) {
	sb, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer sb.Cleanup()

	// Create a source skill to copy from
	sourcePath := filepath.Join(sb.Root, "source-skill.md")
	os.WriteFile(sourcePath, []byte("# Source Content"), 0644)

	err = sb.SimulateInstall("manual-install", ".opencode", sourcePath)
	if err != nil {
		t.Fatalf("SimulateInstall() failed: %v", err)
	}

	path := filepath.Join(sb.Root, ".opencode", "skills", "manual-install", "SKILL.md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file at %s, not found", path)
	}
}

func TestToSummary(t *testing.T) {
	skills := []*types.Skill{
		{
			ID:      "test-id",
			Name:    "test-name",
			Path:    "/path/to/skill",
			RawBody: "body",
			Metadata: types.Metadata{
				Description: "desc",
			},
		},
	}

	summary := ToSummary(skills)
	if len(summary) != 1 {
		t.Fatalf("Expected 1 summary, got %d", len(summary))
	}

	s := summary[0]
	if s.ID != "test-id" || s.Name != "test-name" || s.Path != "/path/to/skill" || s.RawBody != "body" || s.Description != "desc" {
		t.Errorf("Summary fields mismatch: %+v", s)
	}
}

func TestPrintJSON(t *testing.T) {
	sb := &Sandbox{Root: "/tmp/fake"}
	result := RunResult{
		Root:       "/tmp/fake",
		KeepTemp:   true,
		Discovered: []string{"path1", "path2"},
		Skills: []SkillSummary{
			{ID: "id1", Name: "name1"},
		},
		BugReproResult: "OK",
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := sb.PrintJSON(result)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("PrintJSON() failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	var decoded RunResult
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("Failed to decode JSON output: %v", err)
	}

	if decoded.Root != result.Root || len(decoded.Discovered) != 2 || decoded.BugReproResult != "OK" {
		t.Errorf("Decoded result mismatch: %+v", decoded)
	}
}
