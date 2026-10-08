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

package hyperv_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/hyperv"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, hyperv.Name, hyperv.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, hyperv.New, nil)
	testutils.RequireFixtureMetric(t, metrics, hyperv.Name, "windows_hyperv_dynamic_memory_vm_physical_bytes", prometheus.Labels{"vm": "GitHubActions"})
	testutils.RequireFixtureMetric(t, metrics, hyperv.Name, "windows_hyperv_vm_processor_count", prometheus.Labels{"vm": "GitHubActions"})

	// The fixture VM is not replicated, but every VM has a primary replication relationship.
	if metric := testutils.RequireFixtureMetric(t, metrics, hyperv.Name, "windows_hyperv_replica_vm_state", prometheus.Labels{"vm": "GitHubActions", "relationship": "primary"}); metric != nil {
		require.InDelta(t, 0, metric.GetGauge().GetValue(), 0, "expected replication to be disabled")
	}

	// The health query reports failures as 0 instead of an error, so check the value.
	if metric := testutils.RequireFixtureMetric(t, metrics, hyperv.Name, "windows_hyperv_wmi_health", nil); metric != nil {
		require.InDelta(t, 1, metric.GetGauge().GetValue(), 0, "Hyper-V WMI health query failed")
	}
}
