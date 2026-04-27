package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skillsync/tui/internal/sandbox"
	"skillsync/tui/internal/parser"
	"skillsync/tui/internal/types"
)

func main() {
	keepTemp := flag.Bool("keep-temp", false, "Keep temporary fixtures for manual inspection")
	jsonMode := flag.Bool("json", false, "Output results as JSON for automated checks")
	targetPath := flag.String("path", "", "Run sandbox on a specific directory instead of temp fixtures")
	flag.Parse()

	isTemp := *targetPath == ""
	if !*jsonMode {
		fmt.Println("=== SkillSync Sandbox ===")
		if isTemp {
			fmt.Println("Initializing isolated test environment...")
		} else {
			fmt.Printf("Running on target path: %s\n", *targetPath)
		}
	}

	result := sandbox.RunResult{
		KeepTemp: *keepTemp,
		IsTemp:   isTemp,
	}

	var sb *sandbox.Sandbox
	var err error

	if isTemp {
		// 1. New Sandbox (Temp)
		sb, err = sandbox.New()
		if err != nil {
			outputError(err, *jsonMode)
			os.Exit(1)
		}
		if !*keepTemp {
			defer sb.Cleanup()
		}
		result.Root = sb.Root

		if !*jsonMode {
			fmt.Printf("Temp root: %s\n", sb.Root)
			if *keepTemp {
				fmt.Println("[NOTE] --keep-temp is active. Files will NOT be deleted.")
			}
			fmt.Println()
		}

		// 2. Setup fixtures
		if !*jsonMode {
			fmt.Println("--- Creating Fixtures ---")
		}
		err = sb.CreateFixtures(sandbox.DefaultFixtures)
		if err != nil {
			outputError(err, *jsonMode)
			os.Exit(1)
		}
		if !*jsonMode {
			for _, f := range sandbox.DefaultFixtures {
				fmt.Printf("[+] Created: %s\n", f.Path)
			}
			fmt.Println()
		}
	} else {
		// 1. Existing Path
		absPath, absErr := filepath.Abs(*targetPath)
		if absErr == nil {
			*targetPath = absPath
		}
		sb = &sandbox.Sandbox{Root: *targetPath}
		result.Root = sb.Root
	}

	// 3. Discovery (Skills + Virtuals)
	if !*jsonMode {
		fmt.Println("--- Running Discovery ---")
	}
	found, err := sb.RunDiscovery()
	if err != nil {
		outputError(err, *jsonMode)
		os.Exit(1)
	}
	virtualPaths := sb.RunVirtualDiscovery()
	
	result.Discovered = found
	result.VirtualPaths = virtualPaths

	if !*jsonMode {
		fmt.Printf("Found %d standard skill files.\n", len(found))
		fmt.Printf("Found %d virtual skill files.\n\n", len(virtualPaths))
	}

	// 4. Parse & Print
	if !*jsonMode {
		fmt.Println("--- Parsing & Inspecting ---")
	}
	
	// Standard Skills
	skills, err := sb.ParseSkills(found)
	if err != nil {
		outputError(err, *jsonMode)
		os.Exit(1)
	}

	// Virtual Skills (simulating TUI behavior)
	for _, vp := range virtualPaths {
		vSk, err := parser.Parse(vp)
		if err == nil {
			vSk.ID = "virtual:" + strings.ToLower(strings.TrimSuffix(filepath.Base(vp), ".md"))
			vSk.Name = "★ " + filepath.Base(vp)
			// Prepend
			skills = append([]*types.Skill{vSk}, skills...)
		}
	}

	result.Skills = sandbox.ToSummary(skills)
	if !*jsonMode {
		sb.PrintResults(skills)
		fmt.Println()
	}

	// 5. Normalization & Bug Verification (Only relevant for Temp/Repro)
	for _, sk := range skills {
		if !*jsonMode && strings.Contains(sk.ID, "\\") {
			fmt.Printf("  [WARNING] ID contains backslashes (not normalized): %s\n", sk.ID)
		}
		if sk.Name == "bug-repro" {
			if sk.RawBody == "Contenido base." {
				result.BugReproResult = "OK"
				if !*jsonMode {
					fmt.Printf("  [OK] Found bug-repro skill with literal placeholder content.\n")
				}
			} else {
				result.BugReproResult = fmt.Sprintf("FAIL: content mismatch: %q", sk.RawBody)
				if !*jsonMode {
					fmt.Printf("  [FAIL] bug-repro skill content mismatch: %q\n", sk.RawBody)
				}
			}
		}
	}

	// 6. Installation Simulation
	if isTemp {
		if !*jsonMode {
			fmt.Println("--- Installer Simulation ---")
		}
		
		// Use skill-creator as source for test
		sourceSkill := filepath.Join(".", ".agents", "skills", "skill-creator", "SKILL.md")
		installErr := sb.SimulateInstall("new-skill", ".agents", sourceSkill)
		simResult := &sandbox.InstallSummary{}
		if installErr != nil {
			simResult.Success = false
			simResult.Error = installErr.Error()
			if !*jsonMode {
				fmt.Printf("Install failed: %v\n", installErr)
			}
		} else {
			simResult.Success = true
			if !*jsonMode {
				fmt.Println("[OK] Installation successful.")
			}
			
			reFound, _ := sb.RunDiscovery()
			foundNew := false
			for _, rf := range reFound {
				if strings.Contains(rf, "new-skill") {
					foundNew = true
					break
				}
			}
			simResult.Discovered = foundNew
			if !*jsonMode {
				if foundNew {
					fmt.Println("[OK] New skill is discoverable.")
				} else {
					fmt.Println("[FAIL] New skill NOT discovered after install.")
				}
			}
		}
		result.InstallResult = simResult
	}

	if *jsonMode {
		if err := sb.PrintJSON(result); err != nil {
			fmt.Printf(`{"error": "Failed to output JSON: %v"}`, err)
			os.Exit(1)
		}
	} else {
		fmt.Println("\n=== Sandbox Run Complete ===")
	}
}

func outputError(err error, jsonMode bool) {
	if jsonMode {
		fmt.Printf(`{"error": "%v"}`, err)
	} else {
		fmt.Printf("Error: %v\n", err)
	}
}
