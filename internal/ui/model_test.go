package ui

import (
	"testing"
	"skillsync/tui/internal/types"
)

func TestItemTitle(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Unix path .agents",
			path:     ".agents/skills/my-skill/SKILL.md",
			expected: "Test Skill [.agents]",
		},
		{
			name:     "Windows path .opencode",
			path:     ".opencode\\skills\\my-skill\\SKILL.md",
			expected: "Test Skill [.opencode]",
		},
		{
			name:     "False positive prevention (substring in name)",
			path:     ".agents/skills/claude-helper/SKILL.md",
			expected: "Test Skill [.agents]",
		},
		{
			name:     "Virtual agent",
			path:     "virtual",
			expected: "Test Skill",
		},
		{
			name:     "No provider directory",
			path:     "src/skills/my-skill/SKILL.md",
			expected: "Test Skill",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sk := types.Skill{
				Name: "Test Skill",
				Path: tt.path,
			}
			if tt.name == "Virtual agent" {
				sk.ID = "virtual:agents"
			}
			
			it := item{skill: sk}
			title := it.Title()
			
			if title != tt.expected {
				t.Errorf("expected title %q, got %q", tt.expected, title)
			}
		})
	}
}