package registry

import (
	"testing"
)

func BenchmarkQueryPerformanceData(b *testing.B) {
	for b.Loop() {
		_, _ = QueryPerformanceData("Global", "")
	}
}
