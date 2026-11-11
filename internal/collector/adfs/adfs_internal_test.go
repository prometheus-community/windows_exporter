// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package adfs

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

// TestCollectDatabaseQueryTimes checks that the database query times, which
// internal/pdh converts to seconds, and the query counts are exported.
func TestCollectDatabaseQueryTimes(t *testing.T) {
	t.Parallel()

	data := perfDataCounterValues{
		ArtifactDBQueryTimeSeconds: 0.25,
		ArtifactDBQueries:          5,
		ConfigDBQueryTimeSeconds:   1.1521176,
		ConfigDBQueries:            306,
	}

	expected := map[string]float64{
		"windows_adfs_db_artifact_query_time_seconds_total": 0.25,
		"windows_adfs_db_artifact_queries_total":            5,
		"windows_adfs_db_config_query_time_seconds_total":   1.1521176,
		"windows_adfs_db_config_queries_total":              306,
	}

	c := New(nil)
	c.buildDescriptors()

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(metricCollector{collector: c, data: data}))

	families, err := registry.Gather()
	require.NoError(t, err)

	byName := make(map[string]*dto.MetricFamily, len(families))
	for _, family := range families {
		byName[family.GetName()] = family
	}

	for name, value := range expected {
		family, ok := byName[name]
		require.True(t, ok, "missing metric %s", name)
		require.Equal(t, dto.MetricType_COUNTER, family.GetType(), name)
		require.Len(t, family.GetMetric(), 1, name)
		require.InDelta(t, value, family.GetMetric()[0].GetCounter().GetValue(), 0, name)
	}
}

type metricCollector struct {
	collector *Collector
	data      perfDataCounterValues
}

func (c metricCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c metricCollector) Collect(ch chan<- prometheus.Metric) {
	c.collector.collectMetrics(ch, c.data)
}
