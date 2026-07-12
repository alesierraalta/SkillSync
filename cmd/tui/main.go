package main

import (
	"fmt"
	"os"
	"skillsync/tui/internal/ui"
)

// uiRun is a seam for testing ui.Run() calls.
// It can be swapped out in tests to inject errors or control behavior.
var uiRun = func() error {
	return ui.Run()
}

// osExit is a seam for testing os.Exit() calls.
// It can be swapped out in tests to capture exit codes without actually exiting.
var osExit = func(code int) {
	os.Exit(code)
}

func main() {
	if err := uiRun(); err != nil {
		fmt.Printf("Error: %v\n", err)
		osExit(1)
	}
}
