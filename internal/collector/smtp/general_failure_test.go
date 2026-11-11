// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package smtp

import (
	"log/slog"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestGeneralFailureMetric(t *testing.T) {
	t.Parallel()

	c := New(nil)
	// Descriptors are initialized before connecting to the optional SMTP provider.
	_ = c.Build(slog.New(slog.DiscardHandler), nil)
	if c.perfDataCollector != nil {
		c.perfDataCollector.Close()
	}

	c.perfDataCollector = generalFailureFixture{}
	metrics := make(chan prometheus.Metric, 100)
	require.NoError(t, c.Collect(metrics, 0))
	close(metrics)

	found := false

	for metric := range metrics {
		if metric.Desc() != c.badMailedMessagesGeneralFailureTotal {
			continue
		}

		var value dto.Metric
		require.NoError(t, metric.Write(&value))
		require.InDelta(t, 42, value.GetCounter().GetValue(), 1e-9)
		require.Equal(t, "mail", value.GetLabel()[0].GetValue())

		found = true
	}

	require.True(t, found)
}

type generalFailureFixture struct{}

func (generalFailureFixture) Collect(dst *[]perfDataCounterValues) error {
	*dst = []perfDataCounterValues{{Name: "mail", BadmailedMessagesGeneralFailureTotal: 42}}

	return nil
}
func (generalFailureFixture) Close() {}

func TestCloseBeforeProviderInitialization(t *testing.T) {
	t.Parallel()
	require.NoError(t, New(nil).Close())
}
