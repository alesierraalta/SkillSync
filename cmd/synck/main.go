package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"skillsync/tui/internal/bundle"
	"skillsync/tui/internal/diff"
	"skillsync/tui/internal/discovery"
	"skillsync/tui/internal/install"
	"skillsync/tui/internal/opencode"
	"skillsync/tui/internal/remove"
	"skillsync/tui/internal/runner"
	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/syncengine"
	"skillsync/tui/internal/ui"
	"skillsync/tui/internal/vault"
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

var AutoskillsInstaller = func(ctx context.Context) install.AutoskillsResult {
	return install.AutoskillsInstall(ctx)
}

var AutoskillsPreflightChecker = func() error {
	return install.AutoskillsPreflight()
}
var TUIRunner = ui.Run

// BundleExport is the export function used by the bundle subcommand.
// Overridable in tests.
var BundleExport = bundle.Export

// BundleImport is the import function used by the bundle subcommand.
// Overridable in tests.
var BundleImport = bundle.Import

// BundleReadManifest reads a bundle's manifest for listing.
// Overridable in tests.
var BundleReadManifest = bundle.ReadManifest

// exitCodeFromResults returns an error if any ImportResult has Status "failed".
// This ensures the CLI exits with a non-zero code when imports fail.
func exitCodeFromResults(results []bundle.ImportResult) error {
	var failedCount int
	for _, r := range results {
		if r.Status == "failed" {
			failedCount++
		}
	}
	if failedCount > 0 {
		return fmt.Errorf("%d skill(s) failed", failedCount)
	}
	return nil
}

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
		case "bundle":
			return handleBundle(actualArgs[1:])
		case "install":
			// Handle 'synck install -g'
			if len(actualArgs) > 1 && actualArgs[1] == "-g" {
				actualArgs = actualArgs[1:]
			}
			// Handle 'synck install --bundle <path>'
			if len(actualArgs) > 1 && actualArgs[1] == "--bundle" {
				return handleInstallBundle(actualArgs[2:])
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
		case "remove":
			return handleRemove(actualArgs[1:])
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
	f := flag.NewFlagSet("sync", flag.ContinueOnError)
	verbose := f.Bool("verbose", false, "Show full diffs")
	dryRun := f.Bool("dry-run", false, "Show what would be changed without writing")
	autoskills := f.Bool("autoskills", false, "Run autoskills discovery before sync")
	scope := f.String("scope", "", "Sync specific scope only")
	if err := f.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	if *autoskills {
		fmt.Println("🔍 Running autoskills discovery...")
		if err := AutoskillsPreflightChecker(); err != nil {
			return fmt.Errorf("autoskills preflight failed: %w", err)
		}
		res := AutoskillsInstaller(context.Background())
		if !res.Success {
			fmt.Printf("⚠️  Autoskills failed: %v\n", res.Error)
			if res.Output != "" {
				fmt.Println(res.Output)
			}
			return fmt.Errorf("autoskills execution failed")
		}
		fmt.Println("✅ Autoskills discovery complete.")
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
		Scope:      *scope,
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
	plannedAgentsContent := ""
	if *dryRun {
		plannedAgentsContent = plannedAgentsMD(engineReport)
	}
	opencodeReport, err := syncOpenCodeForRootWithAgentsContent(root, opencodeOpts, plannedAgentsContent)
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
	return syncOpenCodeForRootWithAgentsContent(root, opts, "")
}

func syncOpenCodeForRootWithAgentsContent(root string, opts opencode.Options, agentsMDContent string) (*runner.SyncReport, error) {
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

	// Stage 4: Discover skills from source
	if opts.ProgressCb != nil {
		opts.ProgressCb("Discovering skills", 4, 8)
	}
	skills, err := syncengine.DiscoverSkills(root)
	if err != nil {
		return report, fmt.Errorf("discover skills: %w", err)
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
	var copyReport *runner.SyncReport
	if agentsMDContent != "" {
		copyReport, err = opencode.CopyAgentsMDContent(root, agentsMDContent, opts.DryRun)
	} else {
		copyReport, err = opencode.CopyAgentsMD(root, opts.DryRun)
	}
	if err != nil {
		return report, fmt.Errorf("copy agents: %w", err)
	}
	report.Changes = append(report.Changes, copyReport.Changes...)

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

func plannedAgentsMD(report *runner.SyncReport) string {
	if report == nil {
		return ""
	}
	for _, change := range report.Changes {
		if change.Path == "AGENTS.md" && change.After != "" {
			return change.After
		}
	}
	return ""
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

// handleRemove parses and runs the remove subcommand.
// Supports: synck remove [--local] [--force] <skill-name>
//
//	synck remove <skill-name> [--local] [--force]
func handleRemove(args []string) error {
	var name string
	local := false
	force := false

	for _, arg := range args {
		switch arg {
		case "--local":
			local = true
		case "--force":
			force = true
		case "--help", "-h":
			fmt.Println("usage: synck remove [--local] [--force] <skill-name>")
			return nil
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag: %s", arg)
			}
			if name == "" {
				name = arg
			} else {
				return fmt.Errorf("unexpected argument: %s", arg)
			}
		}
	}

	if name == "" {
		return fmt.Errorf("usage: synck remove [--local] [--force] <skill-name>")
	}

	return runRemove(name, local, force)
}

// runRemove performs the full skill removal flow.
// If force is false, it prompts for confirmation via stdin.
func runRemove(name string, local, force bool) error {
	// If not --force, prompt for confirmation
	if !force {
		fmt.Printf("Remove skill '%s'? [y/N]: ", name)
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil || (response != "y" && response != "Y") {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Reject core skills first (before checking existence)
	if remove.IsCoreSkill(name) {
		return fmt.Errorf("cannot remove core skill %q", name)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	root := findProjectRoot(cwd)
	if root == "" {
		return fmt.Errorf("could not find project root (missing .agents or .agent directory)")
	}

	// Check that the skill actually exists
	skillDir := filepath.Join(root, ".agents", "skills", name)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return fmt.Errorf("skill %q not found", name)
	}

	s := &remove.Service{
		RootPath: root,
		Storage:  storage.NewService(""),
	}

	if err := s.RemoveByID(name, remove.Options{Local: local}); err != nil {
		return fmt.Errorf("failed to remove skill %q: %w", name, err)
	}

	fmt.Printf("Skill '%s' deleted.\n", name)

	// Post-delete regeneration is best-effort; non-fatal warnings
	if err := ui.RegenerateAfterDelete(root); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: post-delete regeneration reported errors: %v\n", err)
	}
	return nil
}

// handleBundle dispatches to bundle subcommands (export, import, list).
func handleBundle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: synck bundle <export|import|list> [args]")
	}

	switch args[0] {
	case "export":
		return handleBundleExport(args[1:])
	case "import":
		return handleBundleImport(args[1:])
	case "list":
		return handleBundleList(args[1:])
	default:
		return fmt.Errorf("unknown bundle subcommand %q (use: export, import, list)", args[0])
	}
}

// handleBundleExport exports one or more skills to a .skillsync bundle.
// Default export path is ~/.skillsync/bundles/{name}.skillsync.
func handleBundleExport(args []string) error {
	f := flag.NewFlagSet("bundle export", flag.ContinueOnError)
	output := f.String("output", "", "Output path for the bundle (default: ~/.skillsync/bundles/)")
	if err := f.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	skillNames := f.Args()
	if len(skillNames) == 0 {
		return fmt.Errorf("usage: synck bundle export <skill-name>... [--output <path>]")
	}

	// Resolve default export directory
	exportPath := *output
	if exportPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("home dir: %w", err)
		}
		bundlesDir := filepath.Join(homeDir, ".skillsync", "bundles")
		if len(skillNames) == 1 {
			exportPath = filepath.Join(bundlesDir, skillNames[0]+".skillsync")
		} else {
			exportPath = filepath.Join(bundlesDir, "skillsync-bundle.skillsync")
		}
	}

	v := vault.NewService()
	path, err := BundleExport(v, skillNames, exportPath)
	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}
	fmt.Printf("✅ Bundle created: %s\n", path)
	return nil
}

// handleBundleImport imports skills from a .skillsync bundle.
func handleBundleImport(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: synck bundle import <bundle-path>")
	}

	bundlePath := args[0]
	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		return fmt.Errorf("bundle not found: %s", bundlePath)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root := findProjectRoot(cwd)
	if root == "" {
		return fmt.Errorf("could not find project root")
	}

	opts := bundle.ImportOptions{
		ProjectRoot: root,
	}
	if len(args) > 1 && args[1] == "--force" {
		opts.OnDuplicate = bundle.DuplicateOverwrite
	}

	results, err := BundleImport(bundlePath, opts)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	fmt.Printf("📦 Imported %d skill(s) from %s\n", len(results), bundlePath)
	for _, r := range results {
		icon := "✅"
		if r.Status == "skipped" {
			icon = "⏭️"
		} else if r.Status == "warning" {
			icon = "⚠️"
		}
		if r.Error != nil {
			fmt.Printf("  %s %s: %v\n", icon, r.Skill, r.Error)
		} else {
			fmt.Printf("  %s %s (%s)\n", icon, r.Skill, r.Status)
		}
	}
	return exitCodeFromResults(results)
}

// handleBundleList lists the contents of a .skillsync bundle without extracting.
func handleBundleList(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: synck bundle list <bundle-path>")
	}

	bundlePath := args[0]
	m, err := BundleReadManifest(bundlePath)
	if err != nil {
		return fmt.Errorf("read bundle: %w", err)
	}

	fmt.Printf("📦 Bundle: %s\n", bundlePath)
	fmt.Printf("  Version:     %d\n", m.Version)
	fmt.Printf("  Created:     %s\n", m.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Created by:  %s\n", m.CreatedBy)
	fmt.Printf("  Skills:      %d\n", len(m.Skills))
	for _, s := range m.Skills {
		prov := s.OriginProvider
		if prov == "" {
			prov = "unknown"
		}
		if s.Description != "" {
			fmt.Printf("    - %s (origin: %s) — %s\n", s.Name, prov, s.Description)
		} else {
			fmt.Printf("    - %s (origin: %s)\n", s.Name, prov)
		}
	}
	return nil
}

// handleInstallBundle implements `synck install --bundle <path>`.
func handleInstallBundle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: synck install --bundle <bundle-path>")
	}

	bundlePath := args[0]
	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		return fmt.Errorf("bundle not found: %s", bundlePath)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root := findProjectRoot(cwd)
	if root == "" {
		return fmt.Errorf("could not find project root")
	}

	opts := bundle.ImportOptions{
		ProjectRoot: root,
	}
	// Check for --force after bundle path
	if len(args) > 1 && args[1] == "--force" {
		opts.OnDuplicate = bundle.DuplicateOverwrite
	}

	results, err := BundleImport(bundlePath, opts)
	if err != nil {
		return fmt.Errorf("install from bundle failed: %w", err)
	}

	fmt.Printf("📦 Installed %d skill(s) from bundle\n", len(results))
	for _, r := range results {
		icon := "✅"
		if r.Status == "skipped" {
			icon = "⏭️"
		}
		if r.Error != nil {
			fmt.Printf("  %s %s: %v\n", icon, r.Skill, r.Error)
		} else {
			fmt.Printf("  %s %s (%s)\n", icon, r.Skill, r.Status)
		}
	}
	return exitCodeFromResults(results)
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
