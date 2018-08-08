// returns data points from Win32_OperatingSystem
// https://msdn.microsoft.com/en-us/library/aa394239 - Win32_OperatingSystem class

package collector

import (
	"errors"
	"time"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["os"] = NewOSCollector
}

// A OSCollector is a Prometheus collector for WMI metrics
type OSCollector struct {
	PhysicalMemoryFreeBytes *prometheus.Desc
	PagingFreeBytes         *prometheus.Desc
	VirtualMemoryFreeBytes  *prometheus.Desc
	ProcessesLimit          *prometheus.Desc
	ProcessMemoryLimitBytes *prometheus.Desc
	Processes               *prometheus.Desc
	Users                   *prometheus.Desc
	PagingLimitBytes        *prometheus.Desc
	VirtualMemoryBytes      *prometheus.Desc
	VisibleMemoryBytes      *prometheus.Desc
	Time                    *prometheus.Desc
	Timezone                *prometheus.Desc
}

// NewOSCollector ...
func NewOSCollector() (Collector, error) {
	const subsystem = "os"

	return &OSCollector{
		PagingLimitBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "paging_limit_bytes"),
			"OperatingSystem.SizeStoredInPagingFiles",
			nil,
			nil,
		),
		PagingFreeBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "paging_free_bytes"),
			"OperatingSystem.FreeSpaceInPagingFiles",
			nil,
			nil,
		),
		PhysicalMemoryFreeBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "physical_memory_free_bytes"),
			"OperatingSystem.FreePhysicalMemory",
			nil,
			nil,
		),
		Time: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "time"),
			"OperatingSystem.LocalDateTime",
			nil,
			nil,
		),
		Timezone: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "timezone"),
			"OperatingSystem.LocalDateTime",
			[]string{"timezone"},
			nil,
		),
		Processes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "processes"),
			"OperatingSystem.NumberOfProcesses",
			nil,
			nil,
		),
		ProcessesLimit: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "processes_limit"),
			"OperatingSystem.MaxNumberOfProcesses",
			nil,
			nil,
		),
		ProcessMemoryLimitBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "process_memory_limix_bytes"),
			"OperatingSystem.MaxProcessMemorySize",
			nil,
			nil,
		),
		Users: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "users"),
			"OperatingSystem.NumberOfUsers",
			nil,
			nil,
		),
		VirtualMemoryBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "virtual_memory_bytes"),
			"OperatingSystem.TotalVirtualMemorySize",
			nil,
			nil,
		),
		VisibleMemoryBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "visible_memory_bytes"),
			"OperatingSystem.TotalVisibleMemorySize",
			nil,
			nil,
		),
		VirtualMemoryFreeBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "virtual_memory_free_bytes"),
			"OperatingSystem.FreeVirtualMemory",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *OSCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting os metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_OperatingSystem struct {
	FreePhysicalMemory      uint64
	FreeSpaceInPagingFiles  uint64
	FreeVirtualMemory       uint64
	MaxNumberOfProcesses    uint32
	MaxProcessMemorySize    uint64
	NumberOfProcesses       uint32
	NumberOfUsers           uint32
	SizeStoredInPagingFiles uint64
	TotalVirtualMemorySize  uint64
	TotalVisibleMemorySize  uint64
	LocalDateTime           time.Time
}

func (c *OSCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_OperatingSystem
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.PhysicalMemoryFreeBytes,
		prometheus.GaugeValue,
		float64(dst[0].FreePhysicalMemory*1024), // KiB -> bytes
	)

	time := dst[0].LocalDateTime

	ch <- prometheus.MustNewConstMetric(
		c.Time,
		prometheus.GaugeValue,
		float64(time.Unix()),
	)

	timezoneName, _ := time.Zone()

	ch <- prometheus.MustNewConstMetric(
		c.Timezone,
		prometheus.GaugeValue,
		1.0,
		timezoneName,
	)

	ch <- prometheus.MustNewConstMetric(
		c.PagingFreeBytes,
		prometheus.GaugeValue,
		float64(dst[0].FreeSpaceInPagingFiles*1024), // KiB -> bytes
	)

	ch <- prometheus.MustNewConstMetric(
		c.VirtualMemoryFreeBytes,
		prometheus.GaugeValue,
		float64(dst[0].FreeVirtualMemory*1024), // KiB -> bytes
	)

	ch <- prometheus.MustNewConstMetric(
		c.ProcessesLimit,
		prometheus.GaugeValue,
		float64(dst[0].MaxNumberOfProcesses),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ProcessMemoryLimitBytes,
		prometheus.GaugeValue,
		float64(dst[0].MaxProcessMemorySize*1024), // KiB -> bytes
	)

	ch <- prometheus.MustNewConstMetric(
		c.Processes,
		prometheus.GaugeValue,
		float64(dst[0].NumberOfProcesses),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Users,
		prometheus.GaugeValue,
		float64(dst[0].NumberOfUsers),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PagingLimitBytes,
		prometheus.GaugeValue,
		float64(dst[0].SizeStoredInPagingFiles*1024), // KiB -> bytes
	)

	ch <- prometheus.MustNewConstMetric(
		c.VirtualMemoryBytes,
		prometheus.GaugeValue,
		float64(dst[0].TotalVirtualMemorySize*1024), // KiB -> bytes
	)

	ch <- prometheus.MustNewConstMetric(
		c.VisibleMemoryBytes,
		prometheus.GaugeValue,
		float64(dst[0].TotalVisibleMemorySize*1024), // KiB -> bytes
	)

	return nil, nil
}
