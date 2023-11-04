package scheduled_task_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/scheduled_task"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, scheduled_task.Name, scheduled_task.NewWithFlags)
}
