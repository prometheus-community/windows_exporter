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

package mscluster_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/mscluster"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	"github.com/prometheus/client_golang/prometheus"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, mscluster.Name, mscluster.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, mscluster.New, nil)

	// Resource fixtures from the CI cluster setup cover each resource state.
	for name, state := range map[string]float64{
		"CI IP Address":         2, // Online
		"CI Generic Service":    2, // Online
		"CI Offline IP Address": 3, // Offline
		"CI Failed Service":     4, // Failed
	} {
		metric := testutils.RequireFixtureMetric(t, metrics, mscluster.Name, "windows_mscluster_resource_state", prometheus.Labels{"name": name})
		if metric != nil && metric.GetGauge().GetValue() != state {
			t.Errorf("resource %q state = %v, want %v", name, metric.GetGauge().GetValue(), state)
		}
	}
}
