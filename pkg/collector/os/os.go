//go:build windows

package os

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/headers/kernel32"
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

// A Collector is a Prometheus Collector for WMI metrics.
type Collector struct {
	config Config

	osInformation           *prometheus.Desc
	pagingFreeBytes         *prometheus.Desc
	pagingLimitBytes        *prometheus.Desc
	physicalMemoryFreeBytes *prometheus.Desc
	processMemoryLimitBytes *prometheus.Desc
	processes               *prometheus.Desc
	processesLimit          *prometheus.Desc
	time                    *prometheus.Desc
	timezone                *prometheus.Desc
	users                   *prometheus.Desc
	virtualMemoryBytes      *prometheus.Desc
	virtualMemoryFreeBytes  *prometheus.Desc
	visibleMemoryBytes      *prometheus.Desc
}

type pagingFileCounter struct {
	Name      string
	Usage     float64 `perflib:"% Usage"`
	UsagePeak float64 `perflib:"% Usage Peak"`
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

func (c *Collector) GetPerfCounter(_ log.Logger) ([]string, error) {
	return []string{"Paging File"}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ log.Logger) error {
	c.osInformation = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"OperatingSystem.Caption, OperatingSystem.Version",
		[]string{"product", "version", "major_version", "minor_version", "build_number", "revision"},
		nil,
	)
	c.pagingLimitBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "paging_limit_bytes"),
		"OperatingSystem.SizeStoredInPagingFiles",
		nil,
		nil,
	)
	c.pagingFreeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "paging_free_bytes"),
		"OperatingSystem.FreeSpaceInPagingFiles",
		nil,
		nil,
	)
	c.physicalMemoryFreeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "physical_memory_free_bytes"),
		"OperatingSystem.FreePhysicalMemory",
		nil,
		nil,
	)
	c.time = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "time"),
		"OperatingSystem.LocalDateTime",
		nil,
		nil,
	)
	c.timezone = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "timezone"),
		"OperatingSystem.LocalDateTime",
		[]string{"timezone"},
		nil,
	)
	c.processes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processes"),
		"OperatingSystem.NumberOfProcesses",
		nil,
		nil,
	)
	c.processesLimit = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processes_limit"),
		"OperatingSystem.MaxNumberOfProcesses",
		nil,
		nil,
	)
	c.processMemoryLimitBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "process_memory_limit_bytes"),
		"OperatingSystem.MaxProcessMemorySize",
		nil,
		nil,
	)
	c.users = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "users"),
		"OperatingSystem.NumberOfUsers",
		nil,
		nil,
	)
	c.virtualMemoryBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_memory_bytes"),
		"OperatingSystem.TotalVirtualMemorySize",
		nil,
		nil,
	)
	c.visibleMemoryBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "visible_memory_bytes"),
		"OperatingSystem.TotalVisibleMemorySize",
		nil,
		nil,
	)
	c.virtualMemoryFreeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_memory_free_bytes"),
		"OperatingSystem.FreeVirtualMemory",
		nil,
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	if err := c.collect(ctx, logger, ch); err != nil {
		_ = level.Error(logger).Log("msg", "failed collecting os metrics", "err", err)
		return err
	}
	return nil
}

// Win32_OperatingSystem docs:
// - https://msdn.microsoft.com/en-us/library/aa394239 - Win32_OperatingSystem class.
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

func (c *Collector) collect(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	nwgi, err := netapi32.GetWorkstationInfo()
	if err != nil {
		return err
	}

	gmse, err := sysinfoapi.GlobalMemoryStatusEx()
	if err != nil {
		return err
	}

	currentTime := time.Now()

	timeZoneInfo, err := kernel32.GetDynamicTimeZoneInformation()
	if err != nil {
		return err
	}

	// timeZoneKeyName contains the english name of the timezone.
	timezoneName := syscall.UTF16ToString(timeZoneInfo.TimeZoneKeyName[:])

	// Get total allocation of paging files across all disks.
	memManKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management`, registry.QUERY_VALUE)
	if err != nil {
		return err
	}

	defer memManKey.Close()

	pagingFiles, _, pagingErr := memManKey.GetStringsValue("ExistingPageFiles")

	var fsipf float64
	for _, pagingFile := range pagingFiles {
		fileString := strings.ReplaceAll(pagingFile, `\??\`, "")
		file, err := os.Stat(fileString)
		// For unknown reasons, Windows doesn't always create a page file. Continue collection rather than aborting.
		if err != nil {
			_ = level.Debug(logger).Log("msg", fmt.Sprintf("Failed to read page file (reason: %s): %s\n", err, fileString))
		} else {
			fsipf += float64(file.Size())
		}
	}

	// Get build number and product name from registry
	ntKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return err
	}

	defer ntKey.Close()

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

	gpi, err := psapi.GetPerformanceInfo()
	if err != nil {
		return err
	}

	pfc := make([]pagingFileCounter, 0)
	if err := perflib.UnmarshalObject(ctx.PerfObjects["Paging File"], &pfc, logger); err != nil {
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
		c.osInformation,
		prometheus.GaugeValue,
		1.0,
		"Microsoft "+pn, // Caption
		fmt.Sprintf("%d.%d.%s", nwgi.VersionMajor, nwgi.VersionMinor, bn), // Version
		strconv.FormatUint(uint64(nwgi.VersionMajor), 10),                 // Major Version
		strconv.FormatUint(uint64(nwgi.VersionMinor), 10),                 // Minor Version
		bn,                               // Build number
		strconv.FormatUint(revision, 10), // Revision
	)

	ch <- prometheus.MustNewConstMetric(
		c.physicalMemoryFreeBytes,
		prometheus.GaugeValue,
		float64(gmse.AvailPhys),
	)

	ch <- prometheus.MustNewConstMetric(
		c.time,
		prometheus.GaugeValue,
		float64(currentTime.Unix()),
	)

	ch <- prometheus.MustNewConstMetric(
		c.timezone,
		prometheus.GaugeValue,
		1.0,
		timezoneName,
	)

	if pagingErr == nil {
		ch <- prometheus.MustNewConstMetric(
			c.pagingFreeBytes,
			prometheus.GaugeValue,
			pfb,
		)

		ch <- prometheus.MustNewConstMetric(
			c.pagingLimitBytes,
			prometheus.GaugeValue,
			fsipf,
		)
	} else {
		_ = level.Debug(logger).Log("msg", "Could not find HKLM:\\SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Memory Management key. windows_os_paging_free_bytes and windows_os_paging_limit_bytes will be omitted.")
	}
	ch <- prometheus.MustNewConstMetric(
		c.virtualMemoryFreeBytes,
		prometheus.GaugeValue,
		float64(gmse.AvailPageFile),
	)

	// Windows has no defined limit, and is based off available resources. This currently isn't calculated by WMI and is set to default value.
	// https://techcommunity.microsoft.com/t5/windows-blog-archive/pushing-the-limits-of-windows-processes-and-threads/ba-p/723824
	// https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-operatingsystem
	ch <- prometheus.MustNewConstMetric(
		c.processesLimit,
		prometheus.GaugeValue,
		float64(4294967295),
	)

	ch <- prometheus.MustNewConstMetric(
		c.processMemoryLimitBytes,
		prometheus.GaugeValue,
		float64(gmse.TotalVirtual),
	)

	ch <- prometheus.MustNewConstMetric(
		c.processes,
		prometheus.GaugeValue,
		float64(gpi.ProcessCount),
	)

	ch <- prometheus.MustNewConstMetric(
		c.users,
		prometheus.GaugeValue,
		float64(nwgi.LoggedOnUsers),
	)

	ch <- prometheus.MustNewConstMetric(
		c.virtualMemoryBytes,
		prometheus.GaugeValue,
		float64(gmse.TotalPageFile),
	)

	ch <- prometheus.MustNewConstMetric(
		c.visibleMemoryBytes,
		prometheus.GaugeValue,
		float64(gmse.TotalPhys),
	)

	return nil
}
