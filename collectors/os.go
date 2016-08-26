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
	FreePhysicalMemory      *prometheus.Desc
	FreeSpaceInPagingFiles  *prometheus.Desc
	FreeVirtualMemory       *prometheus.Desc
	MaxNumberOfProcesses    *prometheus.Desc
	MaxProcessMemorySize    *prometheus.Desc
	NumberOfProcesses       *prometheus.Desc
	NumberOfUsers           *prometheus.Desc
	SizeStoredInPagingFiles *prometheus.Desc
	TotalVirtualMemorySize  *prometheus.Desc
	TotalVisibleMemorySize  *prometheus.Desc
}

// NewOSCollector ...
func NewOSCollector() *OSCollector {

	return &OSCollector{
		FreePhysicalMemory: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "free_physical_memory"),
			"Number, in kilobytes, of physical memory currently unused and available.",
			nil,
			nil,
		),

		FreeSpaceInPagingFiles: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "free_space_in_paging_files"),
			"Number, in kilobytes, that can be mapped into the operating system paging files without causing any other pages to be swapped out.",
			nil,
			nil,
		),

		FreeVirtualMemory: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "free_virtual_memory"),
			"Number, in kilobytes, of virtual memory currently unused and available.",
			nil,
			nil,
		),

		MaxNumberOfProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "max_number_of_processes"),
			"Maximum number of process contexts the operating system can support.",
			nil,
			nil,
		),

		MaxProcessMemorySize: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "max_process_memory_size"),
			"Maximum number, in kilobytes, of memory that can be allocated to a process.",
			nil,
			nil,
		),

		NumberOfProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "number_of_processes"),
			"Number of process contexts currently loaded or running on the operating system.",
			nil,
			nil,
		),

		NumberOfUsers: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "number_of_users"),
			"Number of user sessions for which the operating system is storing state information currently.",
			nil,
			nil,
		),

		SizeStoredInPagingFiles: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "size_stored_in_paging_files"),
			"Total number of kilobytes that can be stored in the operating system paging files.",
			nil,
			nil,
		),

		TotalVirtualMemorySize: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "total_virtual_memory_size"),
			"Number, in kilobytes, of virtual memory.",
			nil,
			nil,
		),

		TotalVisibleMemorySize: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "total_visible_memory_size"),
			"Total amount, in kilobytes, of physical memory available to the operating system.",
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

	ch <- c.FreePhysicalMemory
	ch <- c.FreeSpaceInPagingFiles
	ch <- c.FreeVirtualMemory
	ch <- c.MaxNumberOfProcesses
	ch <- c.MaxProcessMemorySize
	ch <- c.NumberOfProcesses
	ch <- c.NumberOfUsers
	ch <- c.SizeStoredInPagingFiles
	ch <- c.TotalVirtualMemorySize
	ch <- c.TotalVisibleMemorySize
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
		c.FreePhysicalMemory,
		prometheus.GaugeValue,
		float64(dst[0].FreePhysicalMemory),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FreeSpaceInPagingFiles,
		prometheus.GaugeValue,
		float64(dst[0].FreeSpaceInPagingFiles),
	)

	ch <- prometheus.MustNewConstMetric(
		c.FreeVirtualMemory,
		prometheus.GaugeValue,
		float64(dst[0].FreeVirtualMemory),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MaxNumberOfProcesses,
		prometheus.GaugeValue,
		float64(dst[0].MaxNumberOfProcesses),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MaxProcessMemorySize,
		prometheus.GaugeValue,
		float64(dst[0].MaxProcessMemorySize),
	)

	ch <- prometheus.MustNewConstMetric(
		c.NumberOfProcesses,
		prometheus.GaugeValue,
		float64(dst[0].NumberOfProcesses),
	)

	ch <- prometheus.MustNewConstMetric(
		c.NumberOfUsers,
		prometheus.GaugeValue,
		float64(dst[0].NumberOfUsers),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SizeStoredInPagingFiles,
		prometheus.GaugeValue,
		float64(dst[0].SizeStoredInPagingFiles),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TotalVirtualMemorySize,
		prometheus.GaugeValue,
		float64(dst[0].TotalVirtualMemorySize),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TotalVisibleMemorySize,
		prometheus.GaugeValue,
		float64(dst[0].TotalVisibleMemorySize),
	)

	return nil, nil
}
