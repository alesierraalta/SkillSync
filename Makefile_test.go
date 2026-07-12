package root_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestMakefile_TargetsExist asserts the five required targets exist,
// verify depends on test/lint/build, and the lint target has the
// graceful-fallback note (AUDIT-11a, AUDIT-12b; S-11, S-14).
func TestMakefile_TargetsExist(t *testing.T) {
	raw, err := os.ReadFile("Makefile")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	body := string(raw)
	for _, target := range []string{"test:", "bench:", "lint:", "build:", "verify:"} {
		if !strings.Contains(body, target) {
			t.Errorf("missing target %q", target)
		}
	}
	for _, line := range strings.Split(body, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "verify:") {
			continue
		}
		for _, dep := range []string{"test", "lint", "build"} {
			if !strings.Contains(line, dep) {
				t.Errorf("verify missing dep %q in %q", dep, line)
			}
		}
	}
	if !strings.Contains(body, "go test") {
		t.Error("Makefile does not invoke `go test`")
	}
	if !strings.Contains(body, "go build") {
		t.Error("Makefile does not invoke `go build`")
	}
	if !strings.Contains(body, "golangci-lint not installed") {
		t.Error("lint target missing graceful-fallback note")
	}
	if !strings.Contains(body, "&& true") {
		t.Error("lint target missing `&& true` exit-0 chain")
	}
}

// TestMakefile_VerifyDryRun: S-12 end-to-end where `make` is on PATH.
func TestMakefile_VerifyDryRun(t *testing.T) {
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("`make` not on PATH; TestMakefile_TargetsExist covers S-12")
	}
	out, err := exec.Command("make", "-n", "verify").CombinedOutput()
	if err != nil {
		t.Fatalf("make -n verify: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "go test") {
		t.Errorf("make -n verify missing `go test`:\n%s", out)
	}
}
