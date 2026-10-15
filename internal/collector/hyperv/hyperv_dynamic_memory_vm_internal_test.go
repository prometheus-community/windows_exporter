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

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestNameDynamicMemoryVMs(t *testing.T) {
	t.Parallel()

	rows := []perfDataCounterValuesDynamicMemoryVM{
		{Name: pdh.InstanceEmpty},
		{Name: pdh.InstanceEmpty, Occurrence: 1},
		{Name: pdh.InstanceEmpty, Occurrence: 2},
		{Name: "GitHubActions"},
		{Name: "GitHubActions", Occurrence: 1},
		// A VM named like a suffixed duplicate keeps its name.
		{Name: "dup#1"},
		{Name: "dup"},
		{Name: "dup", Occurrence: 1},
		{Name: "vm-------"},
	}

	nameDynamicMemoryVMs(rows)

	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.Name)
	}

	require.Equal(t, []string{
		"(unknown)", "(unknown)#1", "(unknown)#2",
		"GitHubActions", "GitHubActions#1",
		"dup#1", "dup", "dup#2",
		"vm-------",
	}, names)
}

// TestVirtualProcessorDuplicateVMs checks that the virtual processors of VMs
// with the same name are published with a suffixed vm label and an unchanged
// core label.
func TestVirtualProcessorDuplicateVMs(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.buildHypervisorVirtualProcessorDescriptors()

	c.perfDataObjectHypervisorVirtualProcessor = []perfDataCounterValuesHypervisorVirtualProcessor{
		{Name: "dup:Hv VP 0", HypervisorVirtualProcessorDispatches: 1},
		{Name: "dup:Hv VP 1", HypervisorVirtualProcessorDispatches: 2},
		{Name: "dup:Hv VP 0", Occurrence: 1, HypervisorVirtualProcessorDispatches: 3},
		{Name: "other:Hv VP 0", HypervisorVirtualProcessorDispatches: 4},
	}

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(publishCollector(func(ch chan<- prometheus.Metric) error {
		return c.publishHypervisorVirtualProcessor(ch, nil)
	})))

	families, err := registry.Gather()
	require.NoError(t, err)

	got := map[string]float64{}

	for _, family := range families {
		if family.GetName() != "windows_hyperv_hypervisor_virtual_processor_dispatches_total" {
			continue
		}

		for _, metric := range family.GetMetric() {
			labels := map[string]string{}
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}

			got[labels["vm"]+"/"+labels["core"]] = metric.GetCounter().GetValue()
		}
	}

	require.Equal(t, map[string]float64{"dup/0": 1, "dup/1": 2, "dup#1/0": 3, "other/0": 4}, got)
}
