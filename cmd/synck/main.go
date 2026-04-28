package main

import (
	"flag"
	"fmt"
	"os"
	"skillsync/tui/internal/install"
	"skillsync/tui/internal/ui"
)

func main() {
	globalInstall := flag.Bool("g", false, "Install synck globally")
	
	// Handle 'install -g' sub-style if needed
	if len(os.Args) > 1 && os.Args[1] == "install" {
		// Shift args for flag parser
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	flag.Parse()

	if *globalInstall {
		fmt.Println("🚀 Installing synck globally...")
		result := install.GlobalInstall()
		if !result.Success {
			fmt.Printf("❌ %s\n", result.Message)
			if result.Error != nil {
				fmt.Printf("Error: %v\n", result.Error)
			}
			os.Exit(1)
		}
		fmt.Println(result.Message)
		return
	}

	// Normal TUI launch
	if err := ui.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
