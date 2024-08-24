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

// A Collector is a Prometheus Collector for WMI metrics.
type Collector struct {
	config Config

	physicalMemoryBytes *prometheus.Desc
	logicalProcessors   *prometheus.Desc
	hostname            *prometheus.Desc
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
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ log.Logger) error {
	c.logicalProcessors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "logical_processors"),
		"ComputerSystem.NumberOfLogicalProcessors",
		nil,
		nil,
	)
	c.physicalMemoryBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "physical_memory_bytes"),
		"ComputerSystem.TotalPhysicalMemory",
		nil,
		nil,
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
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	if err := c.collect(ch); err != nil {
		_ = level.Error(logger).Log("msg", "failed collecting cs metrics", "err", err)
		return err
	}
	return nil
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
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
