package hyperv

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorHypervisorRootPartition Hyper-V Hypervisor Root Partition metrics
type collectorHypervisorRootPartition struct {
	perfDataCollectorHypervisorRootPartition             *perfdata.Collector
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

const (
	hypervisorRootPartitionAddressSpaces                 = "Address Spaces"
	hypervisorRootPartitionAttachedDevices               = "Attached Devices"
	hypervisorRootPartitionDepositedPages                = "Deposited Pages"
	hypervisorRootPartitionDeviceDMAErrors               = "Device DMA Errors"
	hypervisorRootPartitionDeviceInterruptErrors         = "Device Interrupt Errors"
	hypervisorRootPartitionDeviceInterruptMappings       = "Device Interrupt Mappings"
	hypervisorRootPartitionDeviceInterruptThrottleEvents = "Device Interrupt Throttle Events"
	hypervisorRootPartitionGPAPages                      = "GPA Pages"
	hypervisorRootPartitionGPASpaceModifications         = "GPA Space Modifications/sec"
	hypervisorRootPartitionIOTLBFlushCost                = "I/O TLB Flush Cost"
	hypervisorRootPartitionIOTLBFlushes                  = "I/O TLB Flushes/sec"
	hypervisorRootPartitionRecommendedVirtualTLBSize     = "Recommended Virtual TLB Size"
	hypervisorRootPartitionSkippedTimerTicks             = "Skipped Timer Ticks"
	hypervisorRootPartition1GDevicePages                 = "1G device pages"
	hypervisorRootPartition1GGPAPages                    = "1G GPA pages"
	hypervisorRootPartition2MDevicePages                 = "2M device pages"
	hypervisorRootPartition2MGPAPages                    = "2M GPA pages"
	hypervisorRootPartition4KDevicePages                 = "4K device pages"
	hypervisorRootPartition4KGPAPages                    = "4K GPA pages"
	hypervisorRootPartitionVirtualTLBFlushEntries        = "Virtual TLB Flush Entires/sec"
	hypervisorRootPartitionVirtualTLBPages               = "Virtual TLB Pages"
)

func (c *Collector) buildHypervisorRootPartition() error {
	var err error

	c.perfDataCollectorHypervisorRootPartition, err = perfdata.NewCollector("Hyper-V Hypervisor Root Partition", []string{"Root"}, []string{
		hypervisorRootPartitionAddressSpaces,
		hypervisorRootPartitionAttachedDevices,
		hypervisorRootPartitionDepositedPages,
		hypervisorRootPartitionDeviceDMAErrors,
		hypervisorRootPartitionDeviceInterruptErrors,
		hypervisorRootPartitionDeviceInterruptMappings,
		hypervisorRootPartitionDeviceInterruptThrottleEvents,
		hypervisorRootPartitionGPAPages,
		hypervisorRootPartitionGPASpaceModifications,
		hypervisorRootPartitionIOTLBFlushCost,
		hypervisorRootPartitionIOTLBFlushes,
		hypervisorRootPartitionRecommendedVirtualTLBSize,
		hypervisorRootPartitionSkippedTimerTicks,
		hypervisorRootPartition1GDevicePages,
		hypervisorRootPartition1GGPAPages,
		hypervisorRootPartition2MDevicePages,
		hypervisorRootPartition2MGPAPages,
		hypervisorRootPartition4KDevicePages,
		hypervisorRootPartition4KGPAPages,
		hypervisorRootPartitionVirtualTLBFlushEntries,
		hypervisorRootPartitionVirtualTLBPages,
	})
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
	data, err := c.perfDataCollectorHypervisorRootPartition.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Hypervisor Root Partition metrics: %w", err)
	}

	rootData, ok := data["Root"]
	if !ok {
		return errors.New("no data returned from Hyper-V Hypervisor Root Partition")
	}

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionAddressSpaces,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartitionAddressSpaces].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionAttachedDevices,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartitionAttachedDevices].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionDepositedPages,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartitionDepositedPages].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionDeviceDMAErrors,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartitionDeviceDMAErrors].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionDeviceInterruptErrors,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartitionDeviceInterruptErrors].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionDeviceInterruptThrottleEvents,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartitionDeviceInterruptThrottleEvents].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionGPAPages,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartitionGPAPages].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionGPASpaceModifications,
		prometheus.CounterValue,
		rootData[hypervisorRootPartitionGPASpaceModifications].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionIOTLBFlushCost,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartitionIOTLBFlushCost].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionIOTLBFlushes,
		prometheus.CounterValue,
		rootData[hypervisorRootPartitionIOTLBFlushes].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionRecommendedVirtualTLBSize,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartitionRecommendedVirtualTLBSize].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionSkippedTimerTicks,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartitionSkippedTimerTicks].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartition1GDevicePages,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartition1GDevicePages].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartition1GGPAPages,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartition1GGPAPages].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartition2MDevicePages,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartition2MDevicePages].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartition2MGPAPages,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartition2MGPAPages].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartition4KDevicePages,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartition4KDevicePages].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartition4KGPAPages,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartition4KGPAPages].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionVirtualTLBFlushEntries,
		prometheus.CounterValue,
		rootData[hypervisorRootPartitionVirtualTLBFlushEntries].FirstValue,
	)
	ch <- prometheus.MustNewConstMetric(
		c.hypervisorRootPartitionVirtualTLBPages,
		prometheus.GaugeValue,
		rootData[hypervisorRootPartitionVirtualTLBPages].FirstValue,
	)

	return nil
}
