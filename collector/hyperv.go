package collector

import (
	"log"
	"strings"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	Factories["hyperv"] = NewHyperVCollector
}

// HyperVCollector is a Prometheus collector for hyper-v
type HyperVCollector struct {
	// Win32_PerfRawData_VmmsVirtualMachineStats_HyperVVirtualMachineHealthSummary
	HealthCritical *prometheus.Desc
	HealthOk       *prometheus.Desc

	// Win32_PerfRawData_VidPerfProvider_HyperVVMVidPartition
	PhysicalPagesAllocated *prometheus.Desc
	PreferredNUMANodeIndex *prometheus.Desc
	RemotePhysicalPages    *prometheus.Desc

	// Win32_PerfRawData_HvStats_HyperVHypervisorRootPartition
	AddressSpaces                 *prometheus.Desc
	AttachedDevices               *prometheus.Desc
	DepositedPages                *prometheus.Desc
	DeviceDMAErrors               *prometheus.Desc
	DeviceInterruptErrors         *prometheus.Desc
	DeviceInterruptMappings       *prometheus.Desc
	DeviceInterruptThrottleEvents *prometheus.Desc
	GPAPages                      *prometheus.Desc
	GPASpaceModifications         *prometheus.Desc
	IOTLBFlushCost                *prometheus.Desc
	IOTLBFlushes                  *prometheus.Desc
	RecommendedVirtualTLBSize     *prometheus.Desc
	SkippedTimerTicks             *prometheus.Desc
	Value1Gdevicepages            *prometheus.Desc
	Value1GGPApages               *prometheus.Desc
	Value2Mdevicepages            *prometheus.Desc
	Value2MGPApages               *prometheus.Desc
	Value4Kdevicepages            *prometheus.Desc
	Value4KGPApages               *prometheus.Desc
	VirtualTLBFlushEntires        *prometheus.Desc
	VirtualTLBPages               *prometheus.Desc

	// Win32_PerfRawData_HvStats_HyperVHypervisor
	LogicalProcessors *prometheus.Desc
	VirtualProcessors *prometheus.Desc

	// Win32_PerfRawData_HvStats_HyperVHypervisorVirtualProcessor
	PercentGuestRunTime      *prometheus.Desc
	PercentHypervisorRunTime *prometheus.Desc
	PercentRemoteRunTime     *prometheus.Desc
	PercentTotalRunTime      *prometheus.Desc

	// Win32_PerfRawData_NvspSwitchStats_HyperVVirtualSwitch
	BroadcastPacketsReceived         *prometheus.Desc
	BroadcastPacketsSent             *prometheus.Desc
	Bytes                            *prometheus.Desc
	BytesReceived                    *prometheus.Desc
	BytesSent                        *prometheus.Desc
	DirectedPacketsReceived          *prometheus.Desc
	DirectedPacketsSent              *prometheus.Desc
	DroppedPacketsIncoming           *prometheus.Desc
	DroppedPacketsOutgoing           *prometheus.Desc
	ExtensionsDroppedPacketsIncoming *prometheus.Desc
	ExtensionsDroppedPacketsOutgoing *prometheus.Desc
	LearnedMacAddresses              *prometheus.Desc
	MulticastPacketsReceived         *prometheus.Desc
	MulticastPacketsSent             *prometheus.Desc
	NumberofSendChannelMoves         *prometheus.Desc
	NumberofVMQMoves                 *prometheus.Desc
	PacketsFlooded                   *prometheus.Desc
	Packets                          *prometheus.Desc
	PacketsReceived                  *prometheus.Desc
	PacketsSent                      *prometheus.Desc
	PurgedMacAddresses               *prometheus.Desc

	// Win32_PerfRawData_EthernetPerfProvider_HyperVLegacyNetworkAdapter
	AdapterBytesDropped   *prometheus.Desc
	AdapterBytesReceived  *prometheus.Desc
	AdapterBytesSent      *prometheus.Desc
	AdapterFramesDropped  *prometheus.Desc
	AdapterFramesReceived *prometheus.Desc
	AdapterFramesSent     *prometheus.Desc
}

// NewHyperVCollector ...
func NewHyperVCollector() (Collector, error) {
	buildSubsystemName := func(component string) string { return "hyperv_" + component }
	return &HyperVCollector{
		HealthCritical: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("health"), "critical"),
			"This counter represents the number of virtual machines with critical health",
			nil,
			nil,
		),
		HealthOk: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("health"), "ok"),
			"This counter represents the number of virtual machines with ok health",
			nil,
			nil,
		),

		//

		PhysicalPagesAllocated: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vid"), "physical_pages_allocated"),
			"The number of physical pages allocated",
			nil,
			nil,
		),
		PreferredNUMANodeIndex: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vid"), "preferred_numa_node_index"),
			"The preferred NUMA node index associated with this partition",
			nil,
			nil,
		),
		RemotePhysicalPages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vid"), "remote_physical_pages"),
			"The number of physical pages not allocated from the preferred NUMA node",
			nil,
			nil,
		),

		//

		AddressSpaces: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "address_spaces"),
			"The number of address spaces in the virtual TLB of the partition",
			nil,
			nil,
		),
		AttachedDevices: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "attached_devices"),
			"The number of devices attached to the partition",
			nil,
			nil,
		),
		DepositedPages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "deposited_pages"),
			"The number of pages deposited into the partition",
			nil,
			nil,
		),
		DeviceDMAErrors: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "device_dma_errors"),
			"An indicator of illegal DMA requests generated by all devices assigned to the partition",
			nil,
			nil,
		),
		DeviceInterruptErrors: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "device_interrupt_errors"),
			"An indicator of illegal interrupt requests generated by all devices assigned to the partition",
			nil,
			nil,
		),
		DeviceInterruptMappings: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "device_interrupt_mappings"),
			"The number of device interrupt mappings used by the partition",
			nil,
			nil,
		),
		DeviceInterruptThrottleEvents: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "device_interrupt_throttle_events"),
			"The number of times an interrupt from a device assigned to the partition was temporarily throttled because the device was generating too many interrupts",
			nil,
			nil,
		),
		GPAPages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "preferred_numa_node_index"),
			"The number of pages present in the GPA space of the partition (zero for root partition)",
			nil,
			nil,
		),
		GPASpaceModifications: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "gpa_space_modifications"),
			"The rate of modifications to the GPA space of the partition",
			nil,
			nil,
		),
		IOTLBFlushCost: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "io_tlb_flush_cost"),
			"The average time (in nanoseconds) spent processing an I/O TLB flush",
			nil,
			nil,
		),
		IOTLBFlushes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "io_tlb_flush"),
			"The rate of flushes of I/O TLBs of the partition",
			nil,
			nil,
		),
		RecommendedVirtualTLBSize: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "recommended_virtual_tlb_size"),
			"The recommended number of pages to be deposited for the virtual TLB",
			nil,
			nil,
		),
		SkippedTimerTicks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "physical_pages_allocated"),
			"The number of timer interrupts skipped for the partition",
			nil,
			nil,
		),
		Value1Gdevicepages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "1G_device_pages"),
			"The number of 1G pages present in the device space of the partition",
			nil,
			nil,
		),
		Value1GGPApages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "1G_gpa_pages"),
			"The number of 1G pages present in the GPA space of the partition",
			nil,
			nil,
		),
		Value2Mdevicepages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "2M_device_pages"),
			"The number of 2M pages present in the device space of the partition",
			nil,
			nil,
		),
		Value2MGPApages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "2M_gpa_pages"),
			"The number of 2M pages present in the GPA space of the partition",
			nil,
			nil,
		),
		Value4Kdevicepages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "4K_device_pages"),
			"The number of 4K pages present in the device space of the partition",
			nil,
			nil,
		),
		Value4KGPApages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "4K_gpa_pages"),
			"The number of 4K pages present in the GPA space of the partition",
			nil,
			nil,
		),
		VirtualTLBFlushEntires: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "virtual_tlb_flush_entires"),
			"The rate of flushes of the entire virtual TLB",
			nil,
			nil,
		),
		VirtualTLBPages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("root_partition"), "virtual_tlb_pages"),
			"The number of pages used by the virtual TLB of the partition",
			nil,
			nil,
		),

		//

		VirtualProcessors: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("hypervisor"), "virtual_processors"),
			"The number of virtual processors present in the system",
			nil,
			nil,
		),
		LogicalProcessors: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("hypervisor"), "logical_processors"),
			"The number of logical processors present in the system",
			nil,
			nil,
		),

		//

		PercentGuestRunTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vcpu"), "guest_run_time"),
			"The time spent by the virtual processor in guest code",
			[]string{"core"},
			nil,
		),
		PercentHypervisorRunTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vcpu"), "hypervisor_run_time"),
			"The time spent by the virtual processor in hypervisor code",
			[]string{"core"},
			nil,
		),
		PercentRemoteRunTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vcpu"), "remote_run_time"),
			"The time spent by the virtual processor running on a remote node",
			[]string{"core"},
			nil,
		),
		PercentTotalRunTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vcpu"), "total_run_time"),
			"The time spent by the virtual processor in guest and hypervisor code",
			[]string{"core"},
			nil,
		),

		//
		BroadcastPacketsReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "broadcast_packets_received_total"),
			"This represents the total number of broadcast packets received per second by the virtual switch",
			nil,
			nil,
		),
		BroadcastPacketsSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "broadcast_packets_sent_total"),
			"This represents the total number of broadcast packets sent per second by the virtual switch",
			nil,
			nil,
		),
		Bytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "bytes_total"),
			"This represents the total number of bytes per second traversing the virtual switch",
			nil,
			nil,
		),
		BytesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "bytes_received_total"),
			"This represents the total number of bytes received per second by the virtual switch",
			nil,
			nil,
		),
		BytesSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "bytes_sent_total"),
			"This represents the total number of bytes sent per second by the virtual switch",
			nil,
			nil,
		),
		DirectedPacketsReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "directed_packets_received_total"),
			"This represents the total number of directed packets received per second by the virtual switch",
			nil,
			nil,
		),
		DirectedPacketsSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "directed_packets_send_total"),
			"This represents the total number of directed packets sent per second by the virtual switch",
			nil,
			nil,
		),
		DroppedPacketsIncoming: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "dropped_packets_incoming_total"),
			"This represents the total number of packet dropped per second by the virtual switch in the incoming direction",
			nil,
			nil,
		),
		DroppedPacketsOutgoing: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "dropped_packets_outcoming_total"),
			"This represents the total number of packet dropped per second by the virtual switch in the outgoing direction",
			nil,
			nil,
		),
		ExtensionsDroppedPacketsIncoming: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "extensions_dropped_packets_incoming_total"),
			"This represents the total number of packet dropped per second by the virtual switch extensions in the incoming direction",
			nil,
			nil,
		),
		ExtensionsDroppedPacketsOutgoing: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "extensions_dropped_packets_outcoming_total"),
			"This represents the total number of packet dropped per second by the virtual switch extensions in the outgoing direction",
			nil,
			nil,
		),
		LearnedMacAddresses: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "learned_mac_addresses_total"),
			"This counter represents the total number of learned MAC addresses of the virtual switch",
			nil,
			nil,
		),
		MulticastPacketsReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "multicast_packets_received_total"),
			"This represents the total number of multicast packets received per second by the virtual switch",
			nil,
			nil,
		),
		MulticastPacketsSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "multicast_packets_sent_total"),
			"This represents the total number of multicast packets sent per second by the virtual switch",
			nil,
			nil,
		),
		NumberofSendChannelMoves: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "number_of_send_channel_moves_total"),
			"This represents the total number of send channel moves per second on this virtual switch",
			nil,
			nil,
		),
		NumberofVMQMoves: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "number_of_vmq_moves_total"),
			"This represents the total number of VMQ moves per second on this virtual switch",
			nil,
			nil,
		),
		PacketsFlooded: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "packets_flooded_total"),
			"This counter represents the total number of packets flooded by the virtual switch",
			nil,
			nil,
		),
		Packets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "packets_total"),
			"This represents the total number of packets per second traversing the virtual switch",
			nil,
			nil,
		),
		PacketsReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "packets_received_total"),
			"This represents the total number of packets received per second by the virtual switch",
			nil,
			nil,
		),
		PacketsSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "packets_sent_total"),
			"This represents the total number of packets send per second by the virtual switch",
			nil,
			nil,
		),
		PurgedMacAddresses: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "purged_mac_addresses_total"),
			"This counter represents the total number of purged MAC addresses of the virtual switch",
			nil,
			nil,
		),

		//

		AdapterBytesDropped: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("ethernet"), "bytes_dropped"),
			"Bytes Dropped is the number of bytes dropped on the network adapter",
			nil,
			nil,
		),
		AdapterBytesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("ethernet"), "bytes_received"),
			"Bytes received is the number of bytes received on the network adapter",
			nil,
			nil,
		),
		AdapterBytesSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("ethernet"), "bytes_sent"),
			"Bytes sent is the number of bytes sent over the network adapter",
			nil,
			nil,
		),
		AdapterFramesDropped: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("ethernet"), "frames_dropped"),
			"Frames Dropped is the number of frames dropped on the network adapter",
			nil,
			nil,
		),
		AdapterFramesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("ethernet"), "frames_received"),
			"Frames received is the number of frames received on the network adapter",
			nil,
			nil,
		),
		AdapterFramesSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("ethernet"), "frames_sent"),
			"Frames sent is the number of frames sent over the network adapter",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *HyperVCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collectVmHealth(ch); err != nil {
		log.Println("[ERROR] failed collecting hyperV health status metrics:", desc, err)
		return err
	}

	if desc, err := c.collectVmVid(ch); err != nil {
		log.Println("[ERROR] failed collecting hyperV pages metrics:", desc, err)
		return err
	}

	if desc, err := c.collectVmHv(ch); err != nil {
		log.Println("[ERROR] failed collecting hyperV hv status metrics:", desc, err)
		return err
	}

	if desc, err := c.collectVmProcessor(ch); err != nil {
		log.Println("[ERROR] failed collecting hyperV processor metrics:", desc, err)
		return err
	}

	if desc, err := c.collectVmRate(ch); err != nil {
		log.Println("[ERROR] failed collecting hyperV rate metrics:", desc, err)
		return err
	}

	if desc, err := c.collectVmSwitch(ch); err != nil {
		log.Println("[ERROR] failed collecting hyperV switch metrics:", desc, err)
		return err
	}

	if desc, err := c.collectVmEthernet(ch); err != nil {
		log.Println("[ERROR] failed collecting hyperV ethernet metrics:", desc, err)
		return err
	}
	return nil
}

// Win32_PerfRawData_VmmsVirtualMachineStats_HyperVVirtualMachineHealthSummary vm health status
type Win32_PerfRawData_VmmsVirtualMachineStats_HyperVVirtualMachineHealthSummary struct {
	HealthCritical uint32
	HealthOk       uint32
}

func (c *HyperVCollector) collectVmHealth(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_VmmsVirtualMachineStats_HyperVVirtualMachineHealthSummary
	if err := wmi.Query(wmi.CreateQuery(&dst, ""), &dst); err != nil {
		return nil, err
	}

	for _, health := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.HealthCritical,
			prometheus.GaugeValue,
			float64(health.HealthCritical),
		)

		ch <- prometheus.MustNewConstMetric(
			c.HealthOk,
			prometheus.GaugeValue,
			float64(health.HealthOk),
		)

	}

	return nil, nil
}

// Win32_PerfRawData_VidPerfProvider_HyperVVMVidPartition ..,
type Win32_PerfRawData_VidPerfProvider_HyperVVMVidPartition struct {
	Name                   string
	PhysicalPagesAllocated uint64
	PreferredNUMANodeIndex uint64
	RemotePhysicalPages    uint64
}

func (c *HyperVCollector) collectVmVid(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_VidPerfProvider_HyperVVMVidPartition
	if err := wmi.Query(wmi.CreateQuery(&dst, ""), &dst); err != nil {
		return nil, err
	}

	for _, page := range dst {
		if strings.Contains(page.Name, "_Total") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.PhysicalPagesAllocated,
			prometheus.GaugeValue,
			float64(page.PhysicalPagesAllocated),
		)

		ch <- prometheus.MustNewConstMetric(
			c.PreferredNUMANodeIndex,
			prometheus.GaugeValue,
			float64(page.PreferredNUMANodeIndex),
		)

		ch <- prometheus.MustNewConstMetric(
			c.RemotePhysicalPages,
			prometheus.GaugeValue,
			float64(page.RemotePhysicalPages),
		)

	}

	return nil, nil
}

// Win32_PerfRawData_HvStats_HyperVHypervisorRootPartition ...
type Win32_PerfRawData_HvStats_HyperVHypervisorRootPartition struct {
	Name                          string
	AddressSpaces                 uint64
	AttachedDevices               uint64
	DepositedPages                uint64
	DeviceDMAErrors               uint64
	DeviceInterruptErrors         uint64
	DeviceInterruptMappings       uint64
	DeviceInterruptThrottleEvents uint64
	GPAPages                      uint64
	GPASpaceModificationsPersec   uint64
	IOTLBFlushCost                uint64
	IOTLBFlushesPersec            uint64
	RecommendedVirtualTLBSize     uint64
	SkippedTimerTicks             uint64
	Value1Gdevicepages            uint64
	Value1GGPApages               uint64
	Value2Mdevicepages            uint64
	Value2MGPApages               uint64
	Value4Kdevicepages            uint64
	Value4KGPApages               uint64
	VirtualTLBFlushEntiresPersec  uint64
	VirtualTLBPages               uint64
}

func (c *HyperVCollector) collectVmHv(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_HvStats_HyperVHypervisorRootPartition
	if err := wmi.Query(wmi.CreateQuery(&dst, ""), &dst); err != nil {
		return nil, err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "_Total") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.AddressSpaces,
			prometheus.GaugeValue,
			float64(obj.AddressSpaces),
		)

		ch <- prometheus.MustNewConstMetric(
			c.AttachedDevices,
			prometheus.GaugeValue,
			float64(obj.AttachedDevices),
		)

		ch <- prometheus.MustNewConstMetric(
			c.DepositedPages,
			prometheus.GaugeValue,
			float64(obj.DepositedPages),
		)

		ch <- prometheus.MustNewConstMetric(
			c.DeviceDMAErrors,
			prometheus.GaugeValue,
			float64(obj.DeviceDMAErrors),
		)

		ch <- prometheus.MustNewConstMetric(
			c.DeviceInterruptErrors,
			prometheus.GaugeValue,
			float64(obj.DeviceInterruptErrors),
		)

		ch <- prometheus.MustNewConstMetric(
			c.DeviceInterruptThrottleEvents,
			prometheus.GaugeValue,
			float64(obj.DeviceInterruptThrottleEvents),
		)

		ch <- prometheus.MustNewConstMetric(
			c.GPAPages,
			prometheus.GaugeValue,
			float64(obj.GPAPages),
		)

		ch <- prometheus.MustNewConstMetric(
			c.GPASpaceModifications,
			prometheus.CounterValue,
			float64(obj.GPASpaceModificationsPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOTLBFlushCost,
			prometheus.GaugeValue,
			float64(obj.IOTLBFlushCost),
		)

		ch <- prometheus.MustNewConstMetric(
			c.IOTLBFlushes,
			prometheus.CounterValue,
			float64(obj.IOTLBFlushesPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.RecommendedVirtualTLBSize,
			prometheus.GaugeValue,
			float64(obj.RecommendedVirtualTLBSize),
		)

		ch <- prometheus.MustNewConstMetric(
			c.SkippedTimerTicks,
			prometheus.GaugeValue,
			float64(obj.SkippedTimerTicks),
		)

		ch <- prometheus.MustNewConstMetric(
			c.Value1Gdevicepages,
			prometheus.GaugeValue,
			float64(obj.Value1Gdevicepages),
		)

		ch <- prometheus.MustNewConstMetric(
			c.Value1GGPApages,
			prometheus.GaugeValue,
			float64(obj.Value1GGPApages),
		)

		ch <- prometheus.MustNewConstMetric(
			c.Value2Mdevicepages,
			prometheus.GaugeValue,
			float64(obj.Value2Mdevicepages),
		)
		ch <- prometheus.MustNewConstMetric(
			c.Value2MGPApages,
			prometheus.GaugeValue,
			float64(obj.Value2MGPApages),
		)
		ch <- prometheus.MustNewConstMetric(
			c.Value4Kdevicepages,
			prometheus.GaugeValue,
			float64(obj.Value4Kdevicepages),
		)
		ch <- prometheus.MustNewConstMetric(
			c.Value4KGPApages,
			prometheus.GaugeValue,
			float64(obj.Value4KGPApages),
		)
		ch <- prometheus.MustNewConstMetric(
			c.VirtualTLBFlushEntires,
			prometheus.CounterValue,
			float64(obj.VirtualTLBFlushEntiresPersec),
		)
		ch <- prometheus.MustNewConstMetric(
			c.VirtualTLBPages,
			prometheus.GaugeValue,
			float64(obj.VirtualTLBPages),
		)

	}

	return nil, nil
}

// Win32_PerfRawData_HvStats_HyperVHypervisor ...
type Win32_PerfRawData_HvStats_HyperVHypervisor struct {
	LogicalProcessors uint64
	VirtualProcessors uint64
}

func (c *HyperVCollector) collectVmProcessor(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_HvStats_HyperVHypervisor
	if err := wmi.Query(wmi.CreateQuery(&dst, ""), &dst); err != nil {
		return nil, err
	}

	for _, obj := range dst {

		ch <- prometheus.MustNewConstMetric(
			c.LogicalProcessors,
			prometheus.GaugeValue,
			float64(obj.LogicalProcessors),
		)

		ch <- prometheus.MustNewConstMetric(
			c.VirtualProcessors,
			prometheus.GaugeValue,
			float64(obj.VirtualProcessors),
		)

	}

	return nil, nil
}

// Win32_PerfRawData_HvStats_HyperVHypervisorRootVirtualProcessor ...
type Win32_PerfRawData_HvStats_HyperVHypervisorRootVirtualProcessor struct {
	Name                     string
	PercentGuestRunTime      uint64
	PercentHypervisorRunTime uint64
	PercentRemoteRunTime     uint64
	PercentTotalRunTime      uint64
}

func (c *HyperVCollector) collectVmRate(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_HvStats_HyperVHypervisorRootVirtualProcessor
	if err := wmi.Query(wmi.CreateQuery(&dst, ""), &dst); err != nil {
		return nil, err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "_Total") {
			continue
		}
		// Root VP 3
		names := strings.Split(obj.Name, " ")
		if len(names) == 0 {
			continue
		}
		label := names[len(names)-1]

		ch <- prometheus.MustNewConstMetric(
			c.PercentGuestRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentGuestRunTime),
			label,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PercentHypervisorRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentHypervisorRunTime),
			label,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PercentRemoteRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentRemoteRunTime),
			label,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PercentTotalRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentTotalRunTime),
			label,
		)

	}

	return nil, nil
}

// Win32_PerfRawData_NvspSwitchStats_HyperVVirtualSwitch ...
type Win32_PerfRawData_NvspSwitchStats_HyperVVirtualSwitch struct {
	Name                                   string
	BroadcastPacketsReceivedPersec         uint64
	BroadcastPacketsSentPersec             uint64
	BytesPersec                            uint64
	BytesReceivedPersec                    uint64
	BytesSentPersec                        uint64
	DirectedPacketsReceivedPersec          uint64
	DirectedPacketsSentPersec              uint64
	DroppedPacketsIncomingPersec           uint64
	DroppedPacketsOutgoingPersec           uint64
	ExtensionsDroppedPacketsIncomingPersec uint64
	ExtensionsDroppedPacketsOutgoingPersec uint64
	LearnedMacAddresses                    uint64
	LearnedMacAddressesPersec              uint64
	MulticastPacketsReceivedPersec         uint64
	MulticastPacketsSentPersec             uint64
	NumberofSendChannelMovesPersec         uint64
	NumberofVMQMovesPersec                 uint64
	PacketsFlooded                         uint64
	PacketsFloodedPersec                   uint64
	PacketsPersec                          uint64
	PacketsReceivedPersec                  uint64
	PacketsSentPersec                      uint64
	PurgedMacAddresses                     uint64
	PurgedMacAddressesPersec               uint64
}

func (c *HyperVCollector) collectVmSwitch(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_NvspSwitchStats_HyperVVirtualSwitch
	if err := wmi.Query(wmi.CreateQuery(&dst, ""), &dst); err != nil {
		return nil, err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "_Total") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.BroadcastPacketsReceived,
			prometheus.CounterValue,
			float64(obj.BroadcastPacketsReceivedPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.BroadcastPacketsSent,
			prometheus.CounterValue,
			float64(obj.BroadcastPacketsSentPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.Bytes,
			prometheus.CounterValue,
			float64(obj.BytesPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.BytesReceived,
			prometheus.CounterValue,
			float64(obj.BytesReceivedPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.BytesSent,
			prometheus.CounterValue,
			float64(obj.BytesSentPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.DirectedPacketsReceived,
			prometheus.CounterValue,
			float64(obj.DirectedPacketsReceivedPersec),
		)
		ch <- prometheus.MustNewConstMetric(
			c.DirectedPacketsSent,
			prometheus.CounterValue,
			float64(obj.DirectedPacketsSentPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.DroppedPacketsIncoming,
			prometheus.CounterValue,
			float64(obj.DroppedPacketsIncomingPersec),
		)
		ch <- prometheus.MustNewConstMetric(
			c.DroppedPacketsOutgoing,
			prometheus.CounterValue,
			float64(obj.DroppedPacketsOutgoingPersec),
		)
		ch <- prometheus.MustNewConstMetric(
			c.ExtensionsDroppedPacketsIncoming,
			prometheus.CounterValue,
			float64(obj.ExtensionsDroppedPacketsIncomingPersec),
		)
		ch <- prometheus.MustNewConstMetric(
			c.ExtensionsDroppedPacketsOutgoing,
			prometheus.CounterValue,
			float64(obj.ExtensionsDroppedPacketsOutgoingPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.LearnedMacAddresses,
			prometheus.CounterValue,
			float64(obj.LearnedMacAddresses),
		)
		ch <- prometheus.MustNewConstMetric(
			c.MulticastPacketsReceived,
			prometheus.CounterValue,
			float64(obj.MulticastPacketsReceivedPersec),
		)
		ch <- prometheus.MustNewConstMetric(
			c.MulticastPacketsSent,
			prometheus.CounterValue,
			float64(obj.MulticastPacketsSentPersec),
		)
		ch <- prometheus.MustNewConstMetric(
			c.NumberofSendChannelMoves,
			prometheus.CounterValue,
			float64(obj.NumberofSendChannelMovesPersec),
		)
		ch <- prometheus.MustNewConstMetric(
			c.NumberofVMQMoves,
			prometheus.CounterValue,
			float64(obj.NumberofVMQMovesPersec),
		)

		// ...
		ch <- prometheus.MustNewConstMetric(
			c.PacketsFlooded,
			prometheus.CounterValue,
			float64(obj.PacketsFlooded),
		)

		ch <- prometheus.MustNewConstMetric(
			c.Packets,
			prometheus.CounterValue,
			float64(obj.PacketsPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.PacketsReceived,
			prometheus.CounterValue,
			float64(obj.PacketsReceivedPersec),
		)
		ch <- prometheus.MustNewConstMetric(
			c.PurgedMacAddresses,
			prometheus.CounterValue,
			float64(obj.PurgedMacAddresses),
		)
	}

	return nil, nil
}

// Win32_PerfRawData_EthernetPerfProvider_HyperVLegacyNetworkAdapter ...
type Win32_PerfRawData_EthernetPerfProvider_HyperVLegacyNetworkAdapter struct {
	Name                 string
	BytesDropped         uint64
	BytesReceivedPersec  uint64
	BytesSentPersec      uint64
	FramesDropped        uint64
	FramesReceivedPersec uint64
	FramesSentPersec     uint64
}

func (c *HyperVCollector) collectVmEthernet(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_EthernetPerfProvider_HyperVLegacyNetworkAdapter
	if err := wmi.Query(wmi.CreateQuery(&dst, ""), &dst); err != nil {
		return nil, err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "_Total") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.AdapterBytesDropped,
			prometheus.GaugeValue,
			float64(obj.BytesDropped),
		)

		ch <- prometheus.MustNewConstMetric(
			c.AdapterBytesReceived,
			prometheus.CounterValue,
			float64(obj.BytesReceivedPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.AdapterBytesSent,
			prometheus.CounterValue,
			float64(obj.BytesSentPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.AdapterFramesReceived,
			prometheus.CounterValue,
			float64(obj.FramesReceivedPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.AdapterFramesDropped,
			prometheus.CounterValue,
			float64(obj.FramesDropped),
		)

		ch <- prometheus.MustNewConstMetric(
			c.AdapterFramesSent,
			prometheus.CounterValue,
			float64(obj.FramesSentPersec),
		)

	}

	return nil, nil
}
