package agentdetect

import (
	"testing"
)

// BenchmarkDetect measures the performance of a full Detect() scan against the
// real HOME directory. The performance target is < 100ms per operation (REQ-09).
//
// On CI where HOME is sparse, the benchmark is skipped with an informational message.
// Run locally with: go test -bench=BenchmarkDetect -benchmem ./internal/agentdetect/...
func BenchmarkDetect(b *testing.B) {
	result, err := Detect()
	if err != nil {
		b.Skipf("HOME unresolvable: %v", err)
	}
	if len(result) == 0 {
		b.Skip("sparse HOME — no agents detected, skipping perf gate")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Detect()
	}
}
