// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package smbclient_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/smbclient"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, smbclient.Name, smbclient.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, smbclient.New, nil)
	testutils.RequireFixtureMetric(t, metrics, smbclient.Name, "windows_smbclient_data_queue_seconds_total", prometheus.Labels{"share": "CIShare"})
}
