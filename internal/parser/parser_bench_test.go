package parser_test

import (
	"testing"

	"skillsync/tui/internal/parser"
	"skillsync/tui/internal/types"
)

const benchSkillContent = `---
name: bench-skill
description: A skill used for benchmark testing.
---
# Bench Skill

This is the body of the bench skill used in benchmarks.

## Usage

Call this skill to do things.
`

func BenchmarkParseContent(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = parser.ParseContent(benchSkillContent)
	}
}

func BenchmarkFormat(b *testing.B) {
	b.ReportAllocs()
	skill := &types.Skill{
		Name: "bench-skill",
		Metadata: types.Metadata{
			Description: "A skill used for benchmark testing.",
		},
		RawBody: "# Bench Skill\n\nThis is the body of the bench skill used in benchmarks.\n",
	}
	for i := 0; i < b.N; i++ {
		_, _ = parser.Format(skill)
	}
}
