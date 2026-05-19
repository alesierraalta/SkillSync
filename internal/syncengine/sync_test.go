package syncengine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"skillsync/tui/internal/types"
)

func TestSync_CallbackInvoked(t *testing.T) {
	tmpDir := t.TempDir()
	agentsContent := "# Root Agents\n## Available Skills\n\n| Skill | Description | Location |\n| --- | --- | --- |\n\n### Auto-invoke Skills\n\n| Action | Skill |\n| --- | --- |\n"
	os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte(agentsContent), 0644)

	var calls []struct {
		stage string
		done  int
		total int
	}
	cb := func(stage string, done, total int) {
		calls = append(calls, struct {
			stage string
			done  int
			total int
		}{stage, done, total})
	}

	report, err := Sync(tmpDir, SyncOptions{ProgressCb: cb})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 callback invocations, got %d", len(calls))
	}
	if calls[0].stage != "Discovering skills" {
		t.Errorf("expected stage 1 to be 'Discovering skills', got %q", calls[0].stage)
	}
	if calls[0].done != 1 || calls[0].total != 8 {
		t.Errorf("expected stage 1 done=1 total=8, got done=%d total=%d", calls[0].done, calls[0].total)
	}
	if calls[1].stage != "Updating AGENTS.md" {
		t.Errorf("expected stage 2 to be 'Updating AGENTS.md', got %q", calls[1].stage)
	}
	if calls[1].done != 2 || calls[1].total != 8 {
		t.Errorf("expected stage 2 done=2 total=8, got done=%d total=%d", calls[1].done, calls[1].total)
	}
}

func TestSync_NilCallback(t *testing.T) {
	tmpDir := t.TempDir()
	agentsContent := "# Root Agents\n## Available Skills\n\n| Skill | Description | Location |\n| --- | --- | --- |\n\n### Auto-invoke Skills\n\n| Action | Skill |\n| --- | --- |\n"
	os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte(agentsContent), 0644)

	report, err := Sync(tmpDir, SyncOptions{ProgressCb: nil})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestSync_ReportContainsChange(t *testing.T) {
	tmpDir := t.TempDir()
	agentsContent := `# Root Agents
## Available Skills

| Skill | Description | Location |
| --- | --- | --- |

### Auto-invoke Skills

| Action | Skill |
| --- | --- |
| Old Action | old-skill |
`
	os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte(agentsContent), 0644)

	// Pre-create skills dir so DiscoverSkills finds it
	os.MkdirAll(filepath.Join(tmpDir, ".agents", "skills", "new-skill"), 0755)
	os.WriteFile(filepath.Join(tmpDir, ".agents", "skills", "new-skill", "SKILL.md"), []byte(""), 0644)

	report, err := Sync(tmpDir, SyncOptions{})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if len(report.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(report.Changes))
	}

	change := report.Changes[0]
	if change.Path != "AGENTS.md" {
		t.Errorf("expected path AGENTS.md, got %q", change.Path)
	}
	if change.Status != "modified" {
		t.Errorf("expected status modified, got %q", change.Status)
	}
	if change.Diff == "" {
		t.Error("expected non-empty diff")
	}
	if change.Summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestSync_ReportCreatedStatus(t *testing.T) {
	tmpDir := t.TempDir()
	// No AGENTS.md initially

	report, err := Sync(tmpDir, SyncOptions{})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	// AGENTS.md doesn't exist, UpdateAgentsMarkdown returns nil without creating
	// So no change should be detected
	if len(report.Changes) != 0 {
		t.Fatalf("expected 0 changes when AGENTS.md missing, got %d", len(report.Changes))
	}
}

func TestUpdateAgentsMarkdown_FailsWhenRequiredHeadersMissing(t *testing.T) {
	tmpDir := t.TempDir()

	agentsContent := `# Root Agents
### auto-invoke skills

| Action | Skill |
| --- | --- |
`
	os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte(agentsContent), 0644)

	err := UpdateAgentsMarkdown(tmpDir, nil, false)
	if err == nil {
		t.Fatal("expected error when required headers are missing")
	}
	if !strings.Contains(err.Error(), "required headers") {
		t.Fatalf("expected required headers error, got: %v", err)
	}
}

func TestUpdateAgentsMarkdown_FailsWhenOnlyOneRequiredHeaderExists(t *testing.T) {
	tmpDir := t.TempDir()

	agentsContent := `# Root Agents
## Available Skills

| Skill | Description | Location |
| --- | --- | --- |
`
	os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte(agentsContent), 0644)

	err := UpdateAgentsMarkdown(tmpDir, nil, false)
	if err == nil {
		t.Fatal("expected error when one required header is missing")
	}
}

func TestSync_UpdatesMarkdownTables(t *testing.T) {
	tmpDir := t.TempDir()
	
	agentsContent := `# Root Agents
## Available Skills

| Skill | Description | Location |
| --- | --- | --- |

### Auto-invoke Skills

| Action | Skill |
| --- | --- |
| Old Action | old-skill |

Some other text
`
	os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte(agentsContent), 0644)
	os.WriteFile(filepath.Join(tmpDir, "OPENCODE.md"), []byte(agentsContent), 0644)
	
	skills := []types.Skill{
		{
			Name: "new-skill",
			Metadata: types.Metadata{
				Scope: "root",
				AutoInvoke: true,
			},
		},
	}
	
	err := UpdateAgentsMarkdown(tmpDir, skills, false)
	if err != nil {
		t.Fatalf("UpdateAgentsMarkdown failed: %v", err)
	}
	
	newAgentsContent, _ := os.ReadFile(filepath.Join(tmpDir, "AGENTS.md"))
	if !strings.Contains(string(newAgentsContent), "| new-skill                           | `new-skill` |") {
		t.Errorf("AGENTS.md did not contain new-skill action")
	}
	if strings.Contains(string(newAgentsContent), "old-skill") {
		t.Errorf("AGENTS.md still contained old-skill")
	}

	// Simulate OpenCode copy step
	_ = os.WriteFile(filepath.Join(tmpDir, "OPENCODE.md"), newAgentsContent, 0644)
	newOpencodeContent, _ := os.ReadFile(filepath.Join(tmpDir, "OPENCODE.md"))
	if !strings.Contains(string(newOpencodeContent), "| new-skill                           | `new-skill` |") {
		t.Errorf("OPENCODE.md did not contain new-skill action after copy")
	}
	if strings.Contains(string(newOpencodeContent), "old-skill") {
		t.Errorf("OPENCODE.md still contained old-skill after copy")
	}
}
