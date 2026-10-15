// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package dfsr

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

type metricCollector struct {
	collector   *Collector
	connections []perfDataCounterValuesConnection
	folders     []perfDataCounterValuesFolder
	volumes     []perfDataCounterValuesVolume
}

func (m metricCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(m, ch)
}

func (m metricCollector) Collect(ch chan<- prometheus.Metric) {
	m.collector.collectConnectionMetrics(ch, m.connections)
	m.collector.collectFolderMetrics(ch, m.folders)
	m.collector.collectVolumeMetrics(ch, m.volumes)
}

func TestCollectMetrics(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.buildDescriptors()

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(metricCollector{
		collector: c,
		connections: []perfDataCounterValuesConnection{{
			Name: "conn",
			BandwidthSavingsUsingDFSReplicationPercent: 20,
			BytesReceivedTotal:                         4096,
			RdcCompressedSizeOfFilesReceivedTotal:      1024,
		}},
		folders: []perfDataCounterValuesFolder{{
			Name: "SYSVOL Share-{DD658FCA-D7AE-487D-8E75-67481CEBE108}",
			BandwidthSavingsUsingDFSReplicationPercent: 75,
			StagingSpaceInUse:                          2048,
			StagingBytesGeneratedTotal:                 8192,
		}},
		volumes: []perfDataCounterValuesVolume{{
			Name:                       "C:",
			DatabaseCommitsTotal:       800998,
			UsnJournalRecordsReadTotal: 1932712,
			UsnJournalUnreadPercentage: 40,
		}},
	}))

	families, err := registry.Gather()
	require.NoError(t, err)
	require.Len(t, families, 9+27+5)

	expected := map[string]struct {
		metricType dto.MetricType
		value      float64
	}{
		"windows_dfsr_connection_bandwidth_savings_using_dfs_replication_ratio":     {dto.MetricType_GAUGE, 0.2},
		"windows_dfsr_connection_bytes_received_total":                              {dto.MetricType_COUNTER, 4096},
		"windows_dfsr_connection_rdc_compressed_size_of_received_files_bytes_total": {dto.MetricType_COUNTER, 1024},
		"windows_dfsr_folder_bandwidth_savings_using_dfs_replication_ratio":         {dto.MetricType_GAUGE, 0.75},
		"windows_dfsr_folder_staging_space_in_use_bytes":                            {dto.MetricType_GAUGE, 2048},
		"windows_dfsr_folder_staging_generated_bytes_total":                         {dto.MetricType_COUNTER, 8192},
		"windows_dfsr_volume_database_commits_total":                                {dto.MetricType_COUNTER, 800998},
		"windows_dfsr_volume_usn_journal_read_records_total":                        {dto.MetricType_COUNTER, 1932712},
		"windows_dfsr_volume_usn_journal_unread_ratio":                              {dto.MetricType_GAUGE, 0.4},
	}

	seen := 0

	for _, family := range families {
		name := family.GetName()

		// Counters and only counters end in _total.
		if strings.HasSuffix(name, "_total") {
			require.Equal(t, dto.MetricType_COUNTER, family.GetType(), name)
		} else {
			require.Equal(t, dto.MetricType_GAUGE, family.GetType(), name)
		}

		require.NotContains(t, name, "percent", name)
		require.Len(t, family.GetMetric(), 1, name)

		want, ok := expected[name]
		if !ok {
			continue
		}

		seen++

		metric := family.GetMetric()[0]
		require.Equal(t, want.metricType, family.GetType(), name)

		actual := metric.GetGauge().GetValue()
		if want.metricType == dto.MetricType_COUNTER {
			actual = metric.GetCounter().GetValue()
		}

		require.InDelta(t, want.value, actual, 1e-12, name)
	}

	require.Equal(t, len(expected), seen)
}
