// returns data points from Win32_OperatingSystem
// https://msdn.microsoft.com/en-us/library/aa394239(v=vs.85).aspx - Win32_OperatingSystem class
// https://msdn.microsoft.com/en-us/library/aa387937(v=vs.85).aspx - CIM_OperatingSystem class

package collectors

import (
	"log"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

// A OSCollector is a Prometheus collector for WMI Win32_OperatingSystem metrics
type OSCollector struct {
	FreePhysicalMemoryBytes      *prometheus.Desc
	FreeSpaceInPagingFilesBytes  *prometheus.Desc
	FreeVirtualMemoryBytes       *prometheus.Desc
	ProcessesMax                 *prometheus.Desc
	ProcessMemoryBytesBytes      *prometheus.Desc
	Processes                    *prometheus.Desc
	Users                        *prometheus.Desc
	SizeStoredInPagingFilesBytes *prometheus.Desc
	VirtualMemoryBytesTotal      *prometheus.Desc
	VisibleMemoryBytesTotal      *prometheus.Desc
}

// NewOSCollector ...
func NewOSCollector() *OSCollector {

	return &OSCollector{
		FreePhysicalMemoryBytes: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "free_physical_memory_bytes"),
			"Physical memory currently unused and available.",
			nil,
			nil,
		),

		FreeSpaceInPagingFilesBytes: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "free_space_in_paging_files_bytes"),
			"Number of bytes that can be mapped into the operating system paging files without causing any other pages to be swapped out.",
			nil,
			nil,
		),

		FreeVirtualMemoryBytes: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "free_virtual_memory_bytes"),
			"Virtual memory currently unused and available.",
			nil,
			nil,
		),

		ProcessesMax: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "processes_max"),
			"Maximum number of process contexts the operating system can support.",
			nil,
			nil,
		),

		ProcessMemoryBytesBytes: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "process_memory_bytes_max"),
			"Maximum bytes of memory that can be allocated to a process.",
			nil,
			nil,
		),

		Processes: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "processes"),
			"Number of process contexts currently loaded or running on the operating system.",
			nil,
			nil,
		),

		Users: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "users"),
			"Number of user sessions for which the operating system is storing state information currently.",
			nil,
			nil,
		),

		SizeStoredInPagingFilesBytes: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "size_stored_in_paging_files_bytes"),
			"Total number of bytes that can be stored in the operating system paging files.",
			nil,
			nil,
		),

		VirtualMemoryBytesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "virtual_memory_bytes_total"),
			"Total amount of virtual memory.",
			nil,
			nil,
		),

		VisibleMemoryBytesTotal: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "visible_memory_bytes_total"),
			"Total amount of physical memory available to the operating system.",
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

	ch <- c.FreePhysicalMemoryBytes
	ch <- c.FreeSpaceInPagingFilesBytes
	ch <- c.FreeVirtualMemoryBytes
	ch <- c.ProcessesMax
	ch <- c.ProcessMemoryBytesBytes
	ch <- c.Processes
	ch <- c.Users
	ch <- c.SizeStoredInPagingFilesBytes
	ch <- c.VirtualMemoryBytesTotal
	ch <- c.VisibleMemoryBytesTotal
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
	//NumberOfLicensedUsers   *uint32 // XXX returns 0 on Win7
	//TotalSwapSpaceSize      *uint64 // XXX returns 0 on Win7
}

func (c *OSCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_OperatingSystem
	q := wmi.CreateQuery(&dst, "")
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.FreePhysicalMemoryBytes,
		prometheus.GaugeValue,
		float64(dst[0].FreePhysicalMemory*1024), // KiB -> bytes
	)

	ch <- prometheus.MustNewConstMetric(
		c.FreeSpaceInPagingFilesBytes,
		prometheus.GaugeValue,
		float64(dst[0].FreeSpaceInPagingFiles*1024), // KiB -> bytes
	)

	ch <- prometheus.MustNewConstMetric(
		c.FreeVirtualMemoryBytes,
		prometheus.GaugeValue,
		float64(dst[0].FreeVirtualMemory*1024), // KiB -> bytes
	)

	ch <- prometheus.MustNewConstMetric(
		c.ProcessesMax,
		prometheus.GaugeValue,
		float64(dst[0].MaxNumberOfProcesses),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ProcessMemoryBytesBytes,
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
		c.SizeStoredInPagingFilesBytes,
		prometheus.GaugeValue,
		float64(dst[0].SizeStoredInPagingFiles*1024), // KiB -> bytes
	)

	ch <- prometheus.MustNewConstMetric(
		c.VirtualMemoryBytesTotal,
		prometheus.GaugeValue,
		float64(dst[0].TotalVirtualMemorySize*1024), // KiB -> bytes
	)

	ch <- prometheus.MustNewConstMetric(
		c.VisibleMemoryBytesTotal,
		prometheus.GaugeValue,
		float64(dst[0].TotalVisibleMemorySize*1024), // KiB -> bytes
	)

	return nil, nil
}
