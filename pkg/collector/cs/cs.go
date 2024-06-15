//go:build windows

package cs

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/headers/sysinfoapi"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "cs"

type Config struct{}

var ConfigDefaults = Config{}

// A collector is a Prometheus collector for WMI metrics
type collector struct {
	logger log.Logger

	PhysicalMemoryBytes *prometheus.Desc
	LogicalProcessors   *prometheus.Desc
	Hostname            *prometheus.Desc
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
	return []string{}, nil
}

func (c *collector) Build() error {
	c.LogicalProcessors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "logical_processors"),
		"ComputerSystem.NumberOfLogicalProcessors",
		nil,
		nil,
	)
	c.PhysicalMemoryBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "physical_memory_bytes"),
		"ComputerSystem.TotalPhysicalMemory",
		nil,
		nil,
	)
	c.Hostname = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hostname"),
		"Labelled system hostname information as provided by ComputerSystem.DNSHostName and ComputerSystem.Domain",
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
func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting cs metrics", "err", err)
		return err
	}
	return nil
}

func (c *collector) collect(ch chan<- prometheus.Metric) error {
	// Get systeminfo for number of processors
	systemInfo := sysinfoapi.GetSystemInfo()

	// Get memory status for physical memory
	mem, err := sysinfoapi.GlobalMemoryStatusEx()
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.LogicalProcessors,
		prometheus.GaugeValue,
		float64(systemInfo.NumberOfProcessors),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PhysicalMemoryBytes,
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
		c.Hostname,
		prometheus.GaugeValue,
		1.0,
		hostname,
		domain,
		fqdn,
	)

	return nil
}
