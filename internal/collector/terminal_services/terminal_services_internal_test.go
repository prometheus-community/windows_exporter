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

package terminal_services

import (
	"log/slog"
	"strings"
	"testing"
	"time"

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

		require.NoError(t, c.Collect(ch, time.Second))
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
