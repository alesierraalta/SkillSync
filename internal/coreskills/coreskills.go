package coreskills

import "embed"

// EmbeddedSkills is a manual build-time snapshot of the core skills from
// .agents/skills/{find-skills,skill-creator,skill-sync}. This embed must be
// kept in sync manually whenever the source skills change. Consistency is
// enforced by TestCoreSkillDrift in drift_test.go, which fails CI if the
// embedded copy diverges from the source.
//go:embed all:skills
var EmbeddedSkills embed.FS
