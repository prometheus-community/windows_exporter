// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

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
	require.NoError(t, c.collectHypervisorRootPartition(metrics, nil))
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
