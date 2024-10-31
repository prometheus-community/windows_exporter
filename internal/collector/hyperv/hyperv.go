//go:build windows

package hyperv

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "hyperv"

type Config struct{}

var ConfigDefaults = Config{}

// Collector is a Prometheus Collector for hyper-v.
type Collector struct {
	config Config

	collectorDynamicMemoryBalancer
	collectorDynamicMemoryVM
	collectorVirtualMachineHealthSummary
	collectorVirtualMachineVidPartition
	collectorVirtualStorageDevice
	collectorVirtualSwitch

	// Hyper-V Hypervisor Root Partition metrics
	addressSpaces                 *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Address Spaces
	attachedDevices               *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Attached Devices
	depositedPages                *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Deposited Pages
	deviceDMAErrors               *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Device DMA Errors
	deviceInterruptErrors         *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Device Interrupt Errors
	deviceInterruptMappings       *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Device Interrupt Mappings
	deviceInterruptThrottleEvents *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Device Interrupt Throttle Events
	gpaPages                      *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\GPA Pages
	gpaSpaceModifications         *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\GPA Space Modifications/sec
	ioTLBFlushCost                *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\I/O TLB Flush Cost
	ioTLBFlushes                  *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\I/O TLB Flushes/sec
	recommendedVirtualTLBSize     *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Recommended Virtual TLB Size
	skippedTimerTicks             *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Skipped Timer Ticks
	value1Gdevicepages            *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\1G device pages
	value1GGPApages               *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\1G GPA pages
	value2Mdevicepages            *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\2M device pages
	value2MGPApages               *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\2M GPA pages
	value4Kdevicepages            *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\4K device pages
	value4KGPApages               *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\4K GPA pages
	virtualTLBFlushEntires        *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Virtual TLB Flush Entires/sec
	virtualTLBPages               *prometheus.Desc // \Hyper-V Hypervisor Root Partition(*)\Virtual TLB Pages

	// Win32_PerfRawData_HvStats_HyperVHypervisor
	logicalProcessors *prometheus.Desc // \Hyper-V Hypervisor\Logical Processors
	virtualProcessors *prometheus.Desc // \Hyper-V Hypervisor\Virtual Processors

	// Win32_PerfRawData_HvStats_HyperVHypervisorLogicalProcessor
	hostLPGuestRunTimePercent      *prometheus.Desc // \Hyper-V Hypervisor Logical Processor(*)\% Guest Run Time
	hostLPHypervisorRunTimePercent *prometheus.Desc // \Hyper-V Hypervisor Logical Processor(*)\% Hypervisor Run Time
	hostLPTotalRunTimePercent      *prometheus.Desc // \Hyper-V Hypervisor Logical Processor(*)\% Total Run Time
	hostLPIdleRunTimePercent       *prometheus.Desc // \Hyper-V Hypervisor Logical Processor(*)\% Idle Time

	// Win32_PerfRawData_HvStats_HyperVHypervisorRootVirtualProcessor
	hostGuestRunTime           *prometheus.Desc // \Hyper-V Hypervisor Root Virtual Processor(*)\% Guest Run Time
	hostHypervisorRunTime      *prometheus.Desc // \Hyper-V Hypervisor Root Virtual Processor(*)\% Hypervisor Run Time
	hostRemoteRunTime          *prometheus.Desc // \Hyper-V Hypervisor Root Virtual Processor(*)\% Remote Run Time
	hostTotalRunTime           *prometheus.Desc // \Hyper-V Hypervisor Root Virtual Processor(*)\% Total Run Time
	hostCPUWaitTimePerDispatch *prometheus.Desc // \Hyper-V Hypervisor Root Virtual Processor(*)\CPU Wait Time Per Dispatch

	// Win32_PerfRawData_HvStats_HyperVHypervisorVirtualProcessor
	vmGuestRunTime           *prometheus.Desc // \Hyper-V Hypervisor Virtual Processor(*)\% Guest Run Time
	vmHypervisorRunTime      *prometheus.Desc // \Hyper-V Hypervisor Virtual Processor(*)\% Hypervisor Run Time
	vmRemoteRunTime          *prometheus.Desc // \Hyper-V Hypervisor Virtual Processor(*)\% Remote Run Time
	vmTotalRunTime           *prometheus.Desc // \Hyper-V Hypervisor Virtual Processor(*)\% Total Run Time
	vmCPUWaitTimePerDispatch *prometheus.Desc // \Hyper-V Hypervisor Virtual Processor(*)\CPU Wait Time Per Dispatch

	// Hyper-V Legacy Network Adapter metrics
	adapterBytesDropped   *prometheus.Desc // \Hyper-V Legacy Network Adapter(*)\Bytes Dropped
	adapterBytesReceived  *prometheus.Desc // \Hyper-V Legacy Network Adapter(*)\Bytes Received/sec
	adapterBytesSent      *prometheus.Desc // \Hyper-V Legacy Network Adapter(*)\Bytes Sent/sec
	adapterFramesDropped  *prometheus.Desc // \Hyper-V Legacy Network Adapter(*)\Frames Dropped
	adapterFramesReceived *prometheus.Desc // \Hyper-V Legacy Network Adapter(*)\Frames Received/sec
	adapterFramesSent     *prometheus.Desc // \Hyper-V Legacy Network Adapter(*)\Frames Sent/sec

	// Hyper-V Virtual Network Adapter metrics
	vmStorageBytesReceived          *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Bytes Received/sec
	vmStorageBytesSent              *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Bytes Sent/sec
	vmStorageDroppedPacketsIncoming *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Dropped Packets Incoming/sec
	vmStorageDroppedPacketsOutgoing *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Dropped Packets Outgoing/sec
	vmStoragePacketsReceived        *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Packets Received/sec
	vmStoragePacketsSent            *prometheus.Desc // \Hyper-V Virtual Network Adapter(*)\Packets Sent/sec

	// Hyper-V DataStore metrics
	// TODO: \Hyper-V DataStore(*)\Fragmentation ratio
	// TODO: \Hyper-V DataStore(*)\Sector size
	// TODO: \Hyper-V DataStore(*)\Data alignment
	// TODO: \Hyper-V DataStore(*)\Current replay logSize
	// TODO: \Hyper-V DataStore(*)\Number of available entries inside object tables
	// TODO: \Hyper-V DataStore(*)\Number of empty entries inside object tables
	// TODO: \Hyper-V DataStore(*)\Number of free bytes inside key tables
	// TODO: \Hyper-V DataStore(*)\Data end
	// TODO: \Hyper-V DataStore(*)\Number of file objects
	// TODO: \Hyper-V DataStore(*)\Number of object tables
	// TODO: \Hyper-V DataStore(*)\Number of key tables
	// TODO: \Hyper-V DataStore(*)\File data size in bytes
	// TODO: \Hyper-V DataStore(*)\Table data size in bytes
	// TODO: \Hyper-V DataStore(*)\Names size in bytes
	// TODO: \Hyper-V DataStore(*)\Number of keys
	// TODO: \Hyper-V DataStore(*)\Reconnect latency microseconds
	// TODO: \Hyper-V DataStore(*)\Disconnect count
	// TODO: \Hyper-V DataStore(*)\Write to file byte latency microseconds
	// TODO: \Hyper-V DataStore(*)\Write to file byte count
	// TODO: \Hyper-V DataStore(*)\Write to file count
	// TODO: \Hyper-V DataStore(*)\Read from file byte latency microseconds
	// TODO: \Hyper-V DataStore(*)\Read from file byte count
	// TODO: \Hyper-V DataStore(*)\Read from file count
	// TODO: \Hyper-V DataStore(*)\Write to storage byte latency microseconds
	// TODO: \Hyper-V DataStore(*)\Write to storage byte count
	// TODO: \Hyper-V DataStore(*)\Write to storage count
	// TODO: \Hyper-V DataStore(*)\Read from storage byte latency microseconds
	// TODO: \Hyper-V DataStore(*)\Read from storage byte count
	// TODO: \Hyper-V DataStore(*)\Read from storage count
	// TODO: \Hyper-V DataStore(*)\Commit byte latency microseconds
	// TODO: \Hyper-V DataStore(*)\Commit byte count
	// TODO: \Hyper-V DataStore(*)\Commit count
	// TODO: \Hyper-V DataStore(*)\Cache update operation latency microseconds
	// TODO: \Hyper-V DataStore(*)\Cache update operation count
	// TODO: \Hyper-V DataStore(*)\Commit operation latency microseconds
	// TODO: \Hyper-V DataStore(*)\Commit operation count
	// TODO: \Hyper-V DataStore(*)\Compact operation latency microseconds
	// TODO: \Hyper-V DataStore(*)\Compact operation count
	// TODO: \Hyper-V DataStore(*)\Load file operation latency microseconds
	// TODO: \Hyper-V DataStore(*)\Load file operation count
	// TODO: \Hyper-V DataStore(*)\Remove operation latency microseconds
	// TODO: \Hyper-V DataStore(*)\Remove operation count
	// TODO: \Hyper-V DataStore(*)\Query size operation latency microseconds
	// TODO: \Hyper-V DataStore(*)\Query size operation count
	// TODO: \Hyper-V DataStore(*)\Set operation latency microseconds
	// TODO: \Hyper-V DataStore(*)\Set operation count
	// TODO: \Hyper-V DataStore(*)\Get operation latency microseconds
	// TODO: \Hyper-V DataStore(*)\Get operation count
	// TODO: \Hyper-V DataStore(*)\File lock release latency microseconds
	// TODO: \Hyper-V DataStore(*)\File lock acquire latency microseconds
	// TODO: \Hyper-V DataStore(*)\File lock count
	// TODO: \Hyper-V DataStore(*)\Storage lock release latency microseconds
	// TODO: \Hyper-V DataStore(*)\Storage lock acquire latency microseconds
	// TODO: \Hyper-V DataStore(*)\Storage lock count

	// Hyper-V Virtual SMB metrics
	// TODO: \Hyper-V Virtual SMB(*)\Direct-Mapped Sections
	// TODO: \Hyper-V Virtual SMB(*)\Direct-Mapped Pages
	// TODO: \Hyper-V Virtual SMB(*)\Write Bytes/sec (RDMA)
	// TODO: \Hyper-V Virtual SMB(*)\Write Bytes/sec
	// TODO: \Hyper-V Virtual SMB(*)\Read Bytes/sec (RDMA)
	// TODO: \Hyper-V Virtual SMB(*)\Read Bytes/sec
	// TODO: \Hyper-V Virtual SMB(*)\Flush Requests/sec
	// TODO: \Hyper-V Virtual SMB(*)\Write Requests/sec (RDMA)
	// TODO: \Hyper-V Virtual SMB(*)\Write Requests/sec
	// TODO: \Hyper-V Virtual SMB(*)\Read Requests/sec (RDMA)
	// TODO: \Hyper-V Virtual SMB(*)\Read Requests/sec
	// TODO: \Hyper-V Virtual SMB(*)\Avg. sec/Request
	// TODO: \Hyper-V Virtual SMB(*)\Current Pending Requests
	// TODO: \Hyper-V Virtual SMB(*)\Current Open File Count
	// TODO: \Hyper-V Virtual SMB(*)\Tree Connect Count
	// TODO: \Hyper-V Virtual SMB(*)\Requests/sec
	// TODO: \Hyper-V Virtual SMB(*)\Sent Bytes/sec
	// TODO: \Hyper-V Virtual SMB(*)\Received Bytes/sec

}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	c.perfDataCollectorDynamicMemoryBalancer.Close()
	c.perfDataCollectorDynamicMemoryVM.Close()
	c.perfDataCollectorVirtualMachineHealthSummary.Close()
	c.perfDataCollectorVirtualStorageDevice.Close()
	c.perfDataCollectorVirtualSwitch.Close()
	c.perfDataCollectorVMVidPartition.Close()

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *wmi.Client) error {
	if err := c.buildDynamicMemoryBalancer(); err != nil {
		return err
	}

	if err := c.buildDynamicMemoryVM(); err != nil {
		return err
	}

	if err := c.buildVirtualStorageDevice(); err != nil {
		return err
	}

	if err := c.buildVirtualMachineHealthSummary(); err != nil {
		return err
	}

	if err := c.buildVirtualMachineVidPartition(); err != nil {
		return err
	}

	if err := c.buildVirtualSwitch(); err != nil {
		return err
	}

	c.addressSpaces = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_address_spaces"),
		"The number of address spaces in the virtual TLB of the partition",
		nil,
		nil,
	)
	c.attachedDevices = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_attached_devices"),
		"The number of devices attached to the partition",
		nil,
		nil,
	)
	c.depositedPages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_deposited_pages"),
		"The number of pages deposited into the partition",
		nil,
		nil,
	)
	c.deviceDMAErrors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_device_dma_errors"),
		"An indicator of illegal DMA requests generated by all devices assigned to the partition",
		nil,
		nil,
	)
	c.deviceInterruptErrors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_device_interrupt_errors"),
		"An indicator of illegal interrupt requests generated by all devices assigned to the partition",
		nil,
		nil,
	)
	c.deviceInterruptMappings = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_device_interrupt_mappings"),
		"The number of device interrupt mappings used by the partition",
		nil,
		nil,
	)
	c.deviceInterruptThrottleEvents = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_device_interrupt_throttle_events"),
		"The number of times an interrupt from a device assigned to the partition was temporarily throttled because the device was generating too many interrupts",
		nil,
		nil,
	)
	c.gpaPages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_preferred_numa_node_index"),
		"The number of pages present in the GPA space of the partition (zero for root partition)",
		nil,
		nil,
	)
	c.gpaSpaceModifications = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_gpa_space_modifications"),
		"The rate of modifications to the GPA space of the partition",
		nil,
		nil,
	)
	c.ioTLBFlushCost = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_io_tlb_flush_cost"),
		"The average time (in nanoseconds) spent processing an I/O TLB flush",
		nil,
		nil,
	)
	c.ioTLBFlushes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_io_tlb_flush"),
		"The rate of flushes of I/O TLBs of the partition",
		nil,
		nil,
	)
	c.recommendedVirtualTLBSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_recommended_virtual_tlb_size"),
		"The recommended number of pages to be deposited for the virtual TLB",
		nil,
		nil,
	)
	c.skippedTimerTicks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_physical_pages_allocated"),
		"The number of timer interrupts skipped for the partition",
		nil,
		nil,
	)
	c.value1Gdevicepages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_1G_device_pages"),
		"The number of 1G pages present in the device space of the partition",
		nil,
		nil,
	)
	c.value1GGPApages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_1G_gpa_pages"),
		"The number of 1G pages present in the GPA space of the partition",
		nil,
		nil,
	)
	c.value2Mdevicepages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_2M_device_pages"),
		"The number of 2M pages present in the device space of the partition",
		nil,
		nil,
	)
	c.value2MGPApages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_2M_gpa_pages"),
		"The number of 2M pages present in the GPA space of the partition",
		nil,
		nil,
	)
	c.value4Kdevicepages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_4K_device_pages"),
		"The number of 4K pages present in the device space of the partition",
		nil,
		nil,
	)
	c.value4KGPApages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_4K_gpa_pages"),
		"The number of 4K pages present in the GPA space of the partition",
		nil,
		nil,
	)
	c.virtualTLBFlushEntires = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_virtual_tlb_flush_entires"),
		"The rate of flushes of the entire virtual TLB",
		nil,
		nil,
	)
	c.virtualTLBPages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "root_partition_virtual_tlb_pages"),
		"The number of pages used by the virtual TLB of the partition",
		nil,
		nil,
	)

	//

	c.virtualProcessors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_virtual_processors"),
		"The number of virtual processors present in the system",
		nil,
		nil,
	)
	c.logicalProcessors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_logical_processors"),
		"The number of logical processors present in the system",
		nil,
		nil,
	)

	//

	c.hostLPGuestRunTimePercent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "host_lp_guest_run_time_percent"),
		"The percentage of time spent by the processor in guest code",
		[]string{"core"},
		nil,
	)
	c.hostLPHypervisorRunTimePercent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "host_lp_hypervisor_run_time_percent"),
		"The percentage of time spent by the processor in hypervisor code",
		[]string{"core"},
		nil,
	)
	c.hostLPTotalRunTimePercent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "host_lp_total_run_time_percent"),
		"The percentage of time spent by the processor in guest and hypervisor code",
		[]string{"core"},
		nil,
	)

	//

	c.hostGuestRunTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "host_cpu_guest_run_time"),
		"The time spent by the virtual processor in guest code",
		[]string{"core"},
		nil,
	)
	c.hostHypervisorRunTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "host_cpu_hypervisor_run_time"),
		"The time spent by the virtual processor in hypervisor code",
		[]string{"core"},
		nil,
	)
	c.hostRemoteRunTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "host_cpu_remote_run_time"),
		"The time spent by the virtual processor running on a remote node",
		[]string{"core"},
		nil,
	)
	c.hostTotalRunTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "host_cpu_total_run_time"),
		"The time spent by the virtual processor in guest and hypervisor code",
		[]string{"core"},
		nil,
	)
	c.hostCPUWaitTimePerDispatch = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "host_cpu_wait_time_per_dispatch_total"),
		"Time in nanoseconds waiting for a virtual processor to be dispatched onto a logical processor",
		[]string{"core"},
		nil,
	)

	//

	c.vmGuestRunTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vm_cpu_guest_run_time"),
		"The time spent by the virtual processor in guest code",
		[]string{"vm", "core"},
		nil,
	)
	c.vmHypervisorRunTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vm_cpu_hypervisor_run_time"),
		"The time spent by the virtual processor in hypervisor code",
		[]string{"vm", "core"},
		nil,
	)
	c.vmRemoteRunTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vm_cpu_remote_run_time"),
		"The time spent by the virtual processor running on a remote node",
		[]string{"vm", "core"},
		nil,
	)
	c.vmTotalRunTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vm_cpu_total_run_time"),
		"The time spent by the virtual processor in guest and hypervisor code",
		[]string{"vm", "core"},
		nil,
	)
	c.vmCPUWaitTimePerDispatch = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vm_cpu_wait_time_per_dispatch_total"),
		"Time in nanoseconds waiting for a virtual processor to be dispatched onto a logical processor",
		[]string{"vm", "core"},
		nil,
	)

	//

	c.adapterBytesDropped = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ethernet_bytes_dropped"),
		"Bytes Dropped is the number of bytes dropped on the network adapter",
		[]string{"adapter"},
		nil,
	)
	c.adapterBytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ethernet_bytes_received"),
		"Bytes received is the number of bytes received on the network adapter",
		[]string{"adapter"},
		nil,
	)
	c.adapterBytesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ethernet_bytes_sent"),
		"Bytes sent is the number of bytes sent over the network adapter",
		[]string{"adapter"},
		nil,
	)
	c.adapterFramesDropped = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ethernet_frames_dropped"),
		"Frames Dropped is the number of frames dropped on the network adapter",
		[]string{"adapter"},
		nil,
	)
	c.adapterFramesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ethernet_frames_received"),
		"Frames received is the number of frames received on the network adapter",
		[]string{"adapter"},
		nil,
	)
	c.adapterFramesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ethernet_frames_sent"),
		"Frames sent is the number of frames sent over the network adapter",
		[]string{"adapter"},
		nil,
	)

	//

	c.vmStorageBytesReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vm_interface_bytes_received"),
		"This counter represents the total number of bytes received per second by the network adapter",
		[]string{"vm_interface"},
		nil,
	)
	c.vmStorageBytesSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vm_interface_bytes_sent"),
		"This counter represents the total number of bytes sent per second by the network adapter",
		[]string{"vm_interface"},
		nil,
	)
	c.vmStorageDroppedPacketsIncoming = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vm_interface_packets_incoming_dropped"),
		"This counter represents the total number of dropped packets per second in the incoming direction of the network adapter",
		[]string{"vm_interface"},
		nil,
	)
	c.vmStorageDroppedPacketsOutgoing = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vm_interface_packets_outgoing_dropped"),
		"This counter represents the total number of dropped packets per second in the outgoing direction of the network adapter",
		[]string{"vm_interface"},
		nil,
	)
	c.vmStoragePacketsReceived = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vm_interface_packets_received"),
		"This counter represents the total number of packets received per second by the network adapter",
		[]string{"vm_interface"},
		nil,
	)
	c.vmStoragePacketsSent = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vm_interface_packets_sent"),
		"This counter represents the total number of packets sent per second by the network adapter",
		[]string{"vm_interface"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, _ *slog.Logger, ch chan<- prometheus.Metric) error {
	errs := make([]error, 0, 4)

	if err := c.collectDynamicMemoryBalancer(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Hyper-V Dynamic Memory Balancer metrics: %w", err))
	}

	if err := c.collectDynamicMemoryVM(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Hyper-V Dynamic Memory VM metrics: %w", err))
	}

	if err := c.collectVirtualMachineHealthSummary(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Hyper-V Virtual Machine Health Summary metrics: %w", err))
	}

	if err := c.collectVirtualMachineVidPartition(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Hyper-V VM Vid Partition metrics: %w", err))
	}

	if err := c.collectVirtualStorageDevice(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Hyper-V Virtual Storage Device metrics: %w", err))
	}

	if err := c.collectVirtualSwitch(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Hyper-V Virtual Storage Device metrics: %w", err))
	}

	return errors.Join(errs...)
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

func (c *Collector) collectVmHv(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_HvStats_HyperVHypervisorRootPartition
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_HvStats_HyperVHypervisorRootPartition", &dst); err != nil {
		return err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "*") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.addressSpaces,
			prometheus.GaugeValue,
			float64(obj.AddressSpaces),
		)

		ch <- prometheus.MustNewConstMetric(
			c.attachedDevices,
			prometheus.GaugeValue,
			float64(obj.AttachedDevices),
		)

		ch <- prometheus.MustNewConstMetric(
			c.depositedPages,
			prometheus.GaugeValue,
			float64(obj.DepositedPages),
		)

		ch <- prometheus.MustNewConstMetric(
			c.deviceDMAErrors,
			prometheus.GaugeValue,
			float64(obj.DeviceDMAErrors),
		)

		ch <- prometheus.MustNewConstMetric(
			c.deviceInterruptErrors,
			prometheus.GaugeValue,
			float64(obj.DeviceInterruptErrors),
		)

		ch <- prometheus.MustNewConstMetric(
			c.deviceInterruptThrottleEvents,
			prometheus.GaugeValue,
			float64(obj.DeviceInterruptThrottleEvents),
		)

		ch <- prometheus.MustNewConstMetric(
			c.gpaPages,
			prometheus.GaugeValue,
			float64(obj.GPAPages),
		)

		ch <- prometheus.MustNewConstMetric(
			c.gpaSpaceModifications,
			prometheus.CounterValue,
			float64(obj.GPASpaceModificationsPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioTLBFlushCost,
			prometheus.GaugeValue,
			float64(obj.IOTLBFlushCost),
		)

		ch <- prometheus.MustNewConstMetric(
			c.ioTLBFlushes,
			prometheus.CounterValue,
			float64(obj.IOTLBFlushesPersec),
		)

		ch <- prometheus.MustNewConstMetric(
			c.recommendedVirtualTLBSize,
			prometheus.GaugeValue,
			float64(obj.RecommendedVirtualTLBSize),
		)

		ch <- prometheus.MustNewConstMetric(
			c.skippedTimerTicks,
			prometheus.GaugeValue,
			float64(obj.SkippedTimerTicks),
		)

		ch <- prometheus.MustNewConstMetric(
			c.value1Gdevicepages,
			prometheus.GaugeValue,
			float64(obj.Value1Gdevicepages),
		)

		ch <- prometheus.MustNewConstMetric(
			c.value1GGPApages,
			prometheus.GaugeValue,
			float64(obj.Value1GGPApages),
		)

		ch <- prometheus.MustNewConstMetric(
			c.value2Mdevicepages,
			prometheus.GaugeValue,
			float64(obj.Value2Mdevicepages),
		)
		ch <- prometheus.MustNewConstMetric(
			c.value2MGPApages,
			prometheus.GaugeValue,
			float64(obj.Value2MGPApages),
		)
		ch <- prometheus.MustNewConstMetric(
			c.value4Kdevicepages,
			prometheus.GaugeValue,
			float64(obj.Value4Kdevicepages),
		)
		ch <- prometheus.MustNewConstMetric(
			c.value4KGPApages,
			prometheus.GaugeValue,
			float64(obj.Value4KGPApages),
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualTLBFlushEntires,
			prometheus.CounterValue,
			float64(obj.VirtualTLBFlushEntiresPersec),
		)
		ch <- prometheus.MustNewConstMetric(
			c.virtualTLBPages,
			prometheus.GaugeValue,
			float64(obj.VirtualTLBPages),
		)
	}

	return nil
}

// Win32_PerfRawData_HvStats_HyperVHypervisor ...
type Win32_PerfRawData_HvStats_HyperVHypervisor struct {
	LogicalProcessors uint64
	VirtualProcessors uint64
}

func (c *Collector) collectVmProcessor(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_HvStats_HyperVHypervisor
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_HvStats_HyperVHypervisor", &dst); err != nil {
		return err
	}

	for _, obj := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.logicalProcessors,
			prometheus.GaugeValue,
			float64(obj.LogicalProcessors),
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualProcessors,
			prometheus.GaugeValue,
			float64(obj.VirtualProcessors),
		)
	}

	return nil
}

// Win32_PerfRawData_HvStats_HyperVHypervisorLogicalProcessor ...
type Win32_PerfRawData_HvStats_HyperVHypervisorLogicalProcessor struct {
	Name                     string
	PercentGuestRunTime      uint64
	PercentHypervisorRunTime uint64
	PercentTotalRunTime      uint
}

func (c *Collector) collectHostLPUsage(logger *slog.Logger, ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_HvStats_HyperVHypervisorLogicalProcessor
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_HvStats_HyperVHypervisorLogicalProcessor", &dst); err != nil {
		return err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "*") {
			continue
		}

		// The name format is Hv LP <core id>
		parts := strings.Split(obj.Name, " ")
		if len(parts) != 3 {
			logger.Warn(fmt.Sprintf("Unexpected format of Name in collectHostLPUsage: %q", obj.Name))

			continue
		}

		coreId := parts[2]

		ch <- prometheus.MustNewConstMetric(
			c.hostLPGuestRunTimePercent,
			prometheus.GaugeValue,
			float64(obj.PercentGuestRunTime),
			coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.hostLPHypervisorRunTimePercent,
			prometheus.GaugeValue,
			float64(obj.PercentHypervisorRunTime),
			coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.hostLPTotalRunTimePercent,
			prometheus.GaugeValue,
			float64(obj.PercentTotalRunTime),
			coreId,
		)
	}

	return nil
}

// Win32_PerfRawData_HvStats_HyperVHypervisorRootVirtualProcessor ...
type Win32_PerfRawData_HvStats_HyperVHypervisorRootVirtualProcessor struct {
	Name                     string
	PercentGuestRunTime      uint64
	PercentHypervisorRunTime uint64
	PercentRemoteRunTime     uint64
	PercentTotalRunTime      uint64
	CPUWaitTimePerDispatch   uint64
}

func (c *Collector) collectHostCpuUsage(logger *slog.Logger, ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_HvStats_HyperVHypervisorRootVirtualProcessor
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_HvStats_HyperVHypervisorRootVirtualProcessor", &dst); err != nil {
		return err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "*") {
			continue
		}

		// The name format is Root VP <core id>
		parts := strings.Split(obj.Name, " ")
		if len(parts) != 3 {
			logger.Warn("Unexpected format of Name in collectHostCpuUsage: " + obj.Name)

			continue
		}

		coreId := parts[2]

		ch <- prometheus.MustNewConstMetric(
			c.hostGuestRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentGuestRunTime),
			coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.hostHypervisorRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentHypervisorRunTime),
			coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.hostRemoteRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentRemoteRunTime),
			coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.hostTotalRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentTotalRunTime),
			coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.hostCPUWaitTimePerDispatch,
			prometheus.CounterValue,
			float64(obj.CPUWaitTimePerDispatch),
			coreId,
		)
	}

	return nil
}

// Win32_PerfRawData_HvStats_HyperVHypervisorVirtualProcessor ...
type Win32_PerfRawData_HvStats_HyperVHypervisorVirtualProcessor struct {
	Name                     string
	PercentGuestRunTime      uint64
	PercentHypervisorRunTime uint64
	PercentRemoteRunTime     uint64
	PercentTotalRunTime      uint64
	CPUWaitTimePerDispatch   uint64
}

func (c *Collector) collectVmCpuUsage(logger *slog.Logger, ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_HvStats_HyperVHypervisorVirtualProcessor
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_HvStats_HyperVHypervisorVirtualProcessor", &dst); err != nil {
		return err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "*") {
			continue
		}

		// The name format is <VM Name>:Hv VP <vcore id>
		parts := strings.Split(obj.Name, ":")
		if len(parts) != 2 {
			logger.Warn(fmt.Sprintf("Unexpected format of Name in collectVmCpuUsage: %q, expected %q. Skipping.", obj.Name, "<VM Name>:Hv VP <vcore id>"))

			continue
		}

		coreParts := strings.Split(parts[1], " ")
		if len(coreParts) != 3 {
			logger.Warn(fmt.Sprintf("Unexpected format of core identifier in collectVmCpuUsage: %q, expected %q. Skipping.", parts[1], "Hv VP <vcore id>"))

			continue
		}

		vmName := parts[0]
		coreId := coreParts[2]

		ch <- prometheus.MustNewConstMetric(
			c.vmGuestRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentGuestRunTime),
			vmName, coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmHypervisorRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentHypervisorRunTime),
			vmName, coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmRemoteRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentRemoteRunTime),
			vmName, coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmTotalRunTime,
			prometheus.GaugeValue,
			float64(obj.PercentTotalRunTime),
			vmName, coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmCPUWaitTimePerDispatch,
			prometheus.CounterValue,
			float64(obj.CPUWaitTimePerDispatch),
			vmName, coreId,
		)
	}

	return nil
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

func (c *Collector) collectVmEthernet(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_EthernetPerfProvider_HyperVLegacyNetworkAdapter
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_EthernetPerfProvider_HyperVLegacyNetworkAdapter", &dst); err != nil {
		return err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "*") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.adapterBytesDropped,
			prometheus.GaugeValue,
			float64(obj.BytesDropped),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.adapterBytesReceived,
			prometheus.CounterValue,
			float64(obj.BytesReceivedPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.adapterBytesSent,
			prometheus.CounterValue,
			float64(obj.BytesSentPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.adapterFramesReceived,
			prometheus.CounterValue,
			float64(obj.FramesReceivedPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.adapterFramesDropped,
			prometheus.CounterValue,
			float64(obj.FramesDropped),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.adapterFramesSent,
			prometheus.CounterValue,
			float64(obj.FramesSentPersec),
			obj.Name,
		)
	}

	return nil
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

func (c *Collector) collectVmStorage(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_Counters_HyperVVirtualStorageDevice
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_Counters_HyperVVirtualStorageDevice", &dst); err != nil {
		return err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "*") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.vmStorageErrorCount,
			prometheus.CounterValue,
			float64(obj.ErrorCount),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmStorageQueueLength,
			prometheus.CounterValue,
			float64(obj.QueueLength),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmStorageReadBytes,
			prometheus.CounterValue,
			float64(obj.ReadBytesPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmStorageReadOperations,
			prometheus.CounterValue,
			float64(obj.ReadOperationsPerSec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmStorageWriteBytes,
			prometheus.CounterValue,
			float64(obj.WriteBytesPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmStorageWriteOperations,
			prometheus.CounterValue,
			float64(obj.WriteOperationsPerSec),
			obj.Name,
		)
	}

	return nil
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

func (c *Collector) collectVmNetwork(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NvspNicStats_HyperVVirtualNetworkAdapter
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_NvspNicStats_HyperVVirtualNetworkAdapter", &dst); err != nil {
		return err
	}

	for _, obj := range dst {
		if strings.Contains(obj.Name, "*") {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.vmStorageBytesReceived,
			prometheus.CounterValue,
			float64(obj.BytesReceivedPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmStorageBytesSent,
			prometheus.CounterValue,
			float64(obj.BytesSentPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmStorageDroppedPacketsIncoming,
			prometheus.CounterValue,
			float64(obj.DroppedPacketsIncomingPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmStorageDroppedPacketsOutgoing,
			prometheus.CounterValue,
			float64(obj.DroppedPacketsOutgoingPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmStoragePacketsReceived,
			prometheus.CounterValue,
			float64(obj.PacketsReceivedPersec),
			obj.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmStoragePacketsSent,
			prometheus.CounterValue,
			float64(obj.PacketsSentPersec),
			obj.Name,
		)
	}

	return nil
}
