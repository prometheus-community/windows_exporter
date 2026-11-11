// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package fsrmquota_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/fsrmquota"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, fsrmquota.Name, fsrmquota.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, fsrmquota.New, nil)
	testutils.RequireFixtureMetric(t, metrics, fsrmquota.Name, "windows_fsrmquota_size_bytes", prometheus.Labels{"path": `C:\fsrm-ci`})
}
