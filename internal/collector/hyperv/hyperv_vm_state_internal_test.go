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

func TestVMStateName(t *testing.T) {
	t.Parallel()

	for enabledState, want := range map[uint16]string{
		0:     "unknown",
		1:     "unknown",
		2:     "running",
		3:     "off",
		4:     "stopping",
		5:     "unknown",
		6:     "saved",
		9:     "paused",
		10:    "starting",
		11:    "reset",
		32768: "paused",
		32769: "saved",
		32770: "starting",
		32773: "saving",
		32774: "stopping",
		32776: "pausing",
		32777: "resuming",
		32779: "saved",
		32780: "saving",
		32781: "running",
		32782: "off",
		32783: "stopping",
		32784: "saved",
		32785: "paused",
		32786: "starting",
		32787: "reset",
		32788: "saving",
		32789: "pausing",
		32790: "resuming",
		32791: "saved",
		32792: "saving",
		32793: "unknown",
		65535: "unknown",
	} {
		got := vmStateName(enabledState)
		require.Equalf(t, want, got, "EnabledState %d", enabledState)
		require.Containsf(t, vmStates, got, "EnabledState %d", enabledState)
	}
}

func TestIsVMID(t *testing.T) {
	t.Parallel()

	require.True(t, isVMID("8E3A2C14-5B7D-4F11-9C2E-0A1B2C3D4E5F"))
	require.True(t, isVMID("8e3a2c14-5b7d-4f11-9c2e-0a1b2c3d4e5f"))
	require.False(t, isVMID("HYPERV01"))
	require.False(t, isVMID(""))
	require.False(t, isVMID("{8E3A2C14-5B7D-4F11-9C2E-0A1B2C3D4E5F}"))
	require.False(t, isVMID("8E3A2C14-5B7D-4F11-9C2E-0A1B2C3D4E5G"))
	require.False(t, isVMID("8E3A2C145-B7D-4F11-9C2E-0A1B2C3D4E5F"))
}

// TestPublishVMState checks that every VM, whatever its state, publishes one
// series per state with 1 for the current state, that VMs with the same name
// are kept apart by vm_id and that the host itself is dropped.
func TestPublishVMState(t *testing.T) {
	t.Parallel()

	const (
		runningID = "8E3A2C14-5B7D-4F11-9C2E-0A1B2C3D4E5F"
		offID     = "11111111-2222-3333-4444-555555555555"
		savedID   = "AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE"
	)

	c := New(nil)
	c.buildVMStateDescriptors()

	systems := []msvmComputerSystemState{
		{Name: "HYPERV01", ElementName: "HYPERV01", EnabledState: 2}, // the host itself
		{Name: runningID, ElementName: "web", EnabledState: 2},
		{Name: offID, ElementName: "web", EnabledState: 3},
		{Name: savedID, ElementName: "db", EnabledState: 6},
	}

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(publishCollector(func(ch chan<- prometheus.Metric) error {
		c.publishVMState(ch, nil, systems)

		return nil
	})))

	families, err := registry.Gather()
	require.NoError(t, err)
	require.Len(t, families, 1)

	family := families[0]
	require.Equal(t, "windows_hyperv_vm_state", family.GetName())
	require.Equal(t, dto.MetricType_GAUGE, family.GetType())
	require.Len(t, family.GetMetric(), 3*len(vmStates))

	current := map[string]string{}
	statesPerVM := map[string][]string{}

	for _, metric := range family.GetMetric() {
		labels := map[string]string{}
		for _, label := range metric.GetLabel() {
			labels[label.GetName()] = label.GetValue()
		}

		require.NotEqual(t, "HYPERV01", labels["vm_id"], "the host must not be published")

		key := labels["vm"] + "/" + labels["vm_id"]
		statesPerVM[key] = append(statesPerVM[key], labels["state"])

		switch value := metric.GetGauge().GetValue(); value {
		case 1:
			require.NotContains(t, current, key, "more than one current state")

			current[key] = labels["state"]
		case 0:
		default:
			t.Fatalf("unexpected value %v for %s", value, key)
		}
	}

	require.Equal(t, map[string]string{
		"web/" + runningID: "running",
		"web/" + offID:     "off",
		"db/" + savedID:    "saved",
	}, current)

	for key, states := range statesPerVM {
		require.ElementsMatchf(t, vmStates, states, "states of %s", key)
	}
}
