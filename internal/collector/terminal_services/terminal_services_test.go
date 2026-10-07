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

package terminal_services_test

import (
	"os"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/terminal_services"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, terminal_services.Name, terminal_services.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, terminal_services.New, nil)

	userName := os.Getenv("WINDOWS_EXPORTER_TEST_RDP_USER")
	if userName == "" {
		return
	}

	sessionName := os.Getenv("WINDOWS_EXPORTER_TEST_RDP_SESSION")
	require.NotEmpty(t, sessionName, "RDP fixture did not reach an authenticated active session")
	info := testutils.RequireFixtureMetric(t, metrics, terminal_services.Name, "windows_terminal_services_session_info", prometheus.Labels{
		"user":         os.Getenv("COMPUTERNAME") + `\` + userName,
		"session_name": sessionName,
		"state":        "active",
	})
	require.NotNil(t, info)
	require.InDelta(t, 1, info.GetGauge().GetValue(), 0, "RDP fixture is no longer active")
	handles := testutils.RequireFixtureMetric(t, metrics, terminal_services.Name, "windows_terminal_services_handles", prometheus.Labels{"session_name": sessionName})
	require.NotNil(t, handles)
	require.Positive(t, handles.GetGauge().GetValue(), "RDP fixture has no session processes")
}
