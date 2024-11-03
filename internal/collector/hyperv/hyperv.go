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
	collectorHypervisorLogicalProcessor
	collectorHypervisorRootPartition
	collectorVirtualMachineHealthSummary
	collectorVirtualMachineVidPartition
	collectorVirtualNetworkAdapter
	collectorVirtualStorageDevice
	collectorVirtualSwitch

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

	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing LowPowerPacketFilter
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming LowPowerPacketFilter
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InvalidPDQueue
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InvalidPDQueue
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing FilteredIsolationUntagged
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming FilteredIsolationUntagged
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing SwitchDataFlowDisabled
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming SwitchDataFlowDisabled
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing FailedPacketFilter
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming FailedPacketFilter
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing NicDisabled
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming NicDisabled
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing FailedDestinationListUpdate
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming FailedDestinationListUpdate
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InjectedIcmp
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InjectedIcmp
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing StormLimit
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming StormLimit
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InvalidFirstNBTooSmall
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InvalidFirstNBTooSmall
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InvalidSourceMac
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InvalidSourceMac
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InvalidDestMac
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InvalidDestMac
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InvalidVlanFormat
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InvalidVlanFormat
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing NativeFwdingReq
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming NativeFwdingReq
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing MTUMismatch
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming MTUMismatch
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InvalidConfig
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InvalidConfig
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing RequiredExtensionMissing
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming RequiredExtensionMissing
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing VirtualSubnetId
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming VirtualSubnetId
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing BridgeReserved
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming BridgeReserved
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing RouterGuard
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming RouterGuard
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing DhcpGuard
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming DhcpGuard
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing MacSpoofing
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming MacSpoofing
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing Ipsec
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming Ipsec
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing Qos
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming Qos
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing FailedPvlanSetting
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming FailedPvlanSetting
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing FailedSecurityPolicy
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming FailedSecurityPolicy
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing UnauthorizedMAC
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming UnauthorizedMAC
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing UnauthorizedVLAN
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming UnauthorizedVLAN
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing FilteredVLAN
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming FilteredVLAN
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing Filtered
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming Filtered
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing Busy
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming Busy
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing NotAccepted
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming NotAccepted
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing Disconnected
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming Disconnected
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing NotReady
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming NotReady
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing Resources
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming Resources
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InvalidPacket
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InvalidPacket
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing InvalidData
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming InvalidData
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Outgoing Unknown
	// TODO: \Hyper-V Virtual Network Adapter Drop Reasons(*)\Incoming Unknown

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
	c.perfDataCollectorHypervisorLogicalProcessor.Close()
	c.perfDataCollectorHypervisorRootPartition.Close()
	c.perfDataCollectorVirtualMachineHealthSummary.Close()
	c.perfDataCollectorVirtualMachineVidPartition.Close()
	c.perfDataCollectorVirtualNetworkAdapter.Close()
	c.perfDataCollectorVirtualStorageDevice.Close()
	c.perfDataCollectorVirtualSwitch.Close()

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *wmi.Client) error {
	if err := c.buildDynamicMemoryBalancer(); err != nil {
		return err
	}

	if err := c.buildDynamicMemoryVM(); err != nil {
		return err
	}

	if err := c.buildHypervisorLogicalProcessor(); err != nil {
		return err
	}

	if err := c.buildHypervisorRootPartition(); err != nil {
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

	if err := c.buildVirtualNetworkAdapter(); err != nil {
		return err
	}

	if err := c.buildVirtualSwitch(); err != nil {
		return err
	}

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

	if err := c.collectHypervisorLogicalProcessor(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Hyper-V Hypervisor Logical Processor metrics: %w", err))
	}

	if err := c.collectHypervisorRootPartition(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Hyper-V Hypervisor Root Partition metrics: %w", err))
	}

	if err := c.collectVirtualMachineHealthSummary(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Hyper-V Virtual Machine Health Summary metrics: %w", err))
	}

	if err := c.collectVirtualMachineVidPartition(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Hyper-V VM Vid Partition metrics: %w", err))
	}

	if err := c.collectVirtualNetworkAdapter(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Hyper-V Virtual Network Adapter metrics: %w", err))
	}

	if err := c.collectVirtualStorageDevice(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Hyper-V Virtual Storage Device metrics: %w", err))
	}

	if err := c.collectVirtualSwitch(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting Hyper-V Virtual Switch metrics: %w", err))
	}

	return errors.Join(errs...)
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
