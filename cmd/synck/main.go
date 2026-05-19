package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"skillsync/tui/internal/diff"
	"skillsync/tui/internal/discovery"
	"skillsync/tui/internal/install"
	"skillsync/tui/internal/opencode"
	"skillsync/tui/internal/parser"
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/syncengine"
	"skillsync/tui/internal/types"
	"skillsync/tui/internal/ui"
	"strings"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// GlobalInstaller is a function that performs the global installation.
// It can be swapped in tests to avoid actual installation.
var GlobalInstaller = install.GlobalInstall
var TUIRunner = ui.Run

const version = "0.1.0"

func run(args []string) error {
	actualArgs := args[1:]
	if len(actualArgs) > 0 {
		switch actualArgs[0] {
		case "sync":
			return handleSync(actualArgs[1:])
		case "sync-opencode":
			return handleSyncOpenCode(actualArgs[1:])
		case "createskill", "create":
			return handleCreateSkill(actualArgs[1:])
		case "install":
			// Handle 'synck install -g'
			if len(actualArgs) > 1 && actualArgs[1] == "-g" {
				actualArgs = actualArgs[1:]
			}
		case "skill":
			return handleSkill()
		case "find":
			return handleFind(actualArgs[1:])
		case "fullskills":
			return handleFullSkills(actualArgs[1:])
		case "version":
			fmt.Println(versionString())
			return nil
		case "setup-opencode":
			return handleSetupOpenCode()
		default:
			if !strings.HasPrefix(actualArgs[0], "-") {
				return handleSkillSubcommand(actualArgs[0])
			}
		}
	}

	f := flag.NewFlagSet(args[0], flag.ContinueOnError)
	globalInstall := f.Bool("g", false, "Install synck globally")

	if err := f.Parse(actualArgs); err != nil {
		return err
	}

	if *globalInstall {
		fmt.Println("🚀 Installing synck globally...")
		result := GlobalInstaller()
		if !result.Success {
			fmt.Printf("❌ %s\n", result.Message)
			if result.Error != nil {
				return result.Error
			}
			return fmt.Errorf("%s", result.Message)
		}
		fmt.Println(result.Message)
		return nil
	}

	// Normal TUI launch
	return TUIRunner()
}

func handleSync(args []string) error {
	// Preserve legacy --scope behavior by delegating to the bash script path
	hasScope := false
	for _, a := range args {
		if a == "--scope" {
			hasScope = true
			break
		}
	}
	if hasScope {
		return handleSyncLegacy(args)
	}

	f := flag.NewFlagSet("sync", flag.ContinueOnError)
	verbose := f.Bool("verbose", false, "Show full diffs")
	dryRun := f.Bool("dry-run", false, "Show what would be changed without writing")
	if err := f.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	// Find project root
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root := findProjectRoot(cwd)
	if root == "" {
		return fmt.Errorf("could not find project root (missing .agents or .agent directory)")
	}

	fmt.Println("🔄 Running synchronization...")

	progressCb := func(stage string, done, total int) {
		fmt.Printf("→ [%d/%d] %s...\n", done, total, stage)
	}

	// Stages 1-2: sync engine
	opts := syncengine.SyncOptions{
		DryRun:     *dryRun,
		ProgressCb: progressCb,
		Storage:    storage.NewService(""),
	}
	engineReport, err := syncengine.Sync(root, opts)
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	// Stages 3-8: opencode
	opencodeOpts := opencode.Options{
		DryRun:     *dryRun,
		ProgressCb: progressCb,
	}
	opencodeReport, err := syncOpenCodeForRoot(root, opencodeOpts)
	if err != nil {
		fmt.Printf("⚠️ OpenCode sync warning: %v\n", err)
	}

	// Merge reports
	merged := &runner.SyncReport{}
	merged.Changes = append(merged.Changes, engineReport.Changes...)
	if opencodeReport != nil {
		merged.Changes = append(merged.Changes, opencodeReport.Changes...)
	}

	fmt.Println("✅ Synchronization complete.")
	fmt.Println()
	fmt.Print(renderReport(merged, *verbose))

	return nil
}

// handleSyncLegacy preserves the original bash-script execution path for --scope.
func handleSyncLegacy(args []string) error {
	fmt.Println("🔄 Running synchronization...")

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root := findProjectRoot(cwd)
	if root == "" {
		return fmt.Errorf("could not find project root (missing .agents or .agent directory)")
	}

	scriptPath := filepath.Join(root, runner.DefaultSyncPath)
	r := runner.NewRunner(scriptPath)
	ctx := context.Background()

	resChan := r.ExecuteSync(ctx, args)
	res := <-resChan

	if res.Stdout != "" {
		fmt.Print(res.Stdout)
	}
	if res.Stderr != "" {
		fmt.Fprint(os.Stderr, res.Stderr)
	}

	if res.Error != nil {
		return fmt.Errorf("sync failed: %w", res.Error)
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("sync exited with code %d", res.ExitCode)
	}

	// Register project after successful legacy sync
	s := storage.NewService("")
	absRoot, _ := filepath.Abs(root)
	_ = s.RegisterProject(absRoot)

	fmt.Println("✅ Synchronization complete.")

	// Auto-chain: run sync-opencode as a non-fatal post-step
	if err := runSyncOpenCode(root, opencode.Options{}, true); err != nil {
		fmt.Printf("⚠️ OpenCode sync warning: %v\n", err)
	}

	return nil
}

// handleSyncOpenCode runs the OpenCode skill sync command.
func handleSyncOpenCode(args []string) error {
	// Parse flags for sync-opencode command
	opts := opencode.Options{}
	
	f := flag.NewFlagSet("sync-opencode", flag.ContinueOnError)
	prune := f.Bool("prune", false, "Remove orphaned skills from .opencode/skills/")
	dryRun := f.Bool("dry-run", false, "Show what would be changed without writing")
	
	if err := f.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	opts.Prune = *prune
	opts.DryRun = *dryRun

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root := findProjectRoot(cwd)
	if root == "" {
		return fmt.Errorf("could not find project root")
	}
	return runSyncOpenCode(root, opts, false)
}

// runSyncOpenCode executes the OpenCode sync with the given options.
// If nonFatal is true, errors are printed as warnings instead of returned.
func runSyncOpenCode(root string, opts opencode.Options, nonFatal bool) error {
	_, err := syncOpenCodeForRoot(root, opts)
	if err != nil {
		if nonFatal {
			return err // caller will print warning
		}
		return err
	}
	if !nonFatal {
		fmt.Println("✅ OpenCode sync complete.")
	}
	return nil
}

// syncOpenCodeForRoot performs the full OpenCode sync for a given root.
func syncOpenCodeForRoot(root string, opts opencode.Options) (*runner.SyncReport, error) {
	report := &runner.SyncReport{}

	// Stage 3: Mirror skills
	if opts.ProgressCb != nil {
		opts.ProgressCb("Mirroring skills", 3, 8)
	}
	syncReport, err := opencode.SyncSkills(root, opts)
	if err != nil {
		return report, fmt.Errorf("mirror: %w", err)
	}
	report.Changes = append(report.Changes, syncReport.Changes...)

	// Stage 4: Parse mirrored skills
	if opts.ProgressCb != nil {
		opts.ProgressCb("Parsing mirrored skills", 4, 8)
	}
	skills, err := parseMirroredSkills(root)
	if err != nil {
		return report, fmt.Errorf("parse skills: %w", err)
	}

	// Stage 5: Ensure package.json / Regenerate tools
	if opts.ProgressCb != nil {
		opts.ProgressCb("Regenerating tools", 5, 8)
	}
	if !opts.DryRun {
		if err := opencode.EnsurePackageJSON(root); err != nil {
			return report, fmt.Errorf("ensure package.json: %w", err)
		}
	}
	toolsReport, err := opencode.RegenerateTools(root, skills, opts.DryRun)
	if err != nil {
		return report, fmt.Errorf("regenerate tools: %w", err)
	}
	report.Changes = append(report.Changes, toolsReport.Changes...)

	// Stage 6: Regenerate agent
	if opts.ProgressCb != nil {
		opts.ProgressCb("Regenerating agent", 6, 8)
	}
	agentReport, err := opencode.RegenerateAgent(root, skills, opts.DryRun)
	if err != nil {
		return report, fmt.Errorf("regenerate agent: %w", err)
	}
	report.Changes = append(report.Changes, agentReport.Changes...)

	// Stage 7: Copy AGENTS.md to OPENCODE.md
	if opts.ProgressCb != nil {
		opts.ProgressCb("Copying AGENTS.md", 7, 8)
	}
	if !opts.DryRun {
		copyReport, err := opencode.CopyAgentsMD(root)
		if err != nil {
			return report, fmt.Errorf("copy agents: %w", err)
		}
		report.Changes = append(report.Changes, copyReport.Changes...)
	}

	// Stage 8: Regenerate commands
	if opts.ProgressCb != nil {
		opts.ProgressCb("Regenerating commands", 8, 8)
	}
	cmdReport, err := opencode.RegenerateCommands(root, skills, opts.DryRun)
	if err != nil {
		return report, fmt.Errorf("regenerate commands: %w", err)
	}
	report.Changes = append(report.Changes, cmdReport.Changes...)

	return report, nil
}

// parseMirroredSkills parses all mirrored skills from .opencode/skills/.
func parseMirroredSkills(root string) ([]types.Skill, error) {
	skillsPath := filepath.Join(root, ".opencode", "skills")
	var skillPaths []string
	err := filepath.WalkDir(skillsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil
		}
		if info.IsDir() {
			skillFile := filepath.Join(path, "SKILL.md")
			if _, err := os.Stat(skillFile); err == nil {
				skillPaths = append(skillPaths, skillFile)
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	var skills []types.Skill
	for _, p := range skillPaths {
		skill, err := parser.Parse(p)
		if err != nil {
			continue
		}
		skills = append(skills, *skill)
	}
	return skills, nil
}

func handleCreateSkill(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: synck createskill <prompt>")
	}
	prompt := strings.TrimSpace(strings.Join(args, " "))
	if len(strings.Fields(prompt)) < 3 {
		return fmt.Errorf("strict validation failed: prompt must include at least 3 words describing scope, behavior and context")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root := findProjectRoot(cwd)
	if root == "" {
		return fmt.Errorf("could not find project root (missing .agents or .agent directory)")
	}

	skillName := slugifySkillName(prompt)
	skillDir := filepath.Join(root, ".agents", "skills", skillName)
	if _, err := os.Stat(skillDir); err == nil {
		return fmt.Errorf("strict validation failed: skill already exists at %s", skillDir)
	}

	if err := os.MkdirAll(filepath.Join(skillDir, "assets"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(skillDir, "references"), 0755); err != nil {
		return err
	}

	skillBody := renderSkillMarkdown(skillName, prompt)
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillBody), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(skillDir, "assets", "SKILL-TEMPLATE.md"), []byte(skillTemplate), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(skillDir, "references", "README.md"), []byte(referencesTemplate), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(skillDir, "CHECKLIST.md"), []byte(checklistTemplate), 0644); err != nil {
		return err
	}

	fmt.Printf("✅ Skill scaffold created: %s\n", skillDir)
	fmt.Println("Next: run 'synck sync' to register the new skill.")
	return nil
}

func slugifySkillName(prompt string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(prompt) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteRune('-')
			prevDash = true
		}
	}
	name := strings.Trim(b.String(), "-")
	if len(name) > 48 {
		name = strings.Trim(name[:48], "-")
	}
	if name == "" {
		return "new-skill"
	}
	return name
}

func renderSkillMarkdown(skillName, prompt string) string {
	return fmt.Sprintf("---\nname: %s\ndescription: >\n  %s.\n  Trigger: When the user asks for %s.\nmetadata:\n  version: '1.0'\n  scope: [root]\n  auto_invoke: 'Creating or updating %s workflows'\n---\n\n## When to Use\n\n- When this pattern appears repeatedly.\n- When generic guidance is not enough for this project.\n- When the task needs clear operational steps.\n\n## When NOT to Use\n\n- One-off tasks.\n- Cases already covered by an existing skill.\n\n## Critical Patterns\n\n- MUST follow the project conventions before generating changes.\n- MUST keep examples minimal and executable.\n- MUST prefer local references over external links.\n\n## Decision Tree\n\n~~~text\nIs there an existing skill for this? -> Reuse it\nNeed project-specific rules? -> Extend this skill\nOtherwise -> Keep guidance lightweight\n~~~\n\n## Code Examples\n\n### Example 1\n\n~~~text\nReplace with a focused example for %s\n~~~\n\n## Commands\n\n~~~bash\n# Add real commands for this skill's workflow\n~~~\n\n## Resources\n\n- **Templates**: See [assets/](assets/)\n- **Documentation**: See [references/](references/)\n", skillName, prompt, prompt, skillName, skillName)
}

const skillTemplate = `# Skill Name

---
name: { skill-name }
description: >
  {Brief description of what this skill enables}.
  Trigger: {When the AI should load this skill - be specific}.
metadata:
  version: '1.0'
  scope: [api]
  auto_invoke: '{Action that requires this skill}'
---

## When to Use

- {Condition 1}
- {Condition 2}

## Critical Patterns

- {Rule 1}
- {Rule 2}

## Code Examples

### Example

~~~text
{minimal, focused example}
~~~
`

const referencesTemplate = `# Local References

Add links to local project docs only (no web URLs):

- documentation/{topic}.md
- docs/{topic}.md
`

const checklistTemplate = "# Skill Quality Checklist (Strict)\n\n" +
	"- [ ] Frontmatter complete (`name`, `description`, `metadata.version`, `metadata.scope`, `metadata.auto_invoke`)\n" +
	"- [ ] Trigger is specific and actionable\n" +
	"- [ ] Includes When to Use and When NOT to Use\n" +
	"- [ ] Includes Critical Patterns (MUST / MUST NOT)\n" +
	"- [ ] Includes at least one focused example\n" +
	"- [ ] References only local docs\n"

func handleSkill() error {
	fmt.Println("📋 OpenCode Skill Management Commands:")
	fmt.Println("")
	fmt.Println("  /skill    — This help message")
	fmt.Println("  /find     — Search and list existing skills")
	fmt.Println("  /create   — Create a new agent skill from a prompt")
	fmt.Println("  /sync     — Synchronize skills and update AGENTS.md")
	fmt.Println("  /fullskills — Complete flow: ask desired → /find → /create → /sync")
	fmt.Println("")
	fmt.Println("Workflow: use /find to discover → /create to make new → /sync to register")
	return nil
}

func handleFind(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root := findProjectRoot(cwd)
	if root == "" {
		return fmt.Errorf("could not find project root (missing .agents or .agent directory)")
	}

	query := strings.Join(args, " ")
	skills, err := discovery.DiscoverSkills(root)
	if err != nil {
		return fmt.Errorf("failed to discover skills: %w", err)
	}

	if len(skills) == 0 {
		fmt.Println("No skills found in project.")
		fmt.Println("")
		fmt.Println("Agent guidance: Use find-skills skill to search global registry for existing skills.")
		return nil
	}

	fmt.Printf("📦 Found %d skill(s) in project:\n", len(skills))
	for _, s := range skills {
		rel, _ := filepath.Rel(root, s)
		fmt.Printf("  - %s\n", rel)
	}
	fmt.Println("")
	if query != "" {
		fmt.Printf("To search global registry, use: npx skills find %s\n", query)
	}
	return nil
}

func handleFullSkills(args []string) error {
	fmt.Println("🚀 Full Skills Workflow")
	fmt.Println("")

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root := findProjectRoot(cwd)
	if root == "" {
		return fmt.Errorf("could not find project root (missing .agents or .agent directory)")
	}

	fmt.Println("Step 1/3 — Discovering existing skills...")
	skills, err := discovery.DiscoverSkills(root)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}
	if len(skills) > 0 {
		fmt.Printf("  Found %d existing skill(s):\n", len(skills))
		for _, s := range skills {
			rel, _ := filepath.Rel(root, s)
			fmt.Printf("    - %s\n", rel)
		}
	} else {
		fmt.Println("  No existing skills found.")
	}
	fmt.Println("")

	if len(args) > 0 {
		prompt := strings.Join(args, " ")
		fmt.Printf("Step 2/3 — Creating skill for: %q\n", prompt)
		fmt.Println("  Agent guidance for /create:")
		fmt.Printf("    1. Use 'skill-creator' to generate SKILL.md for: %s\n", prompt)
		fmt.Println("    2. Save skill to .agents/skills/<name>/SKILL.md")
		fmt.Println("")
	}

	fmt.Println("Step 3/3 — After skill creation, run:")
	fmt.Println("    synck sync")
	fmt.Println("  Or from OpenCode: /sync")
	fmt.Println("")
	fmt.Println("✅ Full workflow complete. Agent will execute remaining steps.")
	return nil
}

func handleSetupOpenCode() error {
	fmt.Println("🔧 Setting up OpenCode integration...")

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root := findProjectRoot(cwd)
	if root == "" {
		return fmt.Errorf("could not find project root (missing .agents or .opencode directory)")
	}

	// Change to project root to ensure files are created in the right place
	if err := os.Chdir(root); err != nil {
		return fmt.Errorf("failed to chdir to root: %w", err)
	}

	// Ensure .opencode directory exists
	if err := os.MkdirAll(".opencode", 0755); err != nil {
		return fmt.Errorf("failed to create .opencode directory: %w", err)
	}

	if err := ui.RegisterOpenCodeTools(); err != nil {
		return fmt.Errorf("failed to register OpenCode tools: %w", err)
	}
	fmt.Println("  ✓ Registered skill commands in .opencode/package.json")

	if err := ui.RegisterSkillManagerAgent(); err != nil {
		return fmt.Errorf("failed to register skill manager agent: %w", err)
	}
	fmt.Println("  ✓ Created .opencode/agents/skill-manager.md")

	fmt.Println("✅ OpenCode integration complete.")
	fmt.Println("   Restart OpenCode or run 'synck sync' to see updated commands.")
	return nil
}

func handleSkillSubcommand(name string) error {
	fmt.Printf("Skill detected: %s\n", name)
	fmt.Printf("To use this skill, please ask OpenCode to use the command /%s\n", name)
	return nil
}

// supportedRoots lists directories that indicate a project root for skill discovery.
var supportedRoots = []string{
	".agents",
	".agent",
	".claude",
	".qwen",
	".opencode",
	".gemini",
	".cursor",
	".copilot",
	"AGENTS.md",
	"OPENCODE.md",
}

func findProjectRoot(startDir string) string {
	curr := startDir
	for {
		for _, root := range supportedRoots {
			if _, err := os.Stat(filepath.Join(curr, root)); err == nil {
				return curr
			}
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}
	return ""
}

func versionString() string {
	return fmt.Sprintf("synck version %s (build %s)", version, version)
}

func renderReport(report *runner.SyncReport, verbose bool) string {
	if len(report.Changes) == 0 {
		return "No files changed.\n"
	}

	var b strings.Builder
	b.WriteString("Changed files:\n")

	for _, change := range report.Changes {
		diffStr := change.Diff
		if verbose && (change.Status == "modified" || change.Status == "created" || change.Status == "deleted") {
			if change.Before != "" || change.After != "" {
				diffStr, _ = diff.UnifiedDiff(change.Before, change.After, 0)
			}
		}

		b.WriteString(fmt.Sprintf("\n  %s (%s) %s\n", change.Path, change.Status, change.Summary))
		b.WriteString("  " + strings.Repeat("─", 40) + "\n")

		if diffStr != "" {
			for _, line := range strings.Split(diffStr, "\n") {
				b.WriteString("  " + line + "\n")
			}
		}
	}

	return b.String()
}

