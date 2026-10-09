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

package hyperv

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestRootPartitionInterruptMappings(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.buildHypervisorRootPartitionDescriptors()
	c.perfDataCollectorHypervisorRootPartition = interruptMappingFixture{}
	metrics := make(chan prometheus.Metric, 100)
	require.NoError(t, c.collectHypervisorRootPartition(metrics))
	close(metrics)

	found := false

	for metric := range metrics {
		if metric.Desc() != c.hypervisorRootPartitionDeviceInterruptMappings {
			continue
		}

		var value dto.Metric
		require.NoError(t, metric.Write(&value))
		require.InDelta(t, 73, value.GetGauge().GetValue(), 1e-9)

		found = true
	}

	require.True(t, found)
}

type interruptMappingFixture struct{}

func (interruptMappingFixture) Collect(dst *[]perfDataCounterValuesHypervisorRootPartition) error {
	*dst = []perfDataCounterValuesHypervisorRootPartition{{HypervisorRootPartitionDeviceInterruptMappings: 73}}

	return nil
}
func (interruptMappingFixture) Close() {}
