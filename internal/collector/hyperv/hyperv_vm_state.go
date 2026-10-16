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
	"fmt"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/labelcache"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// vmStateQueryTimeout bounds the virtual machine query, so a hung WMI service
// does not block the scrape.
const vmStateQueryTimeout = 5 * time.Second

// vmStateUnknown is the state of an EnabledState value without a mapping.
const vmStateUnknown = "unknown"

// vmStates are the values of the state label of windows_hyperv_vm_state.
// Every VM publishes one series per state.
//
//nolint:gochecknoglobals
var vmStates = []string{
	"running",
	"off",
	"starting",
	"stopping",
	"saving",
	"saved",
	"pausing",
	"paused",
	"resuming",
	"reset",
	vmStateUnknown,
}

// collectorVMState Hyper-V virtual machine state metrics.
//
// The PDH objects only have instances for running VMs. Msvm_ComputerSystem
// lists every VM registered on the host, whatever its state.
type collectorVMState struct {
	miQueryVMState mi.Query

	vmState *prometheus.Desc // Msvm_ComputerSystem.EnabledState
}

// msvmComputerSystemState is the part of Msvm_ComputerSystem needed for the
// state of a VM.
// https://learn.microsoft.com/en-us/windows/win32/hyperv_v2/msvm-computersystem
type msvmComputerSystemState struct {
	Name         string `mi:"Name"`
	ElementName  string `mi:"ElementName"`
	EnabledState uint16 `mi:"EnabledState"`
}

func (c *Collector) buildVMState() error {
	if c.miSession == nil {
		return mi.ErrNotInitialized
	}

	var err error

	c.miQueryVMState, err = mi.NewQuery("SELECT Name, ElementName, EnabledState FROM Msvm_ComputerSystem")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.buildVMStateDescriptors()

	return nil
}

func (c *Collector) buildVMStateDescriptors() {
	c.vmState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vm_state"),
		"The state of the virtual machine, 1 if the current state, 0 otherwise",
		[]string{"vm", "vm_id", "state"},
		nil,
	)
}

func (c *Collector) collectVMState(ch chan<- prometheus.Metric, scrape *labelcache.Scrape) error {
	var dst []msvmComputerSystemState

	if err := c.miSession.Query(&dst, mi.NamespaceRootVirtualizationV2, c.miQueryVMState, vmStateQueryTimeout); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	c.publishVMState(ch, scrape, dst)

	return nil
}

func (c *Collector) publishVMState(ch chan<- prometheus.Metric, scrape *labelcache.Scrape, systems []msvmComputerSystemState) {
	for _, system := range systems {
		// The host itself is an Msvm_ComputerSystem too. Its Name is the
		// NetBIOS name of the host (at most 15 characters), the Name of a VM
		// is its GUID. Caption tells them apart as well, but it is localized.
		if !isVMID(system.Name) {
			continue
		}

		current := vmStateName(system.EnabledState)

		for _, state := range vmStates {
			value := 0.0
			if state == current {
				value = 1.0
			}

			ch <- scrape.MustNewConstMetric(
				c.vmState,
				prometheus.GaugeValue,
				value,
				system.ElementName,
				system.Name,
				state,
			)
		}
	}
}

// vmStateName maps Msvm_ComputerSystem.EnabledState to the state label.
//
// The EnabledState documentation lists the CIM values 0 to 10. 11 (Reset) is
// a RequestedState value, mapped in case a VM reports it. 32768 (Paused),
// 32769 (Suspended), 32770 (Starting) and 32774 (Stopping) are the
// EnabledState values of the version 1 provider; the version 2 provider still
// reports some of them, for example 32768 for a VM paused on low disk space.
// 32773 (Saving), 32776 (Pausing), 32777 (Resuming), 32779 (FastSaved) and
// 32780 (FastSaving) are the Hyper-V specific states of RequestStateChange,
// and 32781 to 32792 their critical variants, in the order RunningCritical,
// OffCritical, StoppingCritical, SavedCritical, PausedCritical,
// StartingCritical, ResetCritical, SavingCritical, PausingCritical,
// ResumingCritical, FastSavedCritical, FastSavingCritical. A critical state
// is published as its base state; HealthState carries the critical part.
// https://learn.microsoft.com/en-us/windows/win32/hyperv_v2/msvm-computersystem
// https://learn.microsoft.com/en-us/windows/win32/hyperv_v2/requeststatechange-msvm-computersystem
func vmStateName(enabledState uint16) string {
	switch enabledState {
	case 2, 32781: // Enabled, RunningCritical
		return "running"
	case 3, 32782: // Disabled, OffCritical
		return "off"
	case 4, 32774, 32783: // Shutting Down, Stopping, StoppingCritical
		return "stopping"
	case 6, 32769, 32779, 32784, 32791: // Enabled but Offline, Suspended, FastSaved, SavedCritical, FastSavedCritical
		return "saved"
	case 9, 32768, 32785: // Quiesce, Paused, PausedCritical
		return "paused"
	case 10, 32770, 32786: // Starting, Starting (version 1), StartingCritical
		return "starting"
	case 11, 32787: // Reset, ResetCritical
		return "reset"
	case 32773, 32780, 32788, 32792: // Saving, FastSaving, SavingCritical, FastSavingCritical
		return "saving"
	case 32776, 32789: // Pausing, PausingCritical
		return "pausing"
	case 32777, 32790: // Resuming, ResumingCritical
		return "resuming"
	default:
		return vmStateUnknown
	}
}

// isVMID reports whether name has the GUID format
// XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX that Hyper-V uses for VM IDs.
func isVMID(name string) bool {
	if len(name) != 36 {
		return false
	}

	for i := range len(name) {
		switch i {
		case 8, 13, 18, 23:
			if name[i] != '-' {
				return false
			}
		default:
			if !isHexDigit(name[i]) {
				return false
			}
		}
	}

	return true
}

func isHexDigit(b byte) bool {
	return ('0' <= b && b <= '9') || ('a' <= b && b <= 'f') || ('A' <= b && b <= 'F')
}
