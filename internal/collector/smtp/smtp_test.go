// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package smtp_test

import (
	"net"
	netsmtp "net/smtp"
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/collector/smtp"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, smtp.Name, smtp.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, smtp.New, nil)
	testutils.RequireFixtureMetric(t, metrics, smtp.Name, "windows_smtp_inbound_connections_total", nil)

	// Exercise the configured virtual server, including its SMTP greeting.
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(t.Context(), "tcp", "127.0.0.1:25")
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	require.NoError(t, conn.SetDeadline(time.Now().Add(5*time.Second)))
	client, err := netsmtp.NewClient(conn, "localhost")
	require.NoError(t, err)
	require.NoError(t, client.Quit())
}
