//go:build windows
// +build windows

package collector

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/prometheus-community/windows_exporter/headers/netapi32"
	"github.com/prometheus-community/windows_exporter/headers/psapi"
	"github.com/prometheus-community/windows_exporter/headers/sysinfoapi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows/registry"
)

func init() {
	registerCollector("os", NewOSCollector, "Paging File")
}

// A OSCollector is a Prometheus collector for WMI metrics
type OSCollector struct {
	OSInformation           *prometheus.Desc
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

type pagingFileCounter struct {
	Name      string
	Usage     float64 `perflib:"% Usage"`
	UsagePeak float64 `perflib:"% Usage Peak"`
}

// NewOSCollector ...
func NewOSCollector() (Collector, error) {
	const subsystem = "os"

	return &OSCollector{
		OSInformation: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "info"),
			"OperatingSystem.Caption, OperatingSystem.Version",
			[]string{"product", "version", "major_version", "minor_version", "build_number"},
			nil,
		),
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
			prometheus.BuildFQName(Namespace, subsystem, "process_memory_limit_bytes"),
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
func (c *OSCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting os metrics:", desc, err)
		return err
	}
	return nil
}

// Win32_OperatingSystem docs:
// - https://msdn.microsoft.com/en-us/library/aa394239 - Win32_OperatingSystem class
type Win32_OperatingSystem struct {
	Caption                 string
	FreePhysicalMemory      uint64
	FreeSpaceInPagingFiles  uint64
	FreeVirtualMemory       uint64
	LocalDateTime           time.Time
	MaxNumberOfProcesses    uint32
	MaxProcessMemorySize    uint64
	NumberOfProcesses       uint32
	NumberOfUsers           uint32
	SizeStoredInPagingFiles uint64
	TotalVirtualMemorySize  uint64
	TotalVisibleMemorySize  uint64
	Version                 string
}

func (c *OSCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	nwgi, err := netapi32.GetWorkstationInfo()
	if err != nil {
		return nil, err
	}

	gmse, err := sysinfoapi.GlobalMemoryStatusEx()
	if err != nil {
		return nil, err
	}

	currentTime := time.Now()
	timezoneName, _ := currentTime.Zone()

	// Get total allocation of paging files across all disks.
	memManKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management`, registry.QUERY_VALUE)
	defer memManKey.Close()

	if err != nil {
		return nil, err
	}
	pagingFiles, _, pagingErr := memManKey.GetStringsValue("ExistingPageFiles")
	// Get build number and product name from registry
	ntKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	defer ntKey.Close()

	if err != nil {
		return nil, err
	}

	pn, _, err := ntKey.GetStringValue("ProductName")
	if err != nil {
		return nil, err
	}

	bn, _, err := ntKey.GetStringValue("CurrentBuildNumber")
	if err != nil {
		return nil, err
	}

	var fsipf float64
	for _, pagingFile := range pagingFiles {
		fileString := strings.ReplaceAll(pagingFile, `\??\`, "")
		file, err := os.Stat(fileString)
		if err != nil {
			return nil, err
		}
		fsipf += float64(file.Size())
	}

	gpi, err := psapi.GetPerformanceInfo()
	if err != nil {
		return nil, err
	}

	var pfc = make([]pagingFileCounter, 0)
	if err := unmarshalObject(ctx.perfObjects["Paging File"], &pfc); err != nil {
		return nil, err
	}

	// Get current page file usage.
	var pfbRaw float64
	for _, pageFile := range pfc {
		if strings.Contains(strings.ToLower(pageFile.Name), "_total") {
			continue
		}
		pfbRaw += pageFile.Usage
	}

	// Subtract from total page file allocation on disk.
	pfb := fsipf - (pfbRaw * float64(gpi.PageSize))

	ch <- prometheus.MustNewConstMetric(
		c.OSInformation,
		prometheus.GaugeValue,
		1.0,
		fmt.Sprintf("Microsoft %s", pn), // Caption
		fmt.Sprintf("%d.%d.%s", nwgi.VersionMajor, nwgi.VersionMinor, bn), // Version
		fmt.Sprintf("%d", nwgi.VersionMajor),                              // Major Version
		fmt.Sprintf("%d", nwgi.VersionMinor),                              // Minor Version
		bn,                                                                // Build number
	)

	ch <- prometheus.MustNewConstMetric(
		c.PhysicalMemoryFreeBytes,
		prometheus.GaugeValue,
		float64(gmse.AvailPhys),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Time,
		prometheus.GaugeValue,
		float64(currentTime.Unix()),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Timezone,
		prometheus.GaugeValue,
		1.0,
		timezoneName,
	)

	if pagingErr == nil {
		ch <- prometheus.MustNewConstMetric(
			c.PagingFreeBytes,
			prometheus.GaugeValue,
			pfb,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PagingLimitBytes,
			prometheus.GaugeValue,
			fsipf,
		)
	} else {
		log.Debugln("Could not find HKLM:\\SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Memory Management key. windows_os_paging_free_bytes and windows_os_paging_limit_bytes will be omitted.")
	}
	ch <- prometheus.MustNewConstMetric(
		c.VirtualMemoryFreeBytes,
		prometheus.GaugeValue,
		float64(gmse.AvailPageFile),
	)

	// Windows has no defined limit, and is based off available resources. This currently isn't calculated by WMI and is set to default value.
	// https://techcommunity.microsoft.com/t5/windows-blog-archive/pushing-the-limits-of-windows-processes-and-threads/ba-p/723824
	// https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-operatingsystem
	ch <- prometheus.MustNewConstMetric(
		c.ProcessesLimit,
		prometheus.GaugeValue,
		float64(4294967295),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ProcessMemoryLimitBytes,
		prometheus.GaugeValue,
		float64(gmse.TotalVirtual),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Processes,
		prometheus.GaugeValue,
		float64(gpi.ProcessCount),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Users,
		prometheus.GaugeValue,
		float64(nwgi.LoggedOnUsers),
	)

	ch <- prometheus.MustNewConstMetric(
		c.VirtualMemoryBytes,
		prometheus.GaugeValue,
		float64(gmse.TotalPageFile),
	)

	ch <- prometheus.MustNewConstMetric(
		c.VisibleMemoryBytes,
		prometheus.GaugeValue,
		float64(gmse.TotalPhys),
	)

	return nil, nil
}
