// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package vmware_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/vmware"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, vmware.Name, vmware.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, vmware.New, nil)
}
