// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package cache

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
)

func TestCopyReadHitCounterWireFormats(t *testing.T) {
	t.Parallel()

	c := New(nil)
	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))
	c.perfDataCollector.Close()
	c.perfDataCollector = copyReadFixture{}
	wrapper := &copyReadTestCollector{collector: c}
	registry := prometheus.NewRegistry()
	registry.MustRegister(wrapper)
	families, err := registry.Gather()
	require.NoError(t, err)
	require.NoError(t, wrapper.err)

	found := false

	for _, family := range families {
		if family.GetName() != "windows_cache_copy_read_hits_total" {
			continue
		}

		found = true

		require.Equal(t, dto.MetricType_COUNTER, family.GetType())
		require.InDelta(t, 25, family.GetMetric()[0].GetCounter().GetValue(), 1e-9)

		var text strings.Builder

		_, err := expfmt.MetricFamilyToText(&text, family)
		require.NoError(t, err)
		require.Contains(t, text.String(), "# TYPE windows_cache_copy_read_hits_total counter")

		var openMetrics strings.Builder

		_, err = expfmt.MetricFamilyToOpenMetrics(&openMetrics, family)
		require.NoError(t, err)
		require.Contains(t, openMetrics.String(), "# TYPE windows_cache_copy_read_hits counter")
		require.Contains(t, openMetrics.String(), "windows_cache_copy_read_hits_total 25")
	}

	require.True(t, found)
}

type copyReadFixture struct{}

func (copyReadFixture) Collect(dst *[]perfDataCounterValues) error {
	*dst = []perfDataCounterValues{{CopyReadHitsTotal: 25}}

	return nil
}

func (copyReadFixture) Close() {}

type copyReadTestCollector struct {
	collector *Collector
	err       error
}

func (c *copyReadTestCollector) Describe(_ chan<- *prometheus.Desc) {}

func (c *copyReadTestCollector) Collect(ch chan<- prometheus.Metric) {
	c.err = c.collector.Collect(ch, 0)
}

func TestCopyReadCollectorCloseBeforeInitialization(t *testing.T) {
	t.Parallel()
	require.NoError(t, New(nil).Close())
}
