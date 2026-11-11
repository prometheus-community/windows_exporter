// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package remote_fx_test

import (
	"os"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/remote_fx"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, remote_fx.Name, remote_fx.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, remote_fx.New, nil)
	labels := prometheus.Labels{"session_name": os.Getenv("WINDOWS_EXPORTER_TEST_RDP_SESSION")}
	testutils.RequireFixtureMetric(t, metrics, remote_fx.Name, "windows_remote_fx_net_sent_bytes_total", labels)
	testutils.RequireFixtureMetric(t, metrics, remote_fx.Name, "windows_remote_fx_gfx_output_frames_total", labels)
}
