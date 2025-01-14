// Copyright 2024 The Prometheus Authors
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

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorHypervisorRootPartition Hyper-V Hypervisor Root Partition metrics
type collectorHypervisorRootPartition struct {
	perfDataCollectorHypervisorRootPartition *pdh.Collector
	perfDataObjectHypervisorRootPartition    []perfDataCounterValuesHypervisorRootPartition

	hypervisorRootPartitionAddressSpaces                 *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Address Spaces
	hypervisorRootPartitionAttachedDevices               *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Attached Devices
	hypervisorRootPartitionDepositedPages                *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Deposited Pages
	hypervisorRootPartitionDeviceDMAErrors               *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Device DMA Errors
	hypervisorRootPartitionDeviceInterruptErrors         *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Device Interrupt Errors
	hypervisorRootPartitionDeviceInterruptMappings       *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Device Interrupt Mappings
	hypervisorRootPartitionDeviceInterruptThrottleEvents *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Device Interrupt Throttle Events
	hypervisorRootPartitionGPAPages                      *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\GPA Pages
	hypervisorRootPartitionGPASpaceModifications         *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\GPA Space Modifications/sec
	hypervisorRootPartitionIOTLBFlushCost                *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\I/O TLB Flush Cost
	hypervisorRootPartitionIOTLBFlushes                  *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\I/O TLB Flushes/sec
	hypervisorRootPartitionRecommendedVirtualTLBSize     *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Recommended Virtual TLB Size
	hypervisorRootPartitionSkippedTimerTicks             *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Skipped Timer Ticks
	hypervisorRootPartition1GDevicePages                 *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\1G device pages
	hypervisorRootPartition1GGPAPages                    *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\1G GPA pages
	hypervisorRootPartition2MDevicePages                 *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\2M device pages
	hypervisorRootPartition2MGPAPages                    *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\2M GPA pages
	hypervisorRootPartition4KDevicePages                 *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\4K device pages
	hypervisorRootPartition4KGPAPages                    *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\4K GPA pages
	hypervisorRootPartitionVirtualTLBFlushEntries        *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Virtual TLB Flush Entries/sec
	hypervisorRootPartitionVirtualTLBPages               *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Virtual TLB Pages
}

type perfDataCounterValuesHypervisorRootPartition struct {
	HypervisorRootPartitionAddressSpaces                 float64 `perfdata:"Address Spaces"`
	HypervisorRootPartitionAttachedDevices               float64 `perfdata:"Attached Devices"`
	HypervisorRootPartitionDepositedPages                float64 `perfdata:"Deposited Pages"`
	HypervisorRootPartitionDeviceDMAErrors               float64 `perfdata:"Device DMA Errors"`
	HypervisorRootPartitionDeviceInterruptErrors         float64 `perfdata:"Device Interrupt Errors"`
	HypervisorRootPartitionDeviceInterruptMappings       float64 `perfdata:"Device Interrupt Mappings"`
	HypervisorRootPartitionDeviceInterruptThrottleEvents float64 `perfdata:"Device Interrupt Throttle Events"`
	HypervisorRootPartitionGPAPages                      float64 `perfdata:"GPA Pages"`
	HypervisorRootPartitionGPASpaceModifications         float64 `perfdata:"GPA Space Modifications/sec"`
	HypervisorRootPartitionIOTLBFlushCost                float64 `perfdata:"I/O TLB Flush Cost"`
	HypervisorRootPartitionIOTLBFlushes                  float64 `perfdata:"I/O TLB Flushes/sec"`
	HypervisorRootPartitionRecommendedVirtualTLBSize     float64 `perfdata:"Recommended Virtual TLB Size"`
	HypervisorRootPartitionSkippedTimerTicks             float64 `perfdata:"Skipped Timer Ticks"`
	HypervisorRootPartition1GDevicePages                 float64 `perfdata:"1G device pages"`
	HypervisorRootPartition1GGPAPages                    float64 `perfdata:"1G GPA pages"`
	HypervisorRootPartition2MDevicePages                 float64 `perfdata:"2M device pages"`
	HypervisorRootPartition2MGPAPages                    float64 `perfdata:"2M GPA pages"`
	HypervisorRootPartition4KDevicePages                 float64 `perfdata:"4K device pages"`
	HypervisorRootPartition4KGPAPages                    float64 `perfdata:"4K GPA pages"`
	HypervisorRootPartitionVirtualTLBFlushEntries        float64 `perfdata:"Virtual TLB Flush Entires/sec"`
	HypervisorRootPartitionVirtualTLBPages               float64 `perfdata:"Virtual TLB Pages"`
}

func (c *Collector) buildHypervisorRootPartition() error {
	var err error

	c.perfDataCollectorHypervisorRootPartition, err = pdh.NewCollector[perfDataCounterValuesHypervisorRootPartition](pdh.CounterTypeRaw, "Hyper-V Hypervisor Root Partition", []string{"Root"})
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Hypervisor Root Partition collector: %w", err)
	}

	c.hypervisorRootPartitionAddressSpaces = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_address_spaces"),
		"The number of address spaces in the virtual TLB of the partition",
		nil,
		nil,
	)
	c.hypervisorRootPartitionAttachedDevices = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_attached_devices"),
		"The number of devices attached to the partition",
		nil,
		nil,
	)
	c.hypervisorRootPartitionDepositedPages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_deposited_pages"),
		"The number of pages deposited into the partition",
		nil,
		nil,
	)
	c.hypervisorRootPartitionDeviceDMAErrors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_device_dma_errors"),
		"An indicator of illegal DMA requests generated by all devices assigned to the partition",
		nil,
		nil,
	)
	c.hypervisorRootPartitionDeviceInterruptErrors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_device_interrupt_errors"),
		"An indicator of illegal interrupt requests generated by all devices assigned to the partition",
		nil,
		nil,
	)
	c.hypervisorRootPartitionDeviceInterruptMappings = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_device_interrupt_mappings"),
		"The number of device interrupt mappings used by the partition",
		nil,
		nil,
	)
	c.hypervisorRootPartitionDeviceInterruptThrottleEvents = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_device_interrupt_throttle_events"),
		"The number of times an interrupt from a device assigned to the partition was temporarily throttled because the device was generating too many interrupts",
		nil,
		nil,
	)
	c.hypervisorRootPartitionGPAPages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_preferred_numa_node_index"),
		"The number of pages present in the GPA space of the partition (zero for root partition)",
		nil,
		nil,
	)
	c.hypervisorRootPartitionGPASpaceModifications = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_gpa_space_modifications"),
		"The rate of modifications to the GPA space of the partition",
		nil,
		nil,
	)
	c.hypervisorRootPartitionIOTLBFlushCost = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_io_tlb_flush_cost"),
		"The average time (in nanoseconds) spent processing an I/O TLB flush",
		nil,
		nil,
	)
	c.hypervisorRootPartitionIOTLBFlushes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_io_tlb_flush"),
		"The rate of flushes of I/O TLBs of the partition",
		nil,
		nil,
	)
	c.hypervisorRootPartitionRecommendedVirtualTLBSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_recommended_virtual_tlb_size"),
		"The recommended number of pages to be deposited for the virtual TLB",
		nil,
		nil,
	)
	c.hypervisorRootPartitionSkippedTimerTicks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_physical_pages_allocated"),
		"The number of timer interrupts skipped for the partition",
		nil,
		nil,
	)
	c.hypervisorRootPartition1GDevicePages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_1G_device_pages"),
		"The number of 1G pages present in the device space of the partition",
		nil,
		nil,
	)
	c.hypervisorRootPartition1GGPAPages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_1G_gpa_pages"),
		"The number of 1G pages present in the GPA space of the partition",
		nil,
		nil,
	)
	c.hypervisorRootPartition2MDevicePages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_2M_device_pages"),
		"The number of 2M pages present in the device space of the partition",
		nil,
		nil,
	)
	c.hypervisorRootPartition2MGPAPages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_2M_gpa_pages"),
		"The number of 2M pages present in the GPA space of the partition",
		nil,
		nil,
	)
	c.hypervisorRootPartition4KDevicePages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_4K_device_pages"),
		"The number of 4K pages present in the device space of the partition",
		nil,
		nil,
	)
	c.hypervisorRootPartition4KGPAPages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_4K_gpa_pages"),
		"The number of 4K pages present in the GPA space of the partition",
		nil,
		nil,
	)
	c.hypervisorRootPartitionVirtualTLBFlushEntries = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_virtual_tlb_flush_entries"),
		"The rate of flushes of the entire virtual TLB",
		nil,
		nil,
	)
	c.hypervisorRootPartitionVirtualTLBPages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_virtual_tlb_pages"),
		"The number of pages used by the virtual TLB of the partition",
		nil,
		nil,
	)

	return nil
}

func (c *Collector) collectHypervisorRootPartition(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorHypervisorRootPartition.Collect(&c.perfDataObjectHypervisorRootPartition)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Hypervisor Root Partition metrics: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionAddressSpaces,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartitionAddressSpaces,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionAttachedDevices,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartitionAttachedDevices,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionDepositedPages,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartitionDepositedPages,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionDeviceDMAErrors,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartitionDeviceDMAErrors,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionDeviceInterruptErrors,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartitionDeviceInterruptErrors,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionDeviceInterruptThrottleEvents,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartitionDeviceInterruptThrottleEvents,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionGPAPages,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartitionGPAPages,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionGPASpaceModifications,
		prometheus.CounterValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartitionGPASpaceModifications,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionIOTLBFlushCost,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartitionIOTLBFlushCost,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionIOTLBFlushes,
		prometheus.CounterValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartitionIOTLBFlushes,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionRecommendedVirtualTLBSize,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartitionRecommendedVirtualTLBSize,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionSkippedTimerTicks,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartitionSkippedTimerTicks,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartition1GDevicePages,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartition1GDevicePages,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartition1GGPAPages,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartition1GGPAPages,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartition2MDevicePages,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartition2MDevicePages,
	)
	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartition2MGPAPages,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartition2MGPAPages,
	)
	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartition4KDevicePages,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartition4KDevicePages,
	)
	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartition4KGPAPages,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartition4KGPAPages,
	)
	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionVirtualTLBFlushEntries,
		prometheus.CounterValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartitionVirtualTLBFlushEntries,
	)
	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionVirtualTLBPages,
		prometheus.GaugeValue,
		c.perfDataObjectHypervisorRootPartition[0].HypervisorRootPartitionVirtualTLBPages,
	)

	return nil
}
