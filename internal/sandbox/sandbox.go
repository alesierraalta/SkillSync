package sandbox

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skillsync/tui/internal/discovery"
	"skillsync/tui/internal/parser"
	"skillsync/tui/internal/types"
)

// Fixture represents a mock skill for testing.
type Fixture struct {
	Path    string
	Content string
}

// DefaultFixtures provides a set of common test scenarios.
var DefaultFixtures = []Fixture{
	{
		Path:    ".agents/skills/deploy/SKILL.md",
		Content: "---\ndescription: Deployment skill in .agents\n---\n# Deploy\nBody content.",
	},
	{
		Path:    ".opencode/skills/deploy/SKILL.md",
		Content: "---\ndescription: Deployment skill in .opencode (Collision Case)\n---\n# Deploy\nDifferent body.",
	},
	{
		Path:    ".agents/skills/nested-skill/SKILL.md",
		Content: "# Nested Top\nThis should be found.",
	},
	{
		Path:    ".agents/skills/nested-skill/sub/SKILL.md",
		Content: "# Nested Sub\nThis should be SKIPPED if parent has SKILL.md.",
	},
	{
		Path:    ".agents/skills/deep/folder/structure/SKILL.md",
		Content: "# Deep Skill\nShould be found even if deep.",
	},
	{
		Path:    ".agents/skills/bug-repro/SKILL.md",
		Content: "Contenido base.",
	},
	{
		Path:    ".agents/skills/bug-repro/nested/SKILL.md",
		Content: "# Nested Skill\nShould be ignored due to SkipDir logic.",
	},
	{
		Path:    ".unsupported/skills/ignored/SKILL.md",
		Content: "# Ignored\nShould not be discovered.",
	},
}

// Sandbox represents an isolated test environment.
type Sandbox struct {
	Root string
}

// RunResult captures the entire execution state for JSON reporting.
type RunResult struct {
	Root           string          `json:"root"`
	IsTemp         bool            `json:"is_temp"`
	KeepTemp       bool            `json:"keep_temp"`
	Discovered     []string        `json:"discovered_paths"`
	VirtualPaths   []string        `json:"virtual_paths"`
	Skills         []SkillSummary  `json:"skills"`
	InstallResult  *InstallSummary `json:"install_simulation"`
	BugReproResult string          `json:"bug_repro_result"`
}

// SkillSummary provides a serializable view of a types.Skill.
type SkillSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Provider    string `json:"provider"`
	RawBody     string `json:"raw_body"`
	Description string `json:"description"`
}

// InstallSummary provides a serializable view of the installation simulation.
type InstallSummary struct {
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	Discovered bool   `json:"discovered_after_install"`
}

// New creates a new sandbox in a temporary directory.
func New() (*Sandbox, error) {
	tmpRoot, err := os.MkdirTemp("", "skill-sandbox-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	return &Sandbox{Root: tmpRoot}, nil
}

// Cleanup removes the sandbox directory.
func (s *Sandbox) Cleanup() error {
	return os.RemoveAll(s.Root)
}

// CreateFixtures populates the sandbox with the given fixtures.
func (s *Sandbox) CreateFixtures(fixtures []Fixture) error {
	for _, f := range fixtures {
		fullPath := filepath.Join(s.Root, filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("failed to create dir for %s: %w", f.Path, err)
		}
		if err := os.WriteFile(fullPath, []byte(f.Content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", f.Path, err)
		}
	}
	return nil
}

// RunDiscovery executes the discovery logic on the sandbox root.
func (s *Sandbox) RunDiscovery() ([]string, error) {
	return discovery.DiscoverSkills(s.Root)
}

// ParseSkills parses a list of skill file paths.
func (s *Sandbox) ParseSkills(paths []string) ([]*types.Skill, error) {
	var skills []*types.Skill
	for _, p := range paths {
		sk, err := parser.Parse(p)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", p, err)
		}
		skills = append(skills, sk)
	}
	return skills, nil
}

// ToSummary converts types.Skill slice to SkillSummary slice.
func ToSummary(skills []*types.Skill) []SkillSummary {
	summary := make([]SkillSummary, len(skills))
	for i, sk := range skills {
		provider := "root"
		// Simple heuristic to find provider (e.g., .agents)
		parts := strings.Split(filepath.ToSlash(sk.Path), "/")
		for _, p := range parts {
			if strings.HasPrefix(p, ".") && p != "." && p != ".." {
				provider = p
				break
			}
		}
		if sk.ID == "virtual:agents" {
			provider = "virtual"
		}

		summary[i] = SkillSummary{
			ID:          sk.ID,
			Name:        sk.Name,
			Path:        sk.Path,
			Provider:    provider,
			RawBody:     sk.RawBody,
			Description: sk.Metadata.Description,
		}
	}
	return summary
}

// RunVirtualDiscovery looks for AGENTS.md in the root.
func (s *Sandbox) RunVirtualDiscovery() []string {
	var virtuals []string
	agentsPath := filepath.Join(s.Root, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err == nil {
		virtuals = append(virtuals, agentsPath)
	}
	return virtuals
}

// PrintResults outputs the results of discovery and parsing to stdout.
func (s *Sandbox) PrintResults(skills []*types.Skill) {
	fmt.Printf("%-15s | %-20s | %-15s | %s\n", "Provider", "Name", "ID", "Description")
	fmt.Println(strings.Repeat("-", 120))

	for _, sk := range skills {
		provider := "root"
		parts := strings.Split(filepath.ToSlash(sk.Path), "/")
		for _, p := range parts {
			if strings.HasPrefix(p, ".") && p != "." && p != ".." {
				provider = p
				break
			}
		}
		if sk.ID == "virtual:agents" {
			provider = "virtual"
		}

		desc := sk.Metadata.Description
		if desc == "" {
			desc = "(No description)"
		}
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}

		name := sk.Name
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		id := sk.ID
		if strings.Contains(id, "/") {
			parts := strings.Split(id, "/")
			id = parts[len(parts)-1] // Just show file name for ID column to save space
		}

		fmt.Printf("%-15s | %-20s | %-15s | %s\n", provider, name, id, desc)
	}
}

// PrintJSON outputs the RunResult as a structured JSON document.
func (s *Sandbox) PrintJSON(result RunResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// SimulateInstall copies a real skill into the sandbox.
func (s *Sandbox) SimulateInstall(name, provider string, sourcePath string) error {
	// Source is real skill path
	destDir := filepath.Join(s.Root, provider, "skills", name)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// Copy content
	sourceContent, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(destDir, "SKILL.md"), sourceContent, 0644)
}
