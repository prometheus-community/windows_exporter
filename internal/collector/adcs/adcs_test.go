// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package adcs_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/adcs"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, adcs.Name, adcs.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, adcs.New, nil)
}
