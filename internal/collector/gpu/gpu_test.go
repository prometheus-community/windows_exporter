// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package gpu_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/gpu"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, gpu.Name, gpu.NewWithFlags)
}

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, gpu.New, nil)
}
