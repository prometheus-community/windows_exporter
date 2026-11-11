// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package terminal_services

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/scrapecache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCollectorWithoutSessionObject covers hosts without the Terminal Services
// Session performance object, e.g. Server Core installations: the collector
// must build and scrape without errors, and skip only the per-session
// process metrics.
func TestCollectorWithoutSessionObject(t *testing.T) {
	original := terminalServicesSessionObject
	terminalServicesSessionObject = "windows_exporter missing object"

	t.Cleanup(func() { terminalServicesSessionObject = original })

	c := New(nil)

	t.Cleanup(func() { assert.NoError(t, c.Close()) })

	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))
	require.Nil(t, c.perfDataCollectorTerminalServicesSession)

	for range 2 {
		ch := make(chan prometheus.Metric, 1024)

		cacheScrape := scrapecache.Begin(c)
		err := c.Collect(ch, time.Second)
		cacheScrape.End(err)
		require.NoError(t, err)
		close(ch)

		for metric := range ch {
			desc := metric.Desc().String()
			assert.NotContains(t, desc, `"windows_terminal_services_handles"`)
			assert.True(t,
				strings.Contains(desc, `"windows_terminal_services_session_info"`) ||
					strings.Contains(desc, `"windows_terminal_services_connection_broker_`),
				"unexpected metric %s", desc)
		}
	}
}
