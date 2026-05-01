package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"skillsync/tui/internal/discovery"
	"skillsync/tui/internal/install"
	"skillsync/tui/internal/opencode"
	"skillsync/tui/internal/parser"
	"skillsync/tui/internal/runner"
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
	fmt.Println("🔄 Running synchronization...")

	// Find project root
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
	err := syncOpenCodeForRoot(root, opts)
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
func syncOpenCodeForRoot(root string, opts opencode.Options) error {
	// Mirror skills
	if err := opencode.SyncSkills(root, opts); err != nil {
		return fmt.Errorf("mirror: %w", err)
	}

	// Parse mirrored skills
	skills, err := parseMirroredSkills(root)
	if err != nil {
		return fmt.Errorf("parse skills: %w", err)
	}

	// Ensure package.json exists before regenerating tools
	if !opts.DryRun {
		if err := opencode.EnsurePackageJSON(root); err != nil {
			return fmt.Errorf("ensure package.json: %w", err)
		}
	}

	// Regenerate tools
	if err := opencode.RegenerateTools(root, skills, opts.DryRun); err != nil {
		return fmt.Errorf("regenerate tools: %w", err)
	}

	// Regenerate agent
	if err := opencode.RegenerateAgent(root, skills, opts.DryRun); err != nil {
		return fmt.Errorf("regenerate agent: %w", err)
	}

	// Copy AGENTS.md to OPENCODE.md
	if !opts.DryRun {
		if err := opencode.CopyAgentsMD(root); err != nil {
			return fmt.Errorf("copy agents: %w", err)
		}
	}

	// Regenerate markdown commands
	if err := opencode.RegenerateCommands(root, skills, opts.DryRun); err != nil {
		return fmt.Errorf("regenerate commands: %w", err)
	}

	return nil
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
	prompt := strings.Join(args, " ")

	fmt.Printf("🛠️ Preparing to create skill: %q\n", prompt)
	fmt.Println("\nGUIDE FOR AI AGENT:")
	fmt.Println("1. SEARCH: Use 'find-skills' to check if a similar skill already exists.")
	fmt.Printf("   Pattern to search: %q\n", prompt)
	fmt.Println("2. CREATE: Use 'skill-creator' to generate the new skill based on the prompt.")
	fmt.Println("3. SYNC: After the skill files are created, run 'synck sync' to register it.")
	fmt.Println("\nWaiting for agent to perform steps...")

	return nil
}

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

