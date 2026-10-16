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

	"github.com/prometheus-community/windows_exporter/internal/labelcache"
	"github.com/prometheus-community/windows_exporter/internal/osversion"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorDynamicMemoryVM Hyper-V Dynamic Memory VM metrics
type collectorDynamicMemoryVM struct {
	perfDataCollectorDynamicMemoryVM *pdh.Collector[perfDataCounterValuesDynamicMemoryVM]
	perfDataObjectDynamicMemoryVM    []perfDataCounterValuesDynamicMemoryVM

	vmMemoryAddedMemory                *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Added Memory
	vmMemoryCurrentPressure            *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Current Pressure
	vmMemoryGuestVisiblePhysicalMemory *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Guest Visible Physical Memory
	vmMemoryMaximumPressure            *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Maximum Pressure
	vmMemoryMemoryAddOperations        *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Memory Add Operations
	vmMemoryMemoryRemoveOperations     *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Memory Remove Operations
	vmMemoryMinimumPressure            *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Minimum Pressure
	vmMemoryPhysicalMemory             *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Physical Memory
	vmMemoryRemovedMemory              *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Removed Memory
	vmMemoryGuestAvailableMemory       *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Guest Available Memory
}

type perfDataCounterValuesDynamicMemoryVM struct {
	Name       string
	Occurrence int

	// Hyper-V Dynamic Memory VM metrics
	VmMemoryAddedMemory                float64 `perfdata:"Added Memory"`
	VmMemoryCurrentPressure            float64 `perfdata:"Current Pressure"`
	VmMemoryGuestAvailableMemory       float64 `perfdata:"Guest Available Memory"        perfdata_min_build:"20348"`
	VmMemoryGuestVisiblePhysicalMemory float64 `perfdata:"Guest Visible Physical Memory"`
	VmMemoryMaximumPressure            float64 `perfdata:"Maximum Pressure"`
	VmMemoryMemoryAddOperations        float64 `perfdata:"Memory Add Operations"`
	VmMemoryMemoryRemoveOperations     float64 `perfdata:"Memory Remove Operations"`
	VmMemoryMinimumPressure            float64 `perfdata:"Minimum Pressure"`
	VmMemoryPhysicalMemory             float64 `perfdata:"Physical Memory"`
	VmMemoryRemovedMemory              float64 `perfdata:"Removed Memory"`
}

func (c *Collector) buildDynamicMemoryVM() error {
	var err error

	c.perfDataCollectorDynamicMemoryVM, err = pdh.NewCollector[perfDataCounterValuesDynamicMemoryVM](c.logger, pdh.CounterTypeRaw, "Hyper-V Dynamic Memory VM", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Dynamic Memory VM collector: %w", err)
	}

	c.vmMemoryAddedMemory = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_added_total"),
		"Represents the cumulative amount of memory added to the VM.",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryCurrentPressure = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_pressure_current_ratio"),
		"Represents the current pressure in the VM.",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryGuestAvailableMemory = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_guest_available_bytes"),
		"Represents the current amount of available memory in the VM (reported by the VM).",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryGuestVisiblePhysicalMemory = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_guest_visible_physical_memory_bytes"),
		"Represents the amount of memory visible in the VM.",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryMaximumPressure = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_pressure_maximum_ratio"),
		"Represents the maximum pressure band in the VM.",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryMemoryAddOperations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_add_operations_total"),
		"Represents the total number of add operations for the VM.",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryMemoryRemoveOperations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_remove_operations_total"),
		"Represents the total number of remove operations for the VM.",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryMinimumPressure = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_pressure_minimum_ratio"),
		"Represents the minimum pressure band in the VM.",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryPhysicalMemory = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_physical_bytes"),
		"Represents the current amount of memory in the VM.",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryRemovedMemory = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_removed_bytes_total"),
		"Represents the cumulative amount of memory removed from the VM.",
		[]string{"vm"},
		nil,
	)

	return nil
}

func (c *Collector) collectDynamicMemoryVM(ch chan<- prometheus.Metric, scrape *labelcache.Scrape) error {
	err := c.perfDataCollectorDynamicMemoryVM.Collect(&c.perfDataObjectDynamicMemoryVM)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Dynamic Memory VM metrics: %w", err)
	}

	nameDynamicMemoryVMs(c.perfDataObjectDynamicMemoryVM)

	for _, data := range c.perfDataObjectDynamicMemoryVM {
		vmName := data.Name

		ch <- scrape.MustNewConstMetric(
			c.vmMemoryAddedMemory,
			prometheus.CounterValue,
			utils.MBToBytes(data.VmMemoryAddedMemory),
			vmName,
		)

		ch <- scrape.MustNewConstMetric(
			c.vmMemoryCurrentPressure,
			prometheus.GaugeValue,
			utils.PercentageToRatio(data.VmMemoryCurrentPressure),
			vmName,
		)

		if osversion.Build() >= osversion.LTSC2022 {
			ch <- scrape.MustNewConstMetric(
				c.vmMemoryGuestAvailableMemory,
				prometheus.GaugeValue,
				utils.MBToBytes(data.VmMemoryGuestAvailableMemory),
				vmName,
			)
		}

		ch <- scrape.MustNewConstMetric(
			c.vmMemoryGuestVisiblePhysicalMemory,
			prometheus.GaugeValue,
			utils.MBToBytes(data.VmMemoryGuestVisiblePhysicalMemory),
			vmName,
		)

		ch <- scrape.MustNewConstMetric(
			c.vmMemoryMaximumPressure,
			prometheus.GaugeValue,
			utils.PercentageToRatio(data.VmMemoryMaximumPressure),
			vmName,
		)

		ch <- scrape.MustNewConstMetric(
			c.vmMemoryMemoryAddOperations,
			prometheus.CounterValue,
			data.VmMemoryMemoryAddOperations,
			vmName,
		)

		ch <- scrape.MustNewConstMetric(
			c.vmMemoryMemoryRemoveOperations,
			prometheus.CounterValue,
			data.VmMemoryMemoryRemoveOperations,
			vmName,
		)

		ch <- scrape.MustNewConstMetric(
			c.vmMemoryMinimumPressure,
			prometheus.GaugeValue,
			utils.PercentageToRatio(data.VmMemoryMinimumPressure),
			vmName,
		)

		ch <- scrape.MustNewConstMetric(
			c.vmMemoryPhysicalMemory,
			prometheus.GaugeValue,
			utils.MBToBytes(data.VmMemoryPhysicalMemory),
			vmName,
		)

		ch <- scrape.MustNewConstMetric(
			c.vmMemoryRemovedMemory,
			prometheus.CounterValue,
			utils.MBToBytes(data.VmMemoryRemovedMemory),
			vmName,
		)
	}

	return nil
}

// vmUnknown is the vm label value for Hyper-V Dynamic Memory VM instances
// without a name.
const vmUnknown = "(unknown)"

// nameDynamicMemoryVMs replaces the instance names of rows with their vm label
// values. Windows reports an empty instance name for VMs managed by the Host
// Compute Service, e.g. WSL 2, which pdh exposes as pdh.InstanceEmpty. Those
// VMs are named vmUnknown before suffixDuplicateVMs keeps several of them, or
// VMs with the same name, apart.
func nameDynamicMemoryVMs(rows []perfDataCounterValuesDynamicMemoryVM) {
	// A VM can be named "(unknown)" itself. The unnamed VMs then count as
	// further occurrences of that name, so that they get a suffix.
	unknownVMs := 0

	for _, row := range rows {
		if row.Name == vmUnknown {
			unknownVMs = max(unknownVMs, row.Occurrence+1)
		}
	}

	for i := range rows {
		if rows[i].Name == pdh.InstanceEmpty {
			rows[i].Name = vmUnknown
			rows[i].Occurrence += unknownVMs
		}
	}

	suffixDuplicateVMs(rows,
		func(row *perfDataCounterValuesDynamicMemoryVM) *string { return &row.Name },
		func(row *perfDataCounterValuesDynamicMemoryVM) int { return row.Occurrence },
	)
}

// suffixDuplicateVMs keeps VMs with the same name apart. Hyper-V allows several
// VMs with the same name, and their counter instances then have the same name
// too. The first keeps the name, later ones get the suffix "#1", "#2" and so on.
// The suffix depends on the order of the sample and does not identify a VM
// across scrapes or across counter objects.
func suffixDuplicateVMs[T any](rows []T, name func(*T) *string, occurrence func(*T) int) {
	pdh.SuffixDuplicates(rows, name, occurrence)
}
