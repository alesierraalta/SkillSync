package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanProjects_Markers(t *testing.T) {
	// 2.1 RED: Test discovery.ScanProjects for marker detection (.agents, .opencode, AGENTS.md).
	tmpDir := t.TempDir()

	// p1: .agents marker
	p1 := filepath.Join(tmpDir, "p1")
	os.MkdirAll(filepath.Join(p1, ".agents"), 0755)

	// p2: .opencode marker
	p2 := filepath.Join(tmpDir, "p2")
	os.MkdirAll(filepath.Join(p2, ".opencode"), 0755)

	// p3: AGENTS.md marker
	p3 := filepath.Join(tmpDir, "p3")
	os.MkdirAll(p3, 0755)
	os.WriteFile(filepath.Join(p3, "AGENTS.md"), []byte("test"), 0644)

	// p4: No marker
	p4 := filepath.Join(tmpDir, "p4")
	os.MkdirAll(p4, 0755)

	projects, err := ScanProjects([]string{tmpDir}, 2)
	if err != nil {
		t.Fatal(err)
	}

	foundP1, foundP2, foundP3, foundP4 := false, false, false, false
	for _, p := range projects {
		switch p {
		case p1:
			foundP1 = true
		case p2:
			foundP2 = true
		case p3:
			foundP3 = true
		case p4:
			foundP4 = true
		}
	}

	if !foundP1 {
		t.Error("expected to find p1 (.agents)")
	}
	if !foundP2 {
		t.Error("expected to find p2 (.opencode)")
	}
	if !foundP3 {
		t.Error("expected to find p3 (AGENTS.md)")
	}
	if foundP4 {
		t.Error("did not expect to find p4 (no marker)")
	}
}

func TestScanProjects_DepthAndExclusion(t *testing.T) {
	// 2.3 RED: Test ScanProjects depth limiting and directory exclusions (node_modules, .git).
	tmpDir := t.TempDir()

	// p1: depth 1
	p1 := filepath.Join(tmpDir, "p1")
	os.MkdirAll(filepath.Join(p1, ".agents"), 0755)

	// p2: depth 3
	p2 := filepath.Join(tmpDir, "sub1", "sub2", "p2")
	os.MkdirAll(filepath.Join(p2, ".agents"), 0755)

	// p3: depth 4 (beyond limit)
	p3 := filepath.Join(tmpDir, "sub1", "sub2", "sub3", "p3")
	os.MkdirAll(filepath.Join(p3, ".agents"), 0755)

	// p4: in node_modules
	p4 := filepath.Join(tmpDir, "node_modules", "p4")
	os.MkdirAll(filepath.Join(p4, ".agents"), 0755)

	projects, err := ScanProjects([]string{tmpDir}, 3)
	if err != nil {
		t.Fatal(err)
	}

	foundP1, foundP2, foundP3, foundP4 := false, false, false, false
	for _, p := range projects {
		if p == p1 {
			foundP1 = true
		}
		if p == p2 {
			foundP2 = true
		}
		if p == p3 {
			foundP3 = true
		}
		if p == p4 {
			foundP4 = true
		}
	}

	if !foundP1 {
		t.Error("expected to find p1 (depth 1)")
	}
	if !foundP2 {
		t.Error("expected to find p2 (depth 3)")
	}
	if foundP3 {
		t.Error("did not expect to find p3 (depth 4 > limit 3)")
	}
	if foundP4 {
		t.Error("did not expect to find p4 (inside node_modules)")
	}
}
