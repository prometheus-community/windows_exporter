//go:build windows

package scheduled_task_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/scheduled_task"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, scheduled_task.Name, scheduled_task.NewWithFlags)
}
