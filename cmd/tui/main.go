package main

import (
	"fmt"
	"os"
	"skillsync/tui/internal/ui"


)

func main() {
	if err := ui.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
