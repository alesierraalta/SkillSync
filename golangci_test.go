package root_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestGolangciConfig_Contract asserts .golangci.yml exists and
// contains the spec-required keys (AUDIT-12a; v2-schema
// `exclude-default: false` is the equivalent of the v1 spec's
// `exclude-use-default: false`).
func TestGolangciConfig_Contract(t *testing.T) {
	raw, err := os.ReadFile(".golangci.yml")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	body := string(raw)
	for _, want := range []string{
		"version:", "timeout: 5m", "linters:",
		"- govet", "- errcheck", "- gofmt", "- ineffassign", "- unused",
		"issues:", "exclude-default: false",
	} {
		if !strings.Contains(body, want) {
			t.Errorf(".golangci.yml missing %q", want)
		}
	}
}

// TestGolangciConfig_VerifyWithBinary: AUDIT-12a real-shell check.
func TestGolangciConfig_VerifyWithBinary(t *testing.T) {
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		t.Skip("golangci-lint not on PATH; TestGolangciConfig_Contract covers AUDIT-12a")
	}
	if out, err := exec.Command("golangci-lint", "config", "verify").CombinedOutput(); err != nil {
		t.Fatalf("golangci-lint config verify: %v\n%s", err, out)
	}
}
