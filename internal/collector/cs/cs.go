//go:build windows

package cs

import (
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/sysinfoapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "cs"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI metrics.
type Collector struct {
	config Config

	// physicalMemoryBytes
	// Deprecated: Use windows_physical_memory_total_bytes instead
	physicalMemoryBytes *prometheus.Desc
	// logicalProcessors
	// Deprecated: Use windows_cpu_logical_processor instead
	logicalProcessors *prometheus.Desc
	// hostname
	// Deprecated: Use windows_os_hostname instead
	hostname *prometheus.Desc
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
	logger.Warn("The cs collector is deprecated and will be removed in a future release. " +
		"Logical processors has been moved to cpu_info collector. " +
		"Physical memory has been moved to memory collector. " +
		"Hostname has been moved to os collector.")

	c.logicalProcessors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "logical_processors"),
		"Deprecated: Use windows_cpu_logical_processor instead",
		nil,
		nil,
	)
	c.physicalMemoryBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "physical_memory_bytes"),
		"Deprecated: Use windows_physical_memory_total_bytes instead",
		nil,
		nil,
	)
	c.hostname = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hostname"),
		"Deprecated: Use windows_os_hostname instead",
		[]string{
			"hostname",
			"domain",
			"fqdn",
		},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	// Get systeminfo for number of processors
	systemInfo := sysinfoapi.GetSystemInfo()

	// Get memory status for physical memory
	mem, err := sysinfoapi.GlobalMemoryStatusEx()
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.logicalProcessors,
		prometheus.GaugeValue,
		float64(systemInfo.NumberOfProcessors),
	)

	ch <- prometheus.MustNewConstMetric(
		c.physicalMemoryBytes,
		prometheus.GaugeValue,
		float64(mem.TotalPhys),
	)

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
