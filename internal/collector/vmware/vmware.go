//go:build windows

package vmware

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "vmware"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_vmGuestLib_VMem/Win32_PerfRawData_vmGuestLib_VCPU metrics.
type Collector struct {
	config                  Config
	perfDataCollectorCPU    *perfdata.Collector
	perfDataCollectorMemory *perfdata.Collector

	memActive      *prometheus.Desc
	memBallooned   *prometheus.Desc
	memLimit       *prometheus.Desc
	memMapped      *prometheus.Desc
	memOverhead    *prometheus.Desc
	memReservation *prometheus.Desc
	memShared      *prometheus.Desc
	memSharedSaved *prometheus.Desc
	memShares      *prometheus.Desc
	memSwapped     *prometheus.Desc
	memTargetSize  *prometheus.Desc
	memUsed        *prometheus.Desc

	cpuLimitMHz            *prometheus.Desc
	cpuReservationMHz      *prometheus.Desc
	cpuShares              *prometheus.Desc
	cpuStolenTotal         *prometheus.Desc
	cpuTimeTotal           *prometheus.Desc
	cpuEffectiveVMSpeedMHz *prometheus.Desc
	hostProcessorSpeedMHz  *prometheus.Desc
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

func (c *Collector) Close() error {
	c.perfDataCollectorCPU.Close()
	c.perfDataCollectorMemory.Close()

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	var err error

	c.perfDataCollectorCPU, err = perfdata.NewCollector("VM Processor", perfdata.InstanceTotal, []string{
		cpuLimitMHz,
		cpuReservationMHz,
		cpuShares,
		cpuStolenMs,
		cpuTimePercents,
		couEffectiveVMSpeedMHz,
		cpuHostProcessorSpeedMHz,
	})
	if err != nil {
		return fmt.Errorf("failed to create VM Processor collector: %w", err)
	}

	c.cpuLimitMHz = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_limit_mhz"),
		"The maximum processing power in MHz allowed to the virtual machine. Assigning a CPU Limit ensures that this virtual machine never consumes more than a certain amount of the available processor power. By limiting the amount of processing power consumed, a portion of the processing power becomes available to other virtual machines.",
		nil,
		nil,
	)
	c.cpuReservationMHz = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_reservation_mhz"),
		"The minimum processing power in MHz available to the virtual machine. Assigning a CPU Reservation ensures that even as other virtual machines on the same host consume shared processing power, there is still a certain minimum amount for this virtual machine.",
		nil,
		nil,
	)
	c.cpuShares = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_shares"),
		"The number of CPU shares allocated to the virtual machine.",
		nil,
		nil,
	)
	c.cpuStolenTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_stolen_seconds_total"),
		"The time that the VM was runnable but not scheduled to run.",
		nil,
		nil,
	)
	c.cpuTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_time_seconds_total"),
		"Current load of the VM’s virtual processor",
		nil,
		nil,
	)
	c.cpuEffectiveVMSpeedMHz = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cpu_effective_vm_speed_mhz_total"),
		"The effective speed of the VM’s virtual CPU",
		nil,
		nil,
	)
	c.hostProcessorSpeedMHz = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "host_processor_speed_mhz"),
		"Host Processor speed",
		nil,
		nil,
	)

	c.perfDataCollectorMemory, err = perfdata.NewCollector("VM Memory", nil, []string{
		memActiveMB,
		memBalloonedMB,
		memLimitMB,
		memMappedMB,
		memOverheadMB,
		memReservationMB,
		memSharedMB,
		memSharedSavedMB,
		memShares,
		memSwappedMB,
		memTargetSizeMB,
		memUsedMB,
	})
	if err != nil {
		return fmt.Errorf("failed to create VM Memory collector: %w", err)
	}

	c.memActive = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_active_bytes"),
		"The estimated amount of memory the virtual machine is actively using.",
		nil,
		nil,
	)
	c.memBallooned = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_ballooned_bytes"),
		"The amount of memory that has been reclaimed from this virtual machine via the VMware Memory Balloon mechanism.",
		nil,
		nil,
	)
	c.memLimit = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_limit_bytes"),
		"The maximum amount of memory that is allowed to the virtual machine. Assigning a Memory Limit ensures that this virtual machine never consumes more than a certain amount of the allowed memory. By limiting the amount of memory consumed, a portion of this shared resource is allowed to other virtual machines.",
		nil,
		nil,
	)
	c.memMapped = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_mapped_bytes"),
		"The mapped memory size of this virtual machine. This is the current total amount of guest memory that is backed by physical memory. Note that this number may include pages of memory shared between multiple virtual machines and thus may be an overestimate of the amount of physical host memory consumed by this virtual machine.",
		nil,
		nil,
	)
	c.memOverhead = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_overhead_bytes"),
		"The amount of overhead memory associated with this virtual machine consumed on the host system.",
		nil,
		nil,
	)
	c.memReservation = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_reservation_bytes"),
		"The minimum amount of memory that is guaranteed to the virtual machine. Assigning a Memory Reservation ensures that even as other virtual machines on the same host consume memory, there is still a certain minimum amount for this virtual machine.",
		nil,
		nil,
	)
	c.memShared = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_shared_bytes"),
		"The amount of physical memory associated with this virtual machine that is copy-on-write (COW) shared on the host.",
		nil,
		nil,
	)
	c.memSharedSaved = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_shared_saved_bytes"),
		"The estimated amount of physical memory on the host saved from copy-on-write (COW) shared guest physical memory.",
		nil,
		nil,
	)
	c.memShares = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_shares"),
		"The number of memory shares allocated to the virtual machine.",
		nil,
		nil,
	)
	c.memSwapped = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_swapped_bytes"),
		"The amount of memory associated with this virtual machine that has been swapped by ESX.",
		nil,
		nil,
	)
	c.memTargetSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_target_size_bytes"),
		"Memory Target Size.",
		nil,
		nil,
	)
	c.memUsed = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mem_used_bytes"),
		"The estimated amount of physical host memory currently consumed for this virtual machine’s physical memory.",
		nil,
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0, 2)

	if err := c.collectCpu(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting vmware cpu metrics: %w", err))
	}

	if err := c.collectMem(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed collecting vmware memory metrics: %w", err))
	}

	return errors.Join(errs...)
}

func (c *Collector) collectMem(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorMemory.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect VM Memory metrics: %w", err)
	}

	data, ok := perfData[perfdata.EmptyInstance]
	if !ok {
		return errors.New("query for VM Memory returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.memActive,
		prometheus.GaugeValue,
		utils.MBToBytes(data[memActiveMB].FirstValue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memBallooned,
		prometheus.GaugeValue,
		utils.MBToBytes(data[memBalloonedMB].FirstValue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memLimit,
		prometheus.GaugeValue,
		utils.MBToBytes(data[memLimitMB].FirstValue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memMapped,
		prometheus.GaugeValue,
		utils.MBToBytes(data[memMappedMB].FirstValue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memOverhead,
		prometheus.GaugeValue,
		utils.MBToBytes(data[memOverheadMB].FirstValue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memReservation,
		prometheus.GaugeValue,
		utils.MBToBytes(data[memReservationMB].FirstValue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memShared,
		prometheus.GaugeValue,
		utils.MBToBytes(data[memSharedMB].FirstValue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memSharedSaved,
		prometheus.GaugeValue,
		utils.MBToBytes(data[memSharedSavedMB].FirstValue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memShares,
		prometheus.GaugeValue,
		data[memShares].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.memSwapped,
		prometheus.GaugeValue,
		utils.MBToBytes(data[memSwappedMB].FirstValue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memTargetSize,
		prometheus.GaugeValue,
		utils.MBToBytes(data[memTargetSizeMB].FirstValue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.memUsed,
		prometheus.GaugeValue,
		utils.MBToBytes(data[memUsedMB].FirstValue),
	)

	return nil
}

func (c *Collector) collectCpu(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorCPU.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect VM Memory metrics: %w", err)
	}

	data, ok := perfData["_Total"]
	if !ok {
		return errors.New("query for VM CPU returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.cpuLimitMHz,
		prometheus.GaugeValue,
		data[cpuLimitMHz].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.cpuReservationMHz,
		prometheus.GaugeValue,
		data[cpuReservationMHz].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.cpuShares,
		prometheus.GaugeValue,
		data[cpuShares].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.cpuStolenTotal,
		prometheus.CounterValue,
		utils.MilliSecToSec(data[cpuStolenMs].FirstValue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.cpuTimeTotal,
		prometheus.CounterValue,
		utils.MilliSecToSec(data[cpuTimePercents].FirstValue),
	)

	ch <- prometheus.MustNewConstMetric(
		c.cpuEffectiveVMSpeedMHz,
		prometheus.GaugeValue,
		data[couEffectiveVMSpeedMHz].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.hostProcessorSpeedMHz,
		prometheus.GaugeValue,
		data[cpuHostProcessorSpeedMHz].FirstValue,
	)

	return nil
}
