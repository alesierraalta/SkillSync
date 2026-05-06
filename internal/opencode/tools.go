package opencode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"skillsync/tui/internal/diff"
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/types"
)

// RegenerateTools rebuilds .opencode/package.json tools array from the
// provided skills list. Only skills with auto_invoke=true produce tool entries.
// It preserves all other package.json fields.
func RegenerateTools(root string, skills []types.Skill, dryRun bool) (*runner.SyncReport, error) {
	report := &runner.SyncReport{}
	pkgPath := filepath.Join(root, ".opencode", "package.json")

	before := ""
	if content, err := os.ReadFile(pkgPath); err == nil {
		before = string(content)
	}

	// Read existing package.json
	data := make(map[string]interface{})
	if content, err := os.ReadFile(pkgPath); err == nil {
		_ = json.Unmarshal(content, &data)
	}

	// Ensure opencode key exists
	oc, ok := data["opencode"].(map[string]interface{})
	if !ok {
		oc = make(map[string]interface{})
		data["opencode"] = oc
	}

	// Ensure dependencies key exists
	deps, ok := data["dependencies"].(map[string]interface{})
	if !ok {
		deps = make(map[string]interface{})
		data["dependencies"] = deps
	}
	if _, ok := deps["@opencode-ai/plugin"]; !ok {
		deps["@opencode-ai/plugin"] = "1.14.19"
	}

	// tools array is now empty because visibility is handled by markdown commands
	tools := make([]map[string]interface{}, 0)

	// Replace tools array (always use non-nil slice)
	oc["tools"] = tools

	if dryRun {
		fmt.Printf("[dry-run] would regenerate tools: %d entries\n", len(tools))
		return report, nil
	}

	newData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return report, fmt.Errorf("marshal package.json: %w", err)
	}

	if err := os.WriteFile(pkgPath, newData, 0644); err != nil {
		return report, err
	}

	after := string(newData)
	if before != after {
		diffStr, summary := diff.UnifiedDiff(before, after, 50)
		status := "modified"
		if before == "" {
			status = "created"
		}
		relPath, _ := filepath.Rel(root, pkgPath)
		report.Changes = append(report.Changes, runner.FileChange{
			Path:    relPath,
			Status:  status,
			Before:  before,
			After:   after,
			Diff:    diffStr,
			Summary: summary,
		})
	}

	return report, nil
}

// EnsurePackageJSON creates a minimal .opencode/package.json if it doesn't exist.
func EnsurePackageJSON(root string) error {
	pkgPath := filepath.Join(root, ".opencode", "package.json")
	if _, err := os.Stat(pkgPath); err == nil {
		return nil // Already exists
	}

	data := map[string]interface{}{
		"name": "opencode-skills",
		"dependencies": map[string]interface{}{
			"@opencode-ai/plugin": "1.14.19",
		},
		"opencode": map[string]interface{}{
			"tools": []interface{}{},
		},
	}

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal new package.json: %w", err)
	}

	dir := filepath.Dir(pkgPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	return os.WriteFile(pkgPath, content, 0644)
}
