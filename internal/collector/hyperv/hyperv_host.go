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
	"os"

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
	computeTopologyQuery      *uint16
}

type perfDataCounterValuesHost struct {
	Name string
}

type vmTopology struct {
	ID       string
	Name     string
	State    string
	Topology *computeTopology
}

type computeTopology struct {
	Processor *processorTopology `json:"Processor,omitempty"`
}

type processorTopology struct {
	Count int `json:"Count,omitempty"`
}

func (c *Collector) buildHost() error {
	var err error

	c.perfDataCollectorLogicalProcessor, err = pdh.NewCollector[perfDataCounterValuesHost](c.logger, pdh.CounterTypeRaw, "Hyper-V Hypervisor Logical Processor", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Hypervisor Logical Processor collector: %w", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	c.computeTopologyQuery, err = windows.UTF16PtrFromString(`{"PropertyTypes":["ComputeTopology"]}`)
	if err != nil {
		return fmt.Errorf("failed to create compute topology query: %w", err)
	}

	c.hostCPURatio = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "host_cpu_ratio"),
		"Ratio of physical logical CPU cores to virtual cores assigned to all VMs",
		[]string{"host"},
		prometheus.Labels{"host": hostname},
	)

	c.vmProcessorCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vm_processor_count"),
		"Number of virtual processors assigned to the VM",
		[]string{"vm_id", "vm_name"},
		nil,
	)

	c.hostLogicalProcessorCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "host_logical_processor_count"),
		"Number of logical processors on the host",
		[]string{"host"},
		prometheus.Labels{"host": hostname},
	)

	c.totalVMProcessorCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "total_vm_processor_count"),
		"Total number of virtual processors assigned to all VMs",
		[]string{"host"},
		prometheus.Labels{"host": hostname},
	)

	return nil
}

func (c *Collector) collectHost(ch chan<- prometheus.Metric) error {
	// Collect logical processor count
	err := c.perfDataCollectorLogicalProcessor.Collect(&c.perfDataObjectLogicalProcessor)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Hypervisor Logical Processor metrics: %w", err)
	}

	logicalCoreCount := float64(len(c.perfDataObjectLogicalProcessor))

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Get VMs and their processor counts via HCS
	vms, err := c.getVMsWithTopology()
	if err != nil {
		return fmt.Errorf("failed to get VMs with topology: %w", err)
	}

	var totalVirtualCoreCount float64

	for _, vm := range vms {
		if vm.Topology != nil && vm.Topology.Processor != nil {
			processorCount := float64(vm.Topology.Processor.Count)
			totalVirtualCoreCount += processorCount

			ch <- prometheus.MustNewConstMetric(
				c.vmProcessorCount,
				prometheus.GaugeValue,
				processorCount,
				vm.ID, vm.Name,
			)
		}
	}

	// Send host logical processor count
	ch <- prometheus.MustNewConstMetric(
		c.hostLogicalProcessorCount,
		prometheus.GaugeValue,
		logicalCoreCount,
		hostname,
	)

	// Send total VM processor count
	ch <- prometheus.MustNewConstMetric(
		c.totalVMProcessorCount,
		prometheus.GaugeValue,
		totalVirtualCoreCount,
		hostname,
	)

	// Calculate and send CPU ratio
	var ratio float64
	if totalVirtualCoreCount > 0 {
		ratio = logicalCoreCount / totalVirtualCoreCount
	}

	ch <- prometheus.MustNewConstMetric(
		c.hostCPURatio,
		prometheus.GaugeValue,
		ratio,
		hostname,
	)

	return nil
}

func (c *Collector) getVMsWithTopology() ([]vmTopology, error) {
	// Query for VirtualMachine type instead of Container
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
		// Skip non-running VMs
		if system.State != "Running" {
			continue
		}

		topology, err := c.getVMTopology(system.ID)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get topology for VM %s: %w", system.ID, err))
			continue
		}

		vms = append(vms, vmTopology{
			ID:       system.ID,
			Name:     system.ID, // HCS doesn't provide a name directly, use ID
			State:    system.State,
			Topology: topology,
		})
	}

	if len(errs) > 0 {
		c.logger.Debug("some VMs failed topology retrieval", "errors", errors.Join(errs...))
	}

	return vms, nil
}

func (c *Collector) getVMTopology(vmID string) (*computeTopology, error) {
	computeSystem, err := hcs.OpenComputeSystem(vmID)
	if err != nil {
		return nil, fmt.Errorf("failed to open compute system: %w", err)
	}
	defer hcs.CloseComputeSystem(computeSystem)

	operation, err := hcs.CreateOperation()
	if err != nil {
		return nil, fmt.Errorf("failed to create operation: %w", err)
	}
	defer hcs.CloseOperation(operation)

	if err := hcs.GetComputeSystemProperties(computeSystem, operation, c.computeTopologyQuery); err != nil {
		return nil, fmt.Errorf("failed to get compute system properties: %w", err)
	}

	resultDocument, err := hcs.WaitForOperationResult(operation, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for operation result: %w", err)
	} else if resultDocument == "" {
		return nil, hcs.ErrEmptyResultDocument
	}

	var properties struct {
		ComputeTopology *computeTopology `json:"ComputeTopology,omitempty"`
	}

	if err := json.Unmarshal([]byte(resultDocument), &properties); err != nil {
		return nil, fmt.Errorf("failed to unmarshal properties: %w", err)
	}

	return properties.ComputeTopology, nil
}
