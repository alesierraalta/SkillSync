package ui

import (
	"testing"

	"skillsync/tui/internal/storage"
	"skillsync/tui/internal/types"
)

// BenchmarkDetailView measures allocations per call for detailView.
// After WU-1 (inline style extraction), allocs/op must be lower than before
// because the per-loop lipgloss.Style allocations are hoisted to package-level vars.
// Spec: THEME-6.
func BenchmarkDetailView(b *testing.B) {
	m := NewModel(NewBackend(storage.NewService("")))
	m.Screen = ScreenDetail
	m.selected = &types.Skill{
		Name: "T",
		Metadata: types.Metadata{
			Description: "d",
			Scope:       "project",
		},
	}
	m.setupInputs() // 3 inputs -> loop runs 3x
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.detailView()
	}
}
