//go:build windows
// +build windows

package collector

import (
	"strings"

	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("hyperv", NewHyperVCollector)
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

	// Win32_PerfRawData_HvStats_HyperVHypervisorLogicalProcessor
	HostLPGuestRunTimePercent      *prometheus.Desc
	HostLPHypervisorRunTimePercent *prometheus.Desc
	HostLPTotalRunTimePercent      *prometheus.Desc

	// Win32_PerfRawData_HvStats_HyperVHypervisorRootVirtualProcessor
	HostGuestRunTime      *prometheus.Desc
	HostHypervisorRunTime *prometheus.Desc
	HostRemoteRunTime     *prometheus.Desc
	HostTotalRunTime      *prometheus.Desc

	// Win32_PerfRawData_HvStats_HyperVHypervisorVirtualProcessor
	VMGuestRunTime      *prometheus.Desc
	VMHypervisorRunTime *prometheus.Desc
	VMRemoteRunTime     *prometheus.Desc
	VMTotalRunTime      *prometheus.Desc

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

	// Win32_PerfRawData_Counters_HyperVVirtualStorageDevice
	VMStorageErrorCount      *prometheus.Desc
	VMStorageQueueLength     *prometheus.Desc
	VMStorageReadBytes       *prometheus.Desc
	VMStorageReadOperations  *prometheus.Desc
	VMStorageWriteBytes      *prometheus.Desc
	VMStorageWriteOperations *prometheus.Desc

	// Win32_PerfRawData_NvspNicStats_HyperVVirtualNetworkAdapter
	VMNetworkBytesReceived          *prometheus.Desc
	VMNetworkBytesSent              *prometheus.Desc
	VMNetworkDroppedPacketsIncoming *prometheus.Desc
	VMNetworkDroppedPacketsOutgoing *prometheus.Desc
	VMNetworkPacketsReceived        *prometheus.Desc
	VMNetworkPacketsSent            *prometheus.Desc

	// Win32_PerfRawData_BalancerStats_HyperVDynamicMemoryVM
	VMMemoryAddedMemory                *prometheus.Desc
	VMMemoryAveragePressure            *prometheus.Desc
	VMMemoryCurrentPressure            *prometheus.Desc
	VMMemoryGuestVisiblePhysicalMemory *prometheus.Desc
	VMMemoryMaximumPressure            *prometheus.Desc
	VMMemoryMemoryAddOperations        *prometheus.Desc
	VMMemoryMemoryRemoveOperations     *prometheus.Desc
	VMMemoryMinimumPressure            *prometheus.Desc
	VMMemoryPhysicalMemory             *prometheus.Desc
	VMMemoryRemovedMemory              *prometheus.Desc
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
			[]string{"vm"},
			nil,
		),
		PreferredNUMANodeIndex: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vid"), "preferred_numa_node_index"),
			"The preferred NUMA node index associated with this partition",
			[]string{"vm"},
			nil,
		),
		RemotePhysicalPages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vid"), "remote_physical_pages"),
			"The number of physical pages not allocated from the preferred NUMA node",
			[]string{"vm"},
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

		HostLPGuestRunTimePercent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("host_lp"), "guest_run_time_percent"),
			"The percentage of time spent by the processor in guest code",
			[]string{"core"},
			nil,
		),
		HostLPHypervisorRunTimePercent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("host_lp"), "hypervisor_run_time_percent"),
			"The percentage of time spent by the processor in hypervisor code",
			[]string{"core"},
			nil,
		),
		HostLPTotalRunTimePercent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("host_lp"), "total_run_time_percent"),
			"The percentage of time spent by the processor in guest and hypervisor code",
			[]string{"core"},
			nil,
		),

		//

		HostGuestRunTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("host_cpu"), "guest_run_time"),
			"The time spent by the virtual processor in guest code",
			[]string{"core"},
			nil,
		),
		HostHypervisorRunTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("host_cpu"), "hypervisor_run_time"),
			"The time spent by the virtual processor in hypervisor code",
			[]string{"core"},
			nil,
		),
		HostRemoteRunTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("host_cpu"), "remote_run_time"),
			"The time spent by the virtual processor running on a remote node",
			[]string{"core"},
			nil,
		),
		HostTotalRunTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("host_cpu"), "total_run_time"),
			"The time spent by the virtual processor in guest and hypervisor code",
			[]string{"core"},
			nil,
		),

		//

		VMGuestRunTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_cpu"), "guest_run_time"),
			"The time spent by the virtual processor in guest code",
			[]string{"vm", "core"},
			nil,
		),
		VMHypervisorRunTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_cpu"), "hypervisor_run_time"),
			"The time spent by the virtual processor in hypervisor code",
			[]string{"vm", "core"},
			nil,
		),
		VMRemoteRunTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_cpu"), "remote_run_time"),
			"The time spent by the virtual processor running on a remote node",
			[]string{"vm", "core"},
			nil,
		),
		VMTotalRunTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_cpu"), "total_run_time"),
			"The time spent by the virtual processor in guest and hypervisor code",
			[]string{"vm", "core"},
			nil,
		),

		//
		BroadcastPacketsReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "broadcast_packets_received_total"),
			"This represents the total number of broadcast packets received per second by the virtual switch",
			[]string{"vswitch"},
			nil,
		),
		BroadcastPacketsSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "broadcast_packets_sent_total"),
			"This represents the total number of broadcast packets sent per second by the virtual switch",
			[]string{"vswitch"},
			nil,
		),
		Bytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "bytes_total"),
			"This represents the total number of bytes per second traversing the virtual switch",
			[]string{"vswitch"},
			nil,
		),
		BytesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "bytes_received_total"),
			"This represents the total number of bytes received per second by the virtual switch",
			[]string{"vswitch"},
			nil,
		),
		BytesSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "bytes_sent_total"),
			"This represents the total number of bytes sent per second by the virtual switch",
			[]string{"vswitch"},
			nil,
		),
		DirectedPacketsReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "directed_packets_received_total"),
			"This represents the total number of directed packets received per second by the virtual switch",
			[]string{"vswitch"},
			nil,
		),
		DirectedPacketsSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "directed_packets_send_total"),
			"This represents the total number of directed packets sent per second by the virtual switch",
			[]string{"vswitch"},
			nil,
		),
		DroppedPacketsIncoming: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "dropped_packets_incoming_total"),
			"This represents the total number of packet dropped per second by the virtual switch in the incoming direction",
			[]string{"vswitch"},
			nil,
		),
		DroppedPacketsOutgoing: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "dropped_packets_outcoming_total"),
			"This represents the total number of packet dropped per second by the virtual switch in the outgoing direction",
			[]string{"vswitch"},
			nil,
		),
		ExtensionsDroppedPacketsIncoming: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "extensions_dropped_packets_incoming_total"),
			"This represents the total number of packet dropped per second by the virtual switch extensions in the incoming direction",
			[]string{"vswitch"},
			nil,
		),
		ExtensionsDroppedPacketsOutgoing: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "extensions_dropped_packets_outcoming_total"),
			"This represents the total number of packet dropped per second by the virtual switch extensions in the outgoing direction",
			[]string{"vswitch"},
			nil,
		),
		LearnedMacAddresses: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "learned_mac_addresses_total"),
			"This counter represents the total number of learned MAC addresses of the virtual switch",
			[]string{"vswitch"},
			nil,
		),
		MulticastPacketsReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "multicast_packets_received_total"),
			"This represents the total number of multicast packets received per second by the virtual switch",
			[]string{"vswitch"},
			nil,
		),
		MulticastPacketsSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "multicast_packets_sent_total"),
			"This represents the total number of multicast packets sent per second by the virtual switch",
			[]string{"vswitch"},
			nil,
		),
		NumberofSendChannelMoves: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "number_of_send_channel_moves_total"),
			"This represents the total number of send channel moves per second on this virtual switch",
			[]string{"vswitch"},
			nil,
		),
		NumberofVMQMoves: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "number_of_vmq_moves_total"),
			"This represents the total number of VMQ moves per second on this virtual switch",
			[]string{"vswitch"},
			nil,
		),
		PacketsFlooded: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "packets_flooded_total"),
			"This counter represents the total number of packets flooded by the virtual switch",
			[]string{"vswitch"},
			nil,
		),
		Packets: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "packets_total"),
			"This represents the total number of packets per second traversing the virtual switch",
			[]string{"vswitch"},
			nil,
		),
		PacketsReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "packets_received_total"),
			"This represents the total number of packets received per second by the virtual switch",
			[]string{"vswitch"},
			nil,
		),
		PacketsSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "packets_sent_total"),
			"This represents the total number of packets send per second by the virtual switch",
			[]string{"vswitch"},
			nil,
		),
		PurgedMacAddresses: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vswitch"), "purged_mac_addresses_total"),
			"This counter represents the total number of purged MAC addresses of the virtual switch",
			[]string{"vswitch"},
			nil,
		),

		//

		AdapterBytesDropped: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("ethernet"), "bytes_dropped"),
			"Bytes Dropped is the number of bytes dropped on the network adapter",
			[]string{"adapter"},
			nil,
		),
		AdapterBytesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("ethernet"), "bytes_received"),
			"Bytes received is the number of bytes received on the network adapter",
			[]string{"adapter"},
			nil,
		),
		AdapterBytesSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("ethernet"), "bytes_sent"),
			"Bytes sent is the number of bytes sent over the network adapter",
			[]string{"adapter"},
			nil,
		),
		AdapterFramesDropped: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("ethernet"), "frames_dropped"),
			"Frames Dropped is the number of frames dropped on the network adapter",
			[]string{"adapter"},
			nil,
		),
		AdapterFramesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("ethernet"), "frames_received"),
			"Frames received is the number of frames received on the network adapter",
			[]string{"adapter"},
			nil,
		),
		AdapterFramesSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("ethernet"), "frames_sent"),
			"Frames sent is the number of frames sent over the network adapter",
			[]string{"adapter"},
			nil,
		),

		//

		VMStorageErrorCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_device"), "error_count"),
			"This counter represents the total number of errors that have occurred on this virtual device",
			[]string{"vm_device"},
			nil,
		),
		VMStorageQueueLength: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_device"), "queue_length"),
			"This counter represents the current queue length on this virtual device",
			[]string{"vm_device"},
			nil,
		),
		VMStorageReadBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_device"), "bytes_read"),
			"This counter represents the total number of bytes that have been read per second on this virtual device",
			[]string{"vm_device"},
			nil,
		),
		VMStorageReadOperations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_device"), "operations_read"),
			"This counter represents the number of read operations that have occurred per second on this virtual device",
			[]string{"vm_device"},
			nil,
		),
		VMStorageWriteBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_device"), "bytes_written"),
			"This counter represents the total number of bytes that have been written per second on this virtual device",
			[]string{"vm_device"},
			nil,
		),
		VMStorageWriteOperations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_device"), "operations_written"),
			"This counter represents the number of write operations that have occurred per second on this virtual device",
			[]string{"vm_device"},
			nil,
		),

		//

		VMNetworkBytesReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_interface"), "bytes_received"),
			"This counter represents the total number of bytes received per second by the network adapter",
			[]string{"vm_interface"},
			nil,
		),
		VMNetworkBytesSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_interface"), "bytes_sent"),
			"This counter represents the total number of bytes sent per second by the network adapter",
			[]string{"vm_interface"},
			nil,
		),
		VMNetworkDroppedPacketsIncoming: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_interface"), "packets_incoming_dropped"),
			"This counter represents the total number of dropped packets per second in the incoming direction of the network adapter",
			[]string{"vm_interface"},
			nil,
		),
		VMNetworkDroppedPacketsOutgoing: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_interface"), "packets_outgoing_dropped"),
			"This counter represents the total number of dropped packets per second in the outgoing direction of the network adapter",
			[]string{"vm_interface"},
			nil,
		),
		VMNetworkPacketsReceived: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_interface"), "packets_received"),
			"This counter represents the total number of packets received per second by the network adapter",
			[]string{"vm_interface"},
			nil,
		),
		VMNetworkPacketsSent: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_interface"), "packets_sent"),
			"This counter represents the total number of packets sent per second by the network adapter",
			[]string{"vm_interface"},
			nil,
		),

		//

		VMMemoryAddedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_memory"), "added_total"),
			"This counter represents memory in MB added to the VM",
			[]string{"vm"},
			nil,
		),
		VMMemoryAveragePressure: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_memory"), "pressure_average"),
			"This gauge represents the average pressure in the VM.",
			[]string{"vm"},
			nil,
		),
		VMMemoryCurrentPressure: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_memory"), "pressure_current"),
			"This gauge represents the current pressure in the VM.",
			[]string{"vm"},
			nil,
		),
		VMMemoryGuestVisiblePhysicalMemory: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_memory"), "physical_guest_visible"),
			"'This gauge represents the amount of memory in MB visible to the VM guest.'",
			[]string{"vm"},
			nil,
		),
		VMMemoryMaximumPressure: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_memory"), "pressure_maximum"),
			"This gauge represents the maximum pressure band in the VM.",
			[]string{"vm"},
			nil,
		),
		VMMemoryMemoryAddOperations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_memory"), "add_operations_total"),
			"This counter represents the number of operations adding memory to the VM.",
			[]string{"vm"},
			nil,
		),
		VMMemoryMemoryRemoveOperations: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_memory"), "remove_operations_total"),
			"This counter represents the number of operations removing memory from the VM.",
			[]string{"vm"},
			nil,
		),
		VMMemoryMinimumPressure: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_memory"), "pressure_minimum"),
			"This gauge represents the minimum pressure band in the VM.",
			[]string{"vm"},
			nil,
		),
		VMMemoryPhysicalMemory: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_memory"), "physical"),
			"This gauge represents the current amount of memory in MB assigned to the VM.",
			[]string{"vm"},
			nil,
		),
		VMMemoryRemovedMemory: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, buildSubsystemName("vm_memory"), "removed_total"),
			"This counter represents memory in MB removed from the VM",
			[]string{"vm"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *HyperVCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collectVmHealth(ch); err != nil {
		log.Error("failed collecting hyperV health status metrics:", desc, err)
		return err
	}

	if desc, err := c.collectVmVid(ch); err != nil {
		log.Error("failed collecting hyperV pages metrics:", desc, err)
		return err
	}

	if desc, err := c.collectVmHv(ch); err != nil {
		log.Error("failed collecting hyperV hv status metrics:", desc, err)
		return err
	}

	if desc, err := c.collectVmProcessor(ch); err != nil {
		log.Error("failed collecting hyperV processor metrics:", desc, err)
		return err
	}

	if desc, err := c.collectHostLPUsage(ch); err != nil {
		log.Error("failed collecting hyperV host logical processors metrics:", desc, err)
		return err
	}

	if desc, err := c.collectHostCpuUsage(ch); err != nil {
		log.Error("failed collecting hyperV host CPU metrics:", desc, err)
		return err
	}

	if desc, err := c.collectVmCpuUsage(ch); err != nil {
		log.Error("failed collecting hyperV VM CPU metrics:", desc, err)
		return err
	}

	if desc, err := c.collectVmSwitch(ch); err != nil {
		log.Error("failed collecting hyperV switch metrics:", desc, err)
		return err
	}

	if desc, err := c.collectVmEthernet(ch); err != nil {
		log.Error("failed collecting hyperV ethernet metrics:", desc, err)
		return err
	}

	if desc, err := c.collectVmStorage(ch); err != nil {
		log.Error("failed collecting hyperV virtual storage metrics:", desc, err)
		return err
	}

	if desc, err := c.collectVmNetwork(ch); err != nil {
		log.Error("failed collecting hyperV virtual network metrics:", desc, err)
		return err
	}

	if desc, err := c.collectVmMemory(ch); err != nil {
		log.Error("failed collecting hyperV virtual memory metrics:", desc, err)
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
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
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
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
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
			page.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PreferredNUMANodeIndex,
			prometheus.GaugeValue,
			float64(page.PreferredNUMANodeIndex),
			page.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RemotePhysicalPages,
			prometheus.GaugeValue,
			float64(page.RemotePhysicalPages),
			page.Name,
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
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
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
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
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

// Win32_PerfRawData_HvStats_HyperVHypervisorLogicalProcessor ...
type Win32_PerfRawData_HvStats_HyperVHypervisorLogicalProcessor struct {
	Name                     string
	PercentGuestRunTime      uint64
	PercentHypervisorRunTime uint64
	PercentTotalRunTime      uint
}

func (c *HyperVCollector) collectHostLPUsage(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_HvStats_HyperVHypervisorLogicalProcessor
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "_Total") {
			continue
		}
		// The name format is Hv LP <core id>
		parts := strings.Split(obj.Name, " ")
		if len(parts) != 3 {
			log.Warnf("Unexpected format of Name in collectHostLPUsage: %q", obj.Name)
			continue
		}
		coreId := parts[2]

		ch <- prometheus.MustNewConstMetric(
			c.HostLPGuestRunTimePercent,
			prometheus.GaugeValue,
			float64(obj.PercentGuestRunTime),
			coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.HostLPHypervisorRunTimePercent,
			prometheus.GaugeValue,
			float64(obj.PercentHypervisorRunTime),
			coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.HostLPTotalRunTimePercent,
			prometheus.GaugeValue,
			float64(obj.PercentTotalRunTime),
			coreId,
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

func (c *HyperVCollector) collectHostCpuUsage(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_HvStats_HyperVHypervisorRootVirtualProcessor
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "_Total") {
			continue
		}
		// The name format is Root VP <core id>
		parts := strings.Split(obj.Name, " ")
		if len(parts) != 3 {
			log.Warnf("Unexpected format of Name in collectHostCpuUsage: %q", obj.Name)
			continue
		}
		coreId := parts[2]

		ch <- prometheus.MustNewConstMetric(
			c.HostGuestRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentGuestRunTime),
			coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.HostHypervisorRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentHypervisorRunTime),
			coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.HostRemoteRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentRemoteRunTime),
			coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.HostTotalRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentTotalRunTime),
			coreId,
		)

	}

	return nil, nil
}

// Win32_PerfRawData_HvStats_HyperVHypervisorVirtualProcessor ...
type Win32_PerfRawData_HvStats_HyperVHypervisorVirtualProcessor struct {
	Name                     string
	PercentGuestRunTime      uint64
	PercentHypervisorRunTime uint64
	PercentRemoteRunTime     uint64
	PercentTotalRunTime      uint64
}

func (c *HyperVCollector) collectVmCpuUsage(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_HvStats_HyperVHypervisorVirtualProcessor
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "_Total") {
			continue
		}
		// The name format is <VM Name>:Hv VP <vcore id>
		parts := strings.Split(obj.Name, ":")
		if len(parts) != 2 {
			log.Warnf("Unexpected format of Name in collectVmCpuUsage: %q, expected %q. Skipping.", obj.Name, "<VM Name>:Hv VP <vcore id>")
			continue
		}
		coreParts := strings.Split(parts[1], " ")
		if len(coreParts) != 3 {
			log.Warnf("Unexpected format of core identifier in collectVmCpuUsage: %q, expected %q. Skipping.", parts[1], "Hv VP <vcore id>")
			continue
		}
		vmName := parts[0]
		coreId := coreParts[2]

		ch <- prometheus.MustNewConstMetric(
			c.VMGuestRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentGuestRunTime),
			vmName, coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMHypervisorRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentHypervisorRunTime),
			vmName, coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMRemoteRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentRemoteRunTime),
			vmName, coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMTotalRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentTotalRunTime),
			vmName, coreId,
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
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
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
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BroadcastPacketsSent,
			prometheus.CounterValue,
			float64(obj.BroadcastPacketsSentPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Bytes,
			prometheus.CounterValue,
			float64(obj.BytesPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BytesReceived,
			prometheus.CounterValue,
			float64(obj.BytesReceivedPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BytesSent,
			prometheus.CounterValue,
			float64(obj.BytesSentPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DirectedPacketsReceived,
			prometheus.CounterValue,
			float64(obj.DirectedPacketsReceivedPersec),
			obj.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.DirectedPacketsSent,
			prometheus.CounterValue,
			float64(obj.DirectedPacketsSentPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DroppedPacketsIncoming,
			prometheus.CounterValue,
			float64(obj.DroppedPacketsIncomingPersec),
			obj.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.DroppedPacketsOutgoing,
			prometheus.CounterValue,
			float64(obj.DroppedPacketsOutgoingPersec),
			obj.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ExtensionsDroppedPacketsIncoming,
			prometheus.CounterValue,
			float64(obj.ExtensionsDroppedPacketsIncomingPersec),
			obj.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ExtensionsDroppedPacketsOutgoing,
			prometheus.CounterValue,
			float64(obj.ExtensionsDroppedPacketsOutgoingPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.LearnedMacAddresses,
			prometheus.CounterValue,
			float64(obj.LearnedMacAddresses),
			obj.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.MulticastPacketsReceived,
			prometheus.CounterValue,
			float64(obj.MulticastPacketsReceivedPersec),
			obj.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.MulticastPacketsSent,
			prometheus.CounterValue,
			float64(obj.MulticastPacketsSentPersec),
			obj.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.NumberofSendChannelMoves,
			prometheus.CounterValue,
			float64(obj.NumberofSendChannelMovesPersec),
			obj.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.NumberofVMQMoves,
			prometheus.CounterValue,
			float64(obj.NumberofVMQMovesPersec),
			obj.Name,
		)

		// ...
		ch <- prometheus.MustNewConstMetric(
			c.PacketsFlooded,
			prometheus.CounterValue,
			float64(obj.PacketsFlooded),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Packets,
			prometheus.CounterValue,
			float64(obj.PacketsPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PacketsReceived,
			prometheus.CounterValue,
			float64(obj.PacketsReceivedPersec),
			obj.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.PurgedMacAddresses,
			prometheus.CounterValue,
			float64(obj.PurgedMacAddresses),
			obj.Name,
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
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
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
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AdapterBytesReceived,
			prometheus.CounterValue,
			float64(obj.BytesReceivedPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AdapterBytesSent,
			prometheus.CounterValue,
			float64(obj.BytesSentPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AdapterFramesReceived,
			prometheus.CounterValue,
			float64(obj.FramesReceivedPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AdapterFramesDropped,
			prometheus.CounterValue,
			float64(obj.FramesDropped),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.AdapterFramesSent,
			prometheus.CounterValue,
			float64(obj.FramesSentPersec),
			obj.Name,
		)

	}

	return nil, nil
}

// Win32_PerfRawData_Counters_HyperVVirtualStorageDevice ...
type Win32_PerfRawData_Counters_HyperVVirtualStorageDevice struct {
	Name                  string
	ErrorCount            uint64
	QueueLength           uint32
	ReadBytesPersec       uint64
	ReadOperationsPerSec  uint64
	WriteBytesPersec      uint64
	WriteOperationsPerSec uint64
}

func (c *HyperVCollector) collectVmStorage(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_HyperVVirtualStorageDevice
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "_Total") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.VMStorageErrorCount,
			prometheus.CounterValue,
			float64(obj.ErrorCount),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMStorageQueueLength,
			prometheus.CounterValue,
			float64(obj.QueueLength),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMStorageReadBytes,
			prometheus.CounterValue,
			float64(obj.ReadBytesPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMStorageReadOperations,
			prometheus.CounterValue,
			float64(obj.ReadOperationsPerSec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMStorageWriteBytes,
			prometheus.CounterValue,
			float64(obj.WriteBytesPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMStorageWriteOperations,
			prometheus.CounterValue,
			float64(obj.WriteOperationsPerSec),
			obj.Name,
		)
	}

	return nil, nil
}

// Win32_PerfRawData_NvspNicStats_HyperVVirtualNetworkAdapter ...
type Win32_PerfRawData_NvspNicStats_HyperVVirtualNetworkAdapter struct {
	Name                         string
	BytesReceivedPersec          uint64
	BytesSentPersec              uint64
	DroppedPacketsIncomingPersec uint64
	DroppedPacketsOutgoingPersec uint64
	PacketsReceivedPersec        uint64
	PacketsSentPersec            uint64
}

func (c *HyperVCollector) collectVmNetwork(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_NvspNicStats_HyperVVirtualNetworkAdapter
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "_Total") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.VMNetworkBytesReceived,
			prometheus.CounterValue,
			float64(obj.BytesReceivedPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMNetworkBytesSent,
			prometheus.CounterValue,
			float64(obj.BytesSentPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMNetworkDroppedPacketsIncoming,
			prometheus.CounterValue,
			float64(obj.DroppedPacketsIncomingPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMNetworkDroppedPacketsOutgoing,
			prometheus.CounterValue,
			float64(obj.DroppedPacketsOutgoingPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMNetworkPacketsReceived,
			prometheus.CounterValue,
			float64(obj.PacketsReceivedPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMNetworkPacketsSent,
			prometheus.CounterValue,
			float64(obj.PacketsSentPersec),
			obj.Name,
		)
	}

	return nil, nil
}

// Win32_PerfRawData_BalancerStats_HyperVDynamicMemoryVM ...
type Win32_PerfRawData_BalancerStats_HyperVDynamicMemoryVM struct {
	Name                       string
	AddedMemory                uint64
	AveragePressure            uint64
	CurrentPressure            uint64
	GuestVisiblePhysicalMemory uint64
	MaximumPressure            uint64
	MemoryAddOperations        uint64
	MemoryRemoveOperations     uint64
	MinimumPressure            uint64
	PhysicalMemory             uint64
	RemovedMemory              uint64
}

func (c *HyperVCollector) collectVmMemory(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_BalancerStats_HyperVDynamicMemoryVM
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "_Total") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.VMMemoryAddedMemory,
			prometheus.CounterValue,
			float64(obj.AddedMemory),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMMemoryAveragePressure,
			prometheus.GaugeValue,
			float64(obj.AveragePressure),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMMemoryCurrentPressure,
			prometheus.GaugeValue,
			float64(obj.CurrentPressure),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMMemoryGuestVisiblePhysicalMemory,
			prometheus.GaugeValue,
			float64(obj.GuestVisiblePhysicalMemory),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMMemoryMaximumPressure,
			prometheus.GaugeValue,
			float64(obj.MaximumPressure),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMMemoryMemoryAddOperations,
			prometheus.CounterValue,
			float64(obj.MemoryAddOperations),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMMemoryMemoryRemoveOperations,
			prometheus.CounterValue,
			float64(obj.MemoryRemoveOperations),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMMemoryMinimumPressure,
			prometheus.GaugeValue,
			float64(obj.MinimumPressure),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMMemoryPhysicalMemory,
			prometheus.GaugeValue,
			float64(obj.PhysicalMemory),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VMMemoryRemovedMemory,
			prometheus.CounterValue,
			float64(obj.RemovedMemory),
			obj.Name,
		)
	}

	return nil, nil
}
