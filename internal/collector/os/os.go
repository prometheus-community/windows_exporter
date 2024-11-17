//go:build windows

package os

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/kernel32"
	"github.com/prometheus-community/windows_exporter/internal/headers/netapi32"
	"github.com/prometheus-community/windows_exporter/internal/headers/sysinfoapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const Name = "os"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI metrics.
type Collector struct {
	config Config

	hostname      *prometheus.Desc
	osInformation *prometheus.Desc

	// users
	// Deprecated: Use windows_system_process_limit instead.
	processesLimit *prometheus.Desc

	// users
	// Deprecated: Use count(windows_logon_logon_type) instead.
	users *prometheus.Desc

	// physicalMemoryFreeBytes
	// Deprecated: Use windows_memory_physical_free_bytes instead.
	physicalMemoryFreeBytes *prometheus.Desc

	// processMemoryLimitBytes
	// Deprecated: Use windows_memory_process_memory_limit_bytes instead.
	processMemoryLimitBytes *prometheus.Desc

	// time
	// Deprecated: Use windows_time_current_timestamp_seconds instead.
	time *prometheus.Desc
	// timezone
	// Deprecated: Use windows_time_timezone instead.
	timezone *prometheus.Desc
	// virtualMemoryBytes
	// Deprecated: Use windows_memory_commit_limit instead.
	virtualMemoryBytes *prometheus.Desc
	// virtualMemoryFreeBytes
	// Deprecated: Use windows_memory_commit_limit instead.
	virtualMemoryFreeBytes *prometheus.Desc
	// visibleMemoryBytes
	// Deprecated: Use windows_memory_physical_total_bytes instead.
	visibleMemoryBytes *prometheus.Desc
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
	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	logger.Warn("The os collect holds a number of deprecated metrics and will be removed mid 2025. "+
		"See https://github.com/prometheus-community/windows_exporter/pull/1596 for more information.",
		slog.String("collector", Name),
	)

	workstationInfo, err := netapi32.GetWorkstationInfo()
	if err != nil {
		return fmt.Errorf("failed to get workstation info: %w", err)
	}

	productName, buildNumber, revision, err := c.getWindowsVersion()
	if err != nil {
		return fmt.Errorf("failed to get Windows version: %w", err)
	}

	c.osInformation = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		`Contains full product name & version in labels. Note that the "major_version" for Windows 11 is \"10\"; a build number greater than 22000 represents Windows 11.`,
		nil,
		prometheus.Labels{
			"product":       productName,
			"version":       fmt.Sprintf("%d.%d.%s", workstationInfo.VersionMajor, workstationInfo.VersionMinor, buildNumber),
			"major_version": strconv.FormatUint(uint64(workstationInfo.VersionMajor), 10),
			"minor_version": strconv.FormatUint(uint64(workstationInfo.VersionMinor), 10),
			"build_number":  buildNumber,
			"revision":      revision,
		},
	)

	c.hostname = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hostname"),
		"Labelled system hostname information as provided by ComputerSystem.DNSHostName and ComputerSystem.Domain",
		[]string{
			"hostname",
			"domain",
			"fqdn",
		},
		nil,
	)
	c.physicalMemoryFreeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "physical_memory_free_bytes"),
		"Deprecated: Use `windows_memory_physical_free_bytes` instead.",
		nil,
		nil,
	)
	c.time = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "time"),
		"Deprecated: Use windows_time_current_timestamp_seconds instead.",
		nil,
		nil,
	)
	c.timezone = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "timezone"),
		"Deprecated: Use windows_time_timezone instead.",
		[]string{"timezone"},
		nil,
	)

	c.processesLimit = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processes_limit"),
		"Deprecated: Use `windows_system_process_limit` instead.",
		nil,
		nil,
	)
	c.processMemoryLimitBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "process_memory_limit_bytes"),
		"Deprecated: Use `windows_memory_process_memory_limit_bytes` instead.",
		nil,
		nil,
	)
	c.users = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "users"),
		"Deprecated: Use `count(windows_logon_logon_type)` instead.",
		nil,
		nil,
	)
	c.virtualMemoryBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_memory_bytes"),
		"Deprecated: Use `windows_memory_commit_limit` instead.",
		nil,
		nil,
	)
	c.visibleMemoryBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "visible_memory_bytes"),
		"Deprecated: Use `windows_memory_physical_total_bytes` instead.",
		nil,
		nil,
	)
	c.virtualMemoryFreeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_memory_free_bytes"),
		"Deprecated: Use `windows_memory_commit_limit - windows_memory_committed_bytes` instead.",
		nil,
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0, 4)

	c.collect(ch)

	if err := c.collectHostname(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect hostname metrics: %w", err))
	}

	if err := c.collectLoggedInUserCount(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect user count metrics: %w", err))
	}

	if err := c.collectMemory(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect memory metrics: %w", err))
	}

	if err := c.collectTime(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect time metrics: %w", err))
	}

	return errors.Join(errs...)
}

func (c *Collector) collectLoggedInUserCount(ch chan<- prometheus.Metric) error {
	workstationInfo, err := netapi32.GetWorkstationInfo()
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.users,
		prometheus.GaugeValue,
		float64(workstationInfo.LoggedOnUsers),
	)

	return nil
}

func (c *Collector) collectHostname(ch chan<- prometheus.Metric) error {
	hostname, err := sysinfoapi.GetComputerName(sysinfoapi.ComputerNameDNSHostname)
	if err != nil {
		return err
	}

	domain, err := sysinfoapi.GetComputerName(sysinfoapi.ComputerNameDNSDomain)
	if err != nil {
		return err
	}

	fqdn, err := sysinfoapi.GetComputerName(sysinfoapi.ComputerNameDNSFullyQualified)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.hostname,
		prometheus.GaugeValue,
		1.0,
		hostname,
		domain,
		fqdn,
	)

	return nil
}

func (c *Collector) collectTime(ch chan<- prometheus.Metric) error {
	timeZoneInfo, err := kernel32.GetDynamicTimeZoneInformation()
	if err != nil {
		return err
	}

	// timeZoneKeyName contains the english name of the timezone.
	timezoneName := windows.UTF16ToString(timeZoneInfo.TimeZoneKeyName[:])

	ch <- prometheus.MustNewConstMetric(
		c.time,
		prometheus.GaugeValue,
		float64(time.Now().Unix()),
	)

	ch <- prometheus.MustNewConstMetric(
		c.timezone,
		prometheus.GaugeValue,
		1.0,
		timezoneName,
	)

	return nil
}

func (c *Collector) collectMemory(ch chan<- prometheus.Metric) error {
	memoryStatusEx, err := sysinfoapi.GlobalMemoryStatusEx()
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.physicalMemoryFreeBytes,
		prometheus.GaugeValue,
		float64(memoryStatusEx.AvailPhys),
	)

	ch <- prometheus.MustNewConstMetric(
		c.virtualMemoryFreeBytes,
		prometheus.GaugeValue,
		float64(memoryStatusEx.AvailPageFile),
	)

	ch <- prometheus.MustNewConstMetric(
		c.virtualMemoryBytes,
		prometheus.GaugeValue,
		float64(memoryStatusEx.TotalPageFile),
	)

	ch <- prometheus.MustNewConstMetric(
		c.visibleMemoryBytes,
		prometheus.GaugeValue,
		float64(memoryStatusEx.TotalPhys),
	)

	ch <- prometheus.MustNewConstMetric(
		c.processMemoryLimitBytes,
		prometheus.GaugeValue,
		float64(memoryStatusEx.TotalVirtual),
	)

	return nil
}

func (c *Collector) collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		c.osInformation,
		prometheus.GaugeValue,
		1.0,
	)

	// Windows has no defined limit, and is based off available resources. This currently isn't calculated by WMI and is set to default value.
	// https://techcommunity.microsoft.com/t5/windows-blog-archive/pushing-the-limits-of-windows-processes-and-threads/ba-p/723824
	// https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-operatingsystem
	ch <- prometheus.MustNewConstMetric(
		c.processesLimit,
		prometheus.GaugeValue,
		float64(4294967295),
	)
}

func (c *Collector) getWindowsVersion() (string, string, string, error) {
	// Get build number and product name from registry
	ntKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to open registry key: %w", err)
	}

	defer ntKey.Close()

	productName, _, err := ntKey.GetStringValue("ProductName")
	if err != nil {
		return "", "", "", err
	}

	buildNumber, _, err := ntKey.GetStringValue("CurrentBuildNumber")
	if err != nil {
		return "", "", "", err
	}

	revision, _, err := ntKey.GetIntegerValue("UBR")
	if errors.Is(err, registry.ErrNotExist) {
		revision = 0
	} else if err != nil {
		return "", "", "", err
	}

	return productName, buildNumber, strconv.FormatUint(revision, 10), nil
}
