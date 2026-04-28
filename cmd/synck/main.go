package main

import (
	"flag"
	"fmt"
	"os"
	"skillsync/tui/internal/install"
	"skillsync/tui/internal/ui"
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

func run(args []string) error {
	f := flag.NewFlagSet(args[0], flag.ContinueOnError)
	globalInstall := f.Bool("g", false, "Install synck globally")

	// Handle 'install -g' sub-style
	actualArgs := args[1:]
	if len(actualArgs) > 0 && actualArgs[0] == "install" {
		actualArgs = actualArgs[1:]
	}

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
