// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package smb_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/smb"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, smb.Name, smb.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, smb.New, nil)
	if testutils.RequireFixtureMetric(t, metrics, smb.Name, "windows_smb_server_shares_received_bytes_total", prometheus.Labels{"share": `\\*\CIShare`}) == nil {
		return
	}

	for _, name := range []string{
		"windows_smb_server_shares_open_files",
		"windows_smb_server_shares_tree_connects",
	} {
		require.Contains(t, metrics, name)
		require.Equal(t, dto.MetricType_GAUGE, metrics[name].GetType(), name)
	}
}
