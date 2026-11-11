// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package iis_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/iis"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, iis.Name, iis.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, iis.New, nil)
	testutils.RequireFixtureMetric(t, metrics, iis.Name, "windows_iis_requests_total", prometheus.Labels{"site": "Default Web Site", "method": "get"})
}
