// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package fsrmquota

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestQuotaTemplateMetric(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.buildDescriptors()

	metrics := make(chan prometheus.Metric, 20)
	c.emitQuotas(metrics, []msftFSRMQuota{{Path: "C:\\data", Template: "Standard"}})
	close(metrics)

	found := false

	for metric := range metrics {
		if metric.Desc() != c.template {
			continue
		}

		var value dto.Metric
		require.NoError(t, metric.Write(&value))
		require.InDelta(t, 1, value.GetGauge().GetValue(), 1e-9)

		labels := make(map[string]string)
		for _, label := range value.GetLabel() {
			labels[label.GetName()] = label.GetValue()
		}

		require.Equal(t, map[string]string{"path": "C:\\data", "template": "Standard"}, labels)

		found = true
	}

	require.True(t, found)
}
