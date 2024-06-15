package smbclient_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/pkg/collector/smbclient"
	"github.com/prometheus-community/windows_exporter/pkg/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, smbclient.Name, smbclient.NewWithFlags)
}
