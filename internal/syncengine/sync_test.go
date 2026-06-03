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

	err := UpdateAgentsMarkdown(tmpDir, nil, "", false)
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

	err := UpdateAgentsMarkdown(tmpDir, nil, "", false)
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
				AutoInvoke: []string{"new-skill"},
			},
		},
	}
	
	err := UpdateAgentsMarkdown(tmpDir, skills, "", false)
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

func TestUpdateAgentsMarkdown_FiltersByScope(t *testing.T) {
	tmpDir := t.TempDir()
	
	agentsContent := `# Root Agents
## Available Skills

| Skill | Description | Location |
| --- | --- | --- |

### Auto-invoke Skills

| Action | Skill |
| --- | --- |
`
	os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte(agentsContent), 0644)
	
	skills := []types.Skill{
		{
			Name: "root-skill",
			Metadata: types.Metadata{
				Scope: "root",
				AutoInvoke: []string{"root-skill"},
			},
		},
		{
			Name: "ui-skill",
			Metadata: types.Metadata{
				Scope: "ui",
				AutoInvoke: []string{"ui-skill"},
			},
		},
	}
	
	// Test with scope "ui"
	err := UpdateAgentsMarkdown(tmpDir, skills, "ui", false)
	if err != nil {
		t.Fatalf("UpdateAgentsMarkdown failed: %v", err)
	}
	
	newAgentsContent, _ := os.ReadFile(filepath.Join(tmpDir, "AGENTS.md"))
	parts := strings.Split(string(newAgentsContent), "### Auto-invoke Skills")
	if len(parts) < 2 {
		t.Fatalf("could not find Auto-invoke Skills section in AGENTS.md")
	}
	autoInvokeSection := parts[1]
	
	if !strings.Contains(autoInvokeSection, "ui-skill") {
		t.Errorf("Auto-invoke section did not contain ui-skill")
	}
	if strings.Contains(autoInvokeSection, "root-skill") {
		t.Errorf("Auto-invoke section contained root-skill but expected only ui-skill")
	}
}

func TestUpdateAgentsMarkdown_MultiActionAndSorting(t *testing.T) {
	tmpDir := t.TempDir()
	
	agentsContent := `# Root Agents
## Available Skills

| Skill | Description | Location |
| --- | --- | --- |

### Auto-invoke Skills

| Action | Skill |
| --- | --- |
`
	os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte(agentsContent), 0644)
	
	skills := []types.Skill{
		{
			Name: "B-skill",
			Metadata: types.Metadata{
				Scope: "root",
				AutoInvoke: []string{"zeta"},
			},
		},
		{
			Name: "A-skill",
			Metadata: types.Metadata{
				Scope: "root",
				AutoInvoke: []string{"zeta", "Beta", "alpha"},
			},
		},
	}
	
	err := UpdateAgentsMarkdown(tmpDir, skills, "", false)
	if err != nil {
		t.Fatalf("UpdateAgentsMarkdown failed: %v", err)
	}
	
	newAgentsContent, _ := os.ReadFile(filepath.Join(tmpDir, "AGENTS.md"))
	lines := strings.Split(string(newAgentsContent), "\n")
	
	var tableRows []string
	inTable := false
	for _, line := range lines {
		if strings.Contains(line, "| ----------------------------------- | ---------- |") {
			inTable = true
			continue
		}
		if inTable {
			if strings.HasPrefix(line, "|") {
				tableRows = append(tableRows, strings.TrimSpace(line))
			} else if line != "" {
				inTable = false
			}
		}
	}
	
	expectedRows := []string{
		"| alpha                               | `A-skill` |",
		"| Beta                                | `A-skill` |",
		"| zeta                                | `A-skill` |",
		"| zeta                                | `B-skill` |",
	}
	
	if len(tableRows) != len(expectedRows) {
		t.Fatalf("Expected %d rows, got %d. Table lines:\n%s", len(expectedRows), len(tableRows), strings.Join(tableRows, "\n"))
	}
	
	for i, exp := range expectedRows {
		if tableRows[i] != exp {
			t.Errorf("At row %d: expected %q, got %q", i, exp, tableRows[i])
		}
	}
}

func TestSync_CleanupLegacyScripts(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create mock skills directories with assets
	assetsDir1 := filepath.Join(tmpDir, ".agents", "skills", "skill-a", "assets")
	assetsDir2 := filepath.Join(tmpDir, ".agents", "skills", "skill-b", "assets")
	err := os.MkdirAll(assetsDir1, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll(assetsDir2, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Write legacy files
	sh1 := filepath.Join(assetsDir1, "sync.sh")
	sh2 := filepath.Join(assetsDir1, "sync_test.sh")
	ps1 := filepath.Join(assetsDir2, "sync_test.ps1")
	otherFile := filepath.Join(assetsDir2, "keep_me.txt")

	os.WriteFile(sh1, []byte("echo sh1"), 0644)
	os.WriteFile(sh2, []byte("echo sh2"), 0644)
	os.WriteFile(ps1, []byte("echo ps1"), 0644)
	os.WriteFile(otherFile, []byte("keep"), 0644)

	// Pre-create AGENTS.md so Sync doesn't error
	agentsContent := "# Root Agents\n## Available Skills\n\n| Skill | Description | Location |\n| --- | --- | --- |\n\n### Auto-invoke Skills\n\n| Action | Skill |\n| --- | --- |\n"
	os.WriteFile(filepath.Join(tmpDir, "AGENTS.md"), []byte(agentsContent), 0644)

	// First, test with DryRun = true
	report, err := Sync(tmpDir, SyncOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify files not deleted in dry run
	for _, path := range []string{sh1, sh2, ps1, otherFile} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("file should not be deleted in dry run: %s", path)
		}
	}

	// Verify report has changes in dry run
	var foundSh1, foundSh2, foundPs1 bool
	for _, change := range report.Changes {
		if strings.Contains(change.Path, "sync.sh") {
			foundSh1 = true
			if change.Status != "deleted" {
				t.Errorf("expected status deleted, got %q", change.Status)
			}
		}
		if strings.Contains(change.Path, "sync_test.sh") {
			foundSh2 = true
		}
		if strings.Contains(change.Path, "sync_test.ps1") {
			foundPs1 = true
		}
	}
	if !foundSh1 || !foundSh2 || !foundPs1 {
		t.Errorf("dry run report did not list all legacy scripts: sh1=%v, sh2=%v, ps1=%v", foundSh1, foundSh2, foundPs1)
	}

	// Now run without DryRun (actual delete)
	report2, err := Sync(tmpDir, SyncOptions{DryRun: false})
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify legacy files deleted
	for _, path := range []string{sh1, sh2, ps1} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("file should have been deleted: %s", path)
		}
	}
	// Verify otherFile still exists
	if _, err := os.Stat(otherFile); os.IsNotExist(err) {
		t.Errorf("non-legacy file should not be deleted: %s", otherFile)
	}

	// Verify report contains the deletions
	foundSh1, foundSh2, foundPs1 = false, false, false
	for _, change := range report2.Changes {
		if strings.Contains(change.Path, "sync.sh") {
			foundSh1 = true
		}
		if strings.Contains(change.Path, "sync_test.sh") {
			foundSh2 = true
		}
		if strings.Contains(change.Path, "sync_test.ps1") {
			foundPs1 = true
		}
	}
	if !foundSh1 || !foundSh2 || !foundPs1 {
		t.Errorf("run report did not list all legacy scripts: sh1=%v, sh2=%v, ps1=%v", foundSh1, foundSh2, foundPs1)
	}
}


