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

// Package collector_test contains tests that apply to every collector.
package collector_test

import (
	"log/slog"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/collector/dfsr"
	"github.com/prometheus-community/windows_exporter/internal/collector/dhcp"
	"github.com/prometheus-community/windows_exporter/internal/collector/dns"
	"github.com/prometheus-community/windows_exporter/internal/collector/exchange"
	"github.com/prometheus-community/windows_exporter/internal/collector/hyperv"
	"github.com/prometheus-community/windows_exporter/internal/collector/logical_disk"
	"github.com/prometheus-community/windows_exporter/internal/collector/mscluster"
	"github.com/prometheus-community/windows_exporter/internal/collector/mssql"
	"github.com/prometheus-community/windows_exporter/internal/collector/net"
	"github.com/prometheus-community/windows_exporter/internal/collector/netframework"
	"github.com/prometheus-community/windows_exporter/internal/collector/storage_spaces"
	"github.com/prometheus-community/windows_exporter/internal/collector/tcp"
	"github.com/prometheus-community/windows_exporter/internal/collector/update"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func newMISession(t *testing.T) *mi.Session {
	t.Helper()

	app, err := mi.ApplicationInitialize()
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, app.Close()) })

	session, err := app.NewSession(nil)
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, session.Close()) })

	return session
}

// TestBuildCloseCycle builds, collects and closes every collector repeatedly and
// fails if goroutines grow with every cycle. Grafana Alloy embeds the collectors
// and rebuilds them on every configuration reload, so a leak per cycle grows
// without bound there. Collectors whose Build fails on this host, e.g. because a
// role is not installed, are closed after the failed Build, which must release
// everything a partial Build allocated. The test does not run in parallel,
// because goroutine counts are process-wide.
func TestBuildCloseCycle(t *testing.T) {
	const cycles = 5

	logger := slog.New(slog.DiscardHandler)
	miSession := newMISession(t)

	for _, name := range collector.Available() {
		t.Run(name, func(t *testing.T) {
			if name == update.Name {
				t.Skip("the update worker finishes a running Windows Update search before it exits")
			}

			cycle := func() {
				app := kingpin.New("windows_exporter", "")
				c := collector.BuildersWithFlags[name](app)

				_, err := app.Parse(nil)
				require.NoError(t, err)

				if err := c.Build(logger, miSession); err == nil {
					collect(t, c)
				}

				require.NoError(t, c.Close())
			}

			// The first cycle loads DLLs and may start goroutines that live as long as the process.
			cycle()

			before := runtime.NumGoroutine()

			for range cycles {
				cycle()
			}

			// Closed PDH collectors stop their worker goroutines asynchronously.
			deadline := time.Now().Add(10 * time.Second)

			for runtime.NumGoroutine()-before >= cycles && time.Now().Before(deadline) {
				time.Sleep(50 * time.Millisecond)
			}

			require.Less(t, runtime.NumGoroutine()-before, cycles,
				"goroutines grew with every Build/Close cycle of collector %s", name,
			)
		})
	}
}

func collect(t *testing.T, c collector.Collector) {
	t.Helper()

	ch := make(chan prometheus.Metric)

	var wg sync.WaitGroup

	wg.Go(func() {
		for range ch {
		}
	})

	// Errors are expected for collectors without data on this host.
	_ = c.Collect(ch, 10*time.Second)

	close(ch)
	wg.Wait()
}

func TestBuildRejectsUnknownSubCollector(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	miSession := newMISession(t)
	unknown := []string{"does_not_exist"}

	for _, c := range []collector.Collector{
		dfsr.New(&dfsr.Config{CollectorsEnabled: unknown}),
		dhcp.New(&dhcp.Config{CollectorsEnabled: unknown}),
		dns.New(&dns.Config{CollectorsEnabled: unknown}),
		exchange.New(&exchange.Config{CollectorsEnabled: unknown}),
		hyperv.New(&hyperv.Config{CollectorsEnabled: unknown}),
		logical_disk.New(&logical_disk.Config{CollectorsEnabled: unknown}),
		mscluster.New(&mscluster.Config{CollectorsEnabled: unknown}),
		mssql.New(&mssql.Config{CollectorsEnabled: unknown}),
		net.New(&net.Config{CollectorsEnabled: unknown}),
		netframework.New(&netframework.Config{CollectorsEnabled: unknown}),
		storage_spaces.New(&storage_spaces.Config{CollectorsEnabled: unknown}),
		tcp.New(&tcp.Config{CollectorsEnabled: unknown}),
	} {
		t.Run(c.GetName(), func(t *testing.T) {
			t.Parallel()

			err := c.Build(logger, miSession)
			require.ErrorContains(t, err, "unknown sub collector: does_not_exist. Possible values: ")
			require.NoError(t, c.Close())
		})
	}
}

// Build must not sort the CollectorsEnabled slice of the caller in place.
// It may be the slice of the exported ConfigDefaults.
func TestBuildDoesNotModifyCollectorsEnabled(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)
	miSession := newMISession(t)

	for _, tc := range []struct {
		enabled []string
		new     func(enabled []string) collector.Collector
	}{
		{
			enabled: []string{"Autodiscover", "ActiveSync"},
			new: func(enabled []string) collector.Collector {
				return exchange.New(&exchange.Config{CollectorsEnabled: enabled})
			},
		},
		{
			enabled: []string{"host", "datastore"},
			new: func(enabled []string) collector.Collector {
				return hyperv.New(&hyperv.Config{CollectorsEnabled: enabled})
			},
		},
		{
			enabled: []string{"locks", "info"},
			new: func(enabled []string) collector.Collector {
				return mssql.New(&mssql.Config{CollectorsEnabled: enabled})
			},
		},
		{
			enabled: []string{"clrmemory", "clrexceptions"},
			new: func(enabled []string) collector.Collector {
				return netframework.New(&netframework.Config{CollectorsEnabled: enabled})
			},
		},
	} {
		c := tc.new(tc.enabled)

		t.Run(c.GetName(), func(t *testing.T) {
			t.Parallel()

			want := slices.Clone(tc.enabled)

			// Build errors are expected if the role is not installed.
			_ = c.Build(logger, miSession)

			require.NoError(t, c.Close())
			require.Equal(t, want, tc.enabled)
		})
	}
}

// New must apply the default sub collectors if CollectorsEnabled is nil,
// otherwise library users that don't set it get no metrics.
func TestNewDefaultsCollectorsEnabled(t *testing.T) {
	t.Parallel()

	c := logical_disk.New(&logical_disk.Config{})

	require.NoError(t, c.Build(slog.New(slog.DiscardHandler), nil))

	t.Cleanup(func() { require.NoError(t, c.Close()) })

	ch := make(chan prometheus.Metric)

	var (
		wg    sync.WaitGroup
		descs []string
	)

	wg.Go(func() {
		for metric := range ch {
			descs = append(descs, metric.Desc().String())
		}
	})

	err := c.Collect(ch, 10*time.Second)

	close(ch)
	wg.Wait()

	require.NoError(t, err)
	require.True(t, slices.ContainsFunc(descs, func(desc string) bool {
		return strings.Contains(desc, `fqName: "windows_logical_disk_read_bytes_total"`)
	}), "metrics sub collector is not enabled by default")
}
