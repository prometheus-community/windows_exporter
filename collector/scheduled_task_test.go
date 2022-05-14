package collector

import (
	"testing"
)

func BenchmarkScheduledTaskCollector(b *testing.B) {
	benchmarkCollector(b, "scheduled_task", NewScheduledTask)
}
