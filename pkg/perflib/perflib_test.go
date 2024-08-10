package perflib

import (
	"testing"
)

func BenchmarkQueryPerformanceData(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = QueryPerformanceData("Global")
	}
}
