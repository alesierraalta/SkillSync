package diff

import (
	"fmt"
	"strings"
)

func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func applyCap(lines []string, cap int) []string {
	if cap > 0 && len(lines) > cap {
		available := cap - 2
		if available < 0 {
			available = 0
		}
		top := available / 2
		bottom := available - top

		newLines := make([]string, 0, cap)
		newLines = append(newLines, lines[0])
		newLines = append(newLines, lines[1:1+top]...)
		if len(lines)-1 > top+bottom {
			newLines = append(newLines, fmt.Sprintf("... (%d more lines) ...", len(lines)-1-top-bottom))
		}
		newLines = append(newLines, lines[len(lines)-bottom:]...)
		lines = newLines
	}
	return lines
}

func UnifiedDiff(before, after string, cap int) (string, string) {
	bLines := splitLines(before)
	aLines := splitLines(after)

	// Identical
	if len(bLines) == len(aLines) {
		same := true
		for i := range bLines {
			if bLines[i] != aLines[i] {
				same = false
				break
			}
		}
		if same {
			return "", ""
		}
	}

	// All added
	if before == "" {
		lines := make([]string, 0, len(aLines)+1)
		lines = append(lines, fmt.Sprintf("@@ -0,0 +1,%d @@", len(aLines)))
		for _, l := range aLines {
			lines = append(lines, "+"+l)
		}
		lines = applyCap(lines, cap)
		return strings.Join(lines, "\n"), fmt.Sprintf("+%d", len(aLines))
	}

	// All removed
	if after == "" {
		lines := make([]string, 0, len(bLines)+1)
		lines = append(lines, fmt.Sprintf("@@ -1,%d +0,0 @@", len(bLines)))
		for _, l := range bLines {
			lines = append(lines, "-"+l)
		}
		lines = applyCap(lines, cap)
		return strings.Join(lines, "\n"), fmt.Sprintf("-%d", len(bLines))
	}

	// Find common prefix
	prefix := 0
	for prefix < len(bLines) && prefix < len(aLines) && bLines[prefix] == aLines[prefix] {
		prefix++
	}

	// Find common suffix
	suffix := 0
	for suffix < len(bLines)-prefix && suffix < len(aLines)-prefix &&
		bLines[len(bLines)-1-suffix] == aLines[len(aLines)-1-suffix] {
		suffix++
	}

	// Build hunk with context
	ctxLines := 3
	ctxStart := 0
	if prefix > ctxLines {
		ctxStart = prefix - ctxLines
	}

	bCtxEnd := len(bLines)
	if suffix > 0 && len(bLines)-suffix+ctxLines < bCtxEnd {
		bCtxEnd = len(bLines) - suffix + ctxLines
	}

	aCtxEnd := len(aLines)
	if suffix > 0 && len(aLines)-suffix+ctxLines < aCtxEnd {
		aCtxEnd = len(aLines) - suffix + ctxLines
	}

	removedStart := ctxStart + 1
	removedCount := bCtxEnd - ctxStart
	addedStart := ctxStart + 1
	addedCount := aCtxEnd - ctxStart

	var lines []string
	lines = append(lines, fmt.Sprintf("@@ -%d,%d +%d,%d @@", removedStart, removedCount, addedStart, addedCount))

	added := 0
	removed := 0

	// Context before change
	for i := ctxStart; i < prefix; i++ {
		lines = append(lines, " "+bLines[i])
	}

	// Removed lines
	for i := prefix; i < len(bLines)-suffix; i++ {
		lines = append(lines, "-"+bLines[i])
		removed++
	}

	// Added lines
	for i := prefix; i < len(aLines)-suffix; i++ {
		lines = append(lines, "+"+aLines[i])
		added++
	}

	// Context after change
	for i := len(bLines) - suffix; i < bCtxEnd; i++ {
		lines = append(lines, " "+bLines[i])
	}

	// Apply cap
	lines = applyCap(lines, cap)

	summary := ""
	if added > 0 && removed > 0 {
		summary = fmt.Sprintf("+%d -%d", added, removed)
	} else if added > 0 {
		summary = fmt.Sprintf("+%d", added)
	} else if removed > 0 {
		summary = fmt.Sprintf("-%d", removed)
	}

	return strings.Join(lines, "\n"), summary
}
