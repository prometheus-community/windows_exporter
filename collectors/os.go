// returns data points from Win32_OperatingSystem
// https://msdn.microsoft.com/en-us/library/aa394239 - Win32_OperatingSystem class

package collectors

import (
	"log"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

// A OSCollector is a Prometheus collector for WMI Win32_OperatingSystem metrics
type OSCollector struct {
	PhysicalMemoryFreeBytes *prometheus.Desc
	PagingFreeBytes         *prometheus.Desc
	VirtualMemoryFreeBytes  *prometheus.Desc
	ProcessesMax            *prometheus.Desc
	ProcessMemoryMaxBytes   *prometheus.Desc
	Processes               *prometheus.Desc
	Users                   *prometheus.Desc
	PagingMaxBytes          *prometheus.Desc
	VirtualMemoryBytes      *prometheus.Desc
	VisibleMemoryBytes      *prometheus.Desc
}

// NewOSCollector ...
func NewOSCollector() *OSCollector {

	return &OSCollector{

		PagingMaxBytes: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "paging_max_bytes"),
			"SizeStoredInPagingFiles",
			nil,
			nil,
		),

		PagingFreeBytes: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "paging_free_bytes"),
			"FreeSpaceInPagingFiles",
			nil,
			nil,
		),

		PhysicalMemoryFreeBytes: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "physical_memory_free_bytes"),
			"FreePhysicalMemory",
			nil,
			nil,
		),

		Processes: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "processes"),
			"NumberOfProcesses",
			nil,
			nil,
		),

		ProcessesMax: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "processes_max"),
			"MaxNumberOfProcesses",
			nil,
			nil,
		),

		ProcessMemoryMaxBytes: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "process_memory_max_bytes"),
			"MaxProcessMemorySize",
			nil,
			nil,
		),

		Users: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "users"),
			"NumberOfUsers",
			nil,
			nil,
		),

		VirtualMemoryBytes: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "virtual_memory_bytes"),
			"TotalVirtualMemorySize",
			nil,
			nil,
		),

		VisibleMemoryBytes: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "visible_memory_bytes"),
			"TotalVisibleMemorySize",
			nil,
			nil,
		),

		VirtualMemoryFreeBytes: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "virtual_memory_free_bytes"),
			"FreeVirtualMemory",
			nil,
			nil,
		),
	}
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *OSCollector) Collect(ch chan<- prometheus.Metric) {
	if desc, err := c.collect(ch); err != nil {
		log.Println("[ERROR] failed collecting os metrics:", desc, err)
		return
	}
}

// Describe sends the descriptors of each metric over to the provided channel.
// The corresponding metric values are sent separately.
func (c *OSCollector) Describe(ch chan<- *prometheus.Desc) {

	ch <- c.PhysicalMemoryFreeBytes
	ch <- c.PagingFreeBytes
	ch <- c.VirtualMemoryFreeBytes
	ch <- c.ProcessesMax
	ch <- c.ProcessMemoryMaxBytes
	ch <- c.Processes
	ch <- c.Users
	ch <- c.PagingMaxBytes
	ch <- c.VirtualMemoryBytes
	ch <- c.VisibleMemoryBytes
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
}

func (c *OSCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_OperatingSystem
	q := wmi.CreateQuery(&dst, "")
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.PhysicalMemoryFreeBytes,
		prometheus.GaugeValue,
		float64(dst[0].FreePhysicalMemory*1024), // KiB -> bytes
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
		c.ProcessesMax,
		prometheus.GaugeValue,
		float64(dst[0].MaxNumberOfProcesses),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ProcessMemoryMaxBytes,
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
		c.PagingMaxBytes,
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
