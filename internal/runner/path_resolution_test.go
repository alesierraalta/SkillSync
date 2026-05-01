package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestExecuteSync_PathResolutionRegression(t *testing.T) {
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("sh"); err != nil {
			if _, err := exec.LookPath("bash"); err != nil {
				t.Skip("sh or bash not found, skipping .sh path resolution test")
			}
		}
	}

	tmpDir := t.TempDir()

	// Create structure
	skillsDir := filepath.Join(tmpDir, ".agents", "skills")
	libDir := filepath.Join(skillsDir, "lib")
	assetsDir := filepath.Join(skillsDir, "skill-sync", "assets")

	if err := os.MkdirAll(libDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatal(err)
	}

	utilsContent := `#!/usr/bin/env bash
export UTILS_LOADED=true
`
	if err := os.WriteFile(filepath.Join(libDir, "utils.sh"), []byte(utilsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Real logic from sync.sh
	syncContent := `#!/usr/bin/env bash
SCRIPT_PATH="${BASH_SOURCE[0]:-$0}"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
LIB_DIR="$SCRIPT_DIR"
PREV_DIR=""
while [ "$LIB_DIR" != "/" ] && [ "$LIB_DIR" != "." ] && [ "$LIB_DIR" != "$PREV_DIR" ] && [ ! -f "$LIB_DIR/lib/utils.sh" ] && [ ! -f "$LIB_DIR/utils.sh" ]; do
    PREV_DIR="$LIB_DIR"
    LIB_DIR=$(dirname "$LIB_DIR")
done

if [ -f "$LIB_DIR/lib/utils.sh" ]; then
    source "$LIB_DIR/lib/utils.sh"
elif [ -f "$LIB_DIR/utils.sh" ]; then
    source "$LIB_DIR/utils.sh"
else
    echo "Error: Could not find lib/utils.sh"
    exit 1
fi

if [ "$UTILS_LOADED" = "true" ]; then
    echo "SUCCESS"
else
    echo "UTILS NOT LOADED"
    exit 1
fi
`
	syncPath := filepath.Join(assetsDir, "sync.sh")
	if err := os.WriteFile(syncPath, []byte(syncContent), 0755); err != nil {
		t.Fatal(err)
	}

	// Change to project root for the test (tmpDir)
	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	// Runner uses relative path from root
	relSyncPath := filepath.Join(".agents", "skills", "skill-sync", "assets", "sync.sh")
	runner := NewRunner(relSyncPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := runner.ExecuteSync(ctx, nil)
	result := <-ch

	if result.ExitCode != 0 {
		t.Logf("Stderr: %s", result.Stderr)
		t.Logf("Stdout: %s", result.Stdout)
		t.Errorf("Sync failed with exit code %d", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, "SUCCESS") {
		t.Errorf("Expected 'SUCCESS' in stdout, got %q", result.Stdout)
	}
}
