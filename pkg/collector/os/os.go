//go:build windows

package os

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/headers/netapi32"
	"github.com/prometheus-community/windows_exporter/pkg/headers/psapi"
	"github.com/prometheus-community/windows_exporter/pkg/headers/sysinfoapi"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows/registry"
)

const Name = "os"

type Config struct{}

var ConfigDefaults = Config{}

// A collector is a Prometheus collector for WMI metrics
type collector struct {
	logger log.Logger

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

func New(logger log.Logger, _ *Config) types.Collector {
	c := &collector{}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(_ *kingpin.Application) types.Collector {
	return &collector{}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{"Paging File"}, nil
}

func (c *collector) Build() error {
	c.OSInformation = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"OperatingSystem.Caption, OperatingSystem.Version",
		[]string{"product", "version", "major_version", "minor_version", "build_number", "revision"},
		nil,
	)
	c.PagingLimitBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "paging_limit_bytes"),
		"OperatingSystem.SizeStoredInPagingFiles",
		nil,
		nil,
	)
	c.PagingFreeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "paging_free_bytes"),
		"OperatingSystem.FreeSpaceInPagingFiles",
		nil,
		nil,
	)
	c.PhysicalMemoryFreeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "physical_memory_free_bytes"),
		"OperatingSystem.FreePhysicalMemory",
		nil,
		nil,
	)
	c.Time = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "time"),
		"OperatingSystem.LocalDateTime",
		nil,
		nil,
	)
	c.Timezone = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "timezone"),
		"OperatingSystem.LocalDateTime",
		[]string{"timezone"},
		nil,
	)
	c.Processes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processes"),
		"OperatingSystem.NumberOfProcesses",
		nil,
		nil,
	)
	c.ProcessesLimit = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processes_limit"),
		"OperatingSystem.MaxNumberOfProcesses",
		nil,
		nil,
	)
	c.ProcessMemoryLimitBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "process_memory_limit_bytes"),
		"OperatingSystem.MaxProcessMemorySize",
		nil,
		nil,
	)
	c.Users = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "users"),
		"OperatingSystem.NumberOfUsers",
		nil,
		nil,
	)
	c.VirtualMemoryBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_memory_bytes"),
		"OperatingSystem.TotalVirtualMemorySize",
		nil,
		nil,
	)
	c.VisibleMemoryBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "visible_memory_bytes"),
		"OperatingSystem.TotalVisibleMemorySize",
		nil,
		nil,
	)
	c.VirtualMemoryFreeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_memory_free_bytes"),
		"OperatingSystem.FreeVirtualMemory",
		nil,
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ctx, ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting os metrics", "err", err)
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

func (c *collector) collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	nwgi, err := netapi32.GetWorkstationInfo()
	if err != nil {
		return err
	}

	gmse, err := sysinfoapi.GlobalMemoryStatusEx()
	if err != nil {
		return err
	}

	currentTime := time.Now()
	timezoneName, _ := currentTime.Zone()

	// Get total allocation of paging files across all disks.
	memManKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management`, registry.QUERY_VALUE)
	defer memManKey.Close()

	if err != nil {
		return err
	}
	pagingFiles, _, pagingErr := memManKey.GetStringsValue("ExistingPageFiles")
	// Get build number and product name from registry
	ntKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	defer ntKey.Close()

	if err != nil {
		return err
	}

	pn, _, err := ntKey.GetStringValue("ProductName")
	if err != nil {
		return err
	}

	bn, _, err := ntKey.GetStringValue("CurrentBuildNumber")
	if err != nil {
		return err
	}

	revision, _, err := ntKey.GetIntegerValue("UBR")
	if errors.Is(err, registry.ErrNotExist) {
		revision = 0
	} else if err != nil {
		return err
	}

	var fsipf float64
	for _, pagingFile := range pagingFiles {
		fileString := strings.ReplaceAll(pagingFile, `\??\`, "")
		file, err := os.Stat(fileString)
		// For unknown reasons, Windows doesn't always create a page file. Continue collection rather than aborting.
		if err != nil {
			_ = level.Debug(c.logger).Log("msg", fmt.Sprintf("Failed to read page file (reason: %s): %s\n", err, fileString))
		} else {
			fsipf += float64(file.Size())
		}
	}

	gpi, err := psapi.GetPerformanceInfo()
	if err != nil {
		return err
	}

	pfc := make([]pagingFileCounter, 0)
	if err := perflib.UnmarshalObject(ctx.PerfObjects["Paging File"], &pfc, c.logger); err != nil {
		return err
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
		fmt.Sprintf("%d", revision),                                       // Revision
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
		_ = level.Debug(c.logger).Log("msg", "Could not find HKLM:\\SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Memory Management key. windows_os_paging_free_bytes and windows_os_paging_limit_bytes will be omitted.")
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

	return nil
}
