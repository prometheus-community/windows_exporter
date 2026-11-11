// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package system_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/system"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, system.Name, system.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, system.New, nil)
}
