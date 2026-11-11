// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package dfsr_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/dfsr"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, dfsr.Name, dfsr.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, dfsr.New, nil)
}
