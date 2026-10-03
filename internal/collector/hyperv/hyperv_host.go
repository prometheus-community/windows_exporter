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
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/prometheus-community/windows_exporter/internal/headers/hcs"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

// collectorHost Hyper-V Host metrics
type collectorHost struct {
	perfDataCollectorLogicalProcessor *pdh.Collector
	perfDataObjectLogicalProcessor    []perfDataCounterValuesHost

	hostCPURatio              *prometheus.Desc
	vmProcessorCount          *prometheus.Desc
	hostLogicalProcessorCount *prometheus.Desc
	totalVMProcessorCount     *prometheus.Desc
	memoryPropertyQuery       *uint16
}

type perfDataCounterValuesHost struct {
	Name string

	HypervisorLogicalProcessorTotalRunTimePercent float64 `perfdata:"% Total Run Time"`
}

type vmTopology struct {
	ID             string
	Name           string
	State          string
	ProcessorCount int
}

// vmMemoryProperties is the "Memory" property of a compute system. It is the only
// HCS property type that exposes a virtual machine's processor allocation: every
// other documented type is rejected for VMMS-owned virtual machines.
type vmMemoryProperties struct {
	VirtualNodeCount int             `json:"VirtualNodeCount,omitempty"`
	VirtualNodes     []vmVirtualNode `json:"VirtualNodes,omitempty"`
}

type vmVirtualNode struct {
	VirtualNodeIndex      int `json:"VirtualNodeIndex"`
	PhysicalNodeNumber    int `json:"PhysicalNodeNumber"`
	VirtualProcessorCount int `json:"VirtualProcessorCount"`
}

func (c *Collector) buildHost() error {
	var err error

	c.perfDataCollectorLogicalProcessor, err = pdh.NewCollector[perfDataCounterValuesHost](c.logger, pdh.CounterTypeRaw, "Hyper-V Hypervisor Logical Processor", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Hypervisor Logical Processor collector: %w", err)
	}

	c.memoryPropertyQuery, err = windows.UTF16PtrFromString(`{"PropertyTypes":["Memory"]}`)
	if err != nil {
		return fmt.Errorf("failed to create memory property query: %w", err)
	}

	c.hostCPURatio = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "host_cpu_ratio"),
		"Virtual cores assigned to all VMs per logical host core. On a hyper-threaded host, multiply by the number of threads per core to get the ratio per physical core.",
		nil,
		nil,
	)

	c.vmProcessorCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vm_processor_count"),
		"Number of virtual processors assigned to the VM",
		[]string{"vm_id", "vm"},
		nil,
	)

	c.hostLogicalProcessorCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "host_logical_processor_count"),
		"Number of logical processors on the host",
		nil,
		nil,
	)

	c.totalVMProcessorCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_vm_processor_count"),
		"Total number of virtual processors assigned to all VMs",
		nil,
		nil,
	)

	return nil
}

func (c *Collector) collectHost(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorLogicalProcessor.Collect(&c.perfDataObjectLogicalProcessor)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Hypervisor Logical Processor metrics: %w", err)
	}

	logicalCoreCount := float64(len(c.perfDataObjectLogicalProcessor))

	vms, err := c.getVMsWithProcessorCounts()
	if err != nil {
		return fmt.Errorf("failed to get VM processor counts: %w", err)
	}

	var totalVirtualCoreCount float64

	for _, vm := range vms {
		if vm.ProcessorCount > 0 {
			processorCount := float64(vm.ProcessorCount)
			totalVirtualCoreCount += processorCount

			ch <- prometheus.MustNewConstMetric(
				c.vmProcessorCount,
				prometheus.GaugeValue,
				processorCount,
				vm.ID, vm.Name,
			)
		}
	}

	ch <- prometheus.MustNewConstMetric(
		c.hostLogicalProcessorCount,
		prometheus.GaugeValue,
		logicalCoreCount,
	)

	ch <- prometheus.MustNewConstMetric(
		c.totalVMProcessorCount,
		prometheus.GaugeValue,
		totalVirtualCoreCount,
	)

	var ratio float64
	if logicalCoreCount > 0 {
		ratio = totalVirtualCoreCount / logicalCoreCount
	}

	ch <- prometheus.MustNewConstMetric(
		c.hostCPURatio,
		prometheus.GaugeValue,
		ratio,
	)

	return nil
}

func (c *Collector) getVMsWithProcessorCounts() ([]vmTopology, error) {
	vmQuery, err := windows.UTF16PtrFromString(`{"Types":["VirtualMachine"]}`)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM query: %w", err)
	}

	operation, err := hcs.CreateOperation()
	if err != nil {
		return nil, fmt.Errorf("failed to create operation: %w", err)
	}
	defer hcs.CloseOperation(operation)

	if err := hcs.EnumerateComputeSystems(vmQuery, operation); err != nil {
		return nil, fmt.Errorf("failed to enumerate compute systems: %w", err)
	}

	resultDocument, err := hcs.WaitForOperationResult(operation, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for operation result: %w - %s", err, resultDocument)
	} else if resultDocument == "" {
		return nil, hcs.ErrEmptyResultDocument
	}

	var computeSystems []hcs.Properties
	if err := json.Unmarshal([]byte(resultDocument), &computeSystems); err != nil {
		return nil, fmt.Errorf("failed to unmarshal compute systems: %w", err)
	}

	vms := make([]vmTopology, 0, len(computeSystems))
	errs := make([]error, 0)

	for _, system := range computeSystems {
		// HcsEnumerateComputeSystems does not populate State for VMMS-owned virtual
		// machines, so only skip a system when it reports a state that is not running.
		if system.State != "" && system.State != "Running" {
			continue
		}

		processorCount, err := c.getVMProcessorCount(system.ID)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get processor count for VM %s: %w", system.ID, err))

			continue
		}

		vms = append(vms, vmTopology{
			ID:             system.ID,
			Name:           system.Name,
			State:          system.State,
			ProcessorCount: processorCount,
		})
	}

	if len(errs) > 0 {
		c.logger.Debug("some VMs failed processor count retrieval",
			slog.Any("err", errors.Join(errs...)),
		)
	}

	return vms, nil
}

func (c *Collector) getVMProcessorCount(vmID string) (int, error) {
	computeSystem, err := hcs.OpenComputeSystem(vmID)
	if err != nil {
		return 0, fmt.Errorf("failed to open compute system: %w", err)
	}

	defer hcs.CloseComputeSystem(computeSystem)

	operation, err := hcs.CreateOperation()
	if err != nil {
		return 0, fmt.Errorf("failed to create operation: %w", err)
	}

	defer hcs.CloseOperation(operation)

	if err := hcs.GetComputeSystemProperties(computeSystem, operation, c.memoryPropertyQuery); err != nil {
		return 0, fmt.Errorf("failed to get compute system properties: %w", err)
	}

	resultDocument, err := hcs.WaitForOperationResult(operation, 1000)
	if err != nil {
		return 0, fmt.Errorf("failed to wait for operation result: %w", err)
	} else if resultDocument == "" {
		return 0, hcs.ErrEmptyResultDocument
	}

	var properties struct {
		Memory *vmMemoryProperties `json:"Memory,omitempty"`
	}

	if err := json.Unmarshal([]byte(resultDocument), &properties); err != nil {
		return 0, fmt.Errorf("failed to unmarshal properties: %w", err)
	}

	if properties.Memory == nil {
		return 0, nil
	}

	// A virtual machine's processors are reported per virtual NUMA node.
	processorCount := 0
	for _, node := range properties.Memory.VirtualNodes {
		processorCount += node.VirtualProcessorCount
	}

	return processorCount, nil
}
