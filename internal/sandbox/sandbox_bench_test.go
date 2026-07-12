package sandbox_test

import (
	"testing"

	"skillsync/tui/internal/sandbox"
)

func BenchmarkSimulateInstall(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sb, err := sandbox.New()
		if err != nil {
			b.Fatal(err)
		}
		if err := sb.CreateFixtures(sandbox.DefaultFixtures); err != nil {
			_ = sb.Cleanup()
			b.Fatal(err)
		}
		if err := sb.SimulateInstall("deploy", ".agents", ".agents/skills/deploy/SKILL.md"); err != nil {
			_ = sb.Cleanup()
			b.Fatal(err)
		}
		_ = sb.Cleanup()
	}
}
