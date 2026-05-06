package diff

import (
	"fmt"
	"strings"
	"testing"
)

func TestUnifiedDiff(t *testing.T) {
	tests := []struct {
		name        string
		before      string
		after       string
		cap         int
		wantDiff    string
		wantSummary string
	}{
		{
			name:        "identical",
			before:      "line1\nline2\n",
			after:       "line1\nline2\n",
			cap:         50,
			wantDiff:    "",
			wantSummary: "",
		},
		{
			name:        "empty both",
			before:      "",
			after:       "",
			cap:         50,
			wantDiff:    "",
			wantSummary: "",
		},
		{
			name:        "added lines",
			before:      "",
			after:       "line1\nline2\n",
			cap:         50,
			wantDiff:    "@@ -0,0 +1,2 @@\n+line1\n+line2",
			wantSummary: "+2",
		},
		{
			name:        "removed lines",
			before:      "line1\nline2\n",
			after:       "",
			cap:         50,
			wantDiff:    "@@ -1,2 +0,0 @@\n-line1\n-line2",
			wantSummary: "-2",
		},
		{
			name:        "mixed changes",
			before:      "line1\nline2\nline3\n",
			after:       "line1\nline2modified\nline3\n",
			cap:         50,
			wantDiff:    "@@ -1,3 +1,3 @@\n line1\n-line2\n+line2modified\n line3",
			wantSummary: "+1 -1",
		},
		{
			name:        "single line replaced",
			before:      "hello\n",
			after:       "world\n",
			cap:         50,
			wantDiff:    "@@ -1,1 +1,1 @@\n-hello\n+world",
			wantSummary: "+1 -1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDiff, gotSummary := UnifiedDiff(tt.before, tt.after, tt.cap)
			if gotDiff != tt.wantDiff {
				t.Errorf("diff = %q, want %q", gotDiff, tt.wantDiff)
			}
			if gotSummary != tt.wantSummary {
				t.Errorf("summary = %q, want %q", gotSummary, tt.wantSummary)
			}
		})
	}
}

func TestUnifiedDiff_Cap(t *testing.T) {
	before := "a\nb\nc\nd\ne\nf\ng\nh\ni\nj\n"
	after := "a\nb\nc\nd\nX\nf\ng\nh\ni\nj\n"

	gotDiff, gotSummary := UnifiedDiff(before, after, 5)

	if !strings.Contains(gotDiff, "... (") {
		t.Errorf("expected truncated diff to contain '... (', got:\n%s", gotDiff)
	}

	lineCount := len(strings.Split(gotDiff, "\n"))
	if lineCount > 5 {
		t.Errorf("diff has %d lines, want <= 5", lineCount)
	}

	if gotSummary != "+1 -1" {
		t.Errorf("summary = %q, want +1 -1", gotSummary)
	}
}

func TestUnifiedDiff_CapAllAdded(t *testing.T) {
	var after strings.Builder
	for i := 1; i <= 60; i++ {
		after.WriteString(fmt.Sprintf("line%d\n", i))
	}

	gotDiff, gotSummary := UnifiedDiff("", after.String(), 50)

	if !strings.Contains(gotDiff, "... (") {
		t.Errorf("expected truncated diff to contain '... (', got:\n%s", gotDiff)
	}

	lineCount := len(strings.Split(gotDiff, "\n"))
	if lineCount > 50 {
		t.Errorf("diff has %d lines, want <= 50", lineCount)
	}

	if gotSummary != "+60" {
		t.Errorf("summary = %q, want +60", gotSummary)
	}
}

func TestUnifiedDiff_CapAllRemoved(t *testing.T) {
	var before strings.Builder
	for i := 1; i <= 60; i++ {
		before.WriteString(fmt.Sprintf("line%d\n", i))
	}

	gotDiff, gotSummary := UnifiedDiff(before.String(), "", 50)

	if !strings.Contains(gotDiff, "... (") {
		t.Errorf("expected truncated diff to contain '... (', got:\n%s", gotDiff)
	}

	lineCount := len(strings.Split(gotDiff, "\n"))
	if lineCount > 50 {
		t.Errorf("diff has %d lines, want <= 50", lineCount)
	}

	if gotSummary != "-60" {
		t.Errorf("summary = %q, want -60", gotSummary)
	}
}
