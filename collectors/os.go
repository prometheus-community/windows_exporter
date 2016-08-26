// returns data points from Win32_OperatingSystem

package collectors

import (
	"log"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

// A OSCollector is a Prometheus collector for WMI OperatingSystem metrics
type OSCollector struct {
	FreePhysicalMemory     *prometheus.Desc
	FreeSpaceInPagingFiles *prometheus.Desc
	FreeVirtualMemory      *prometheus.Desc
	NumberOfProcesses      *prometheus.Desc
	NumberOfUsers          *prometheus.Desc
}

// NewOSCollector ...
func NewOSCollector() *OSCollector {

	return &OSCollector{
		FreePhysicalMemory: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "free_physical_memory"),
			"Free physical memory.",
			nil,
			nil,
		),

		FreeSpaceInPagingFiles: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "free_space_in_paging_files"),
			"Free space in paging files.",
			nil,
			nil,
		),

		FreeVirtualMemory: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "free_virtual_memory"),
			"Free virtual memory.",
			nil,
			nil,
		),

		NumberOfProcesses: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "number_of_processes"),
			"No. of processes running on the system.",
			nil,
			nil,
		),

		NumberOfUsers: prometheus.NewDesc(
			prometheus.BuildFQName(wmiNamespace, "os", "number_of_users"),
			"No. of users logged in.",
			nil,
			nil,
		),
	}
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *OSCollector) Collect(ch chan<- prometheus.Metric) {
	if desc, err := c.collect(ch); err != nil {
		log.Println("[ERROR] failed collecting process metrics:", desc, err)
		return
	}
}

// Describe sends the descriptors of each metric over to the provided channel.
// The corresponding metric values are sent separately.
func (c *OSCollector) Describe(ch chan<- *prometheus.Desc) {

	ch <- c.FreePhysicalMemory
	ch <- c.FreeSpaceInPagingFiles
	ch <- c.FreeVirtualMemory
	ch <- c.NumberOfProcesses
	ch <- c.NumberOfUsers
}

type Win32_OperatingSystem struct {
	FreePhysicalMemory     uint64
	FreeSpaceInPagingFiles uint64
	FreeVirtualMemory      uint64
	NumberOfProcesses      uint32
	NumberOfUsers          uint32
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
		c.NumberOfProcesses,
		prometheus.GaugeValue,
		float64(dst[0].NumberOfProcesses),
	)

	ch <- prometheus.MustNewConstMetric(
		c.NumberOfUsers,
		prometheus.GaugeValue,
		float64(dst[0].NumberOfUsers),
	)

	return nil, nil
}
