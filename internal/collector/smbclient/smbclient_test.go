//go:build windows

package smbclient_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/smbclient"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, smbclient.Name, smbclient.NewWithFlags)
}
