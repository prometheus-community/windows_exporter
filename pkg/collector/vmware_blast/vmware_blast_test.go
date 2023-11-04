package vmware_blast_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/vmware_blast"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, vmware_blast.Name, vmware_blast.NewWithFlags)
}
