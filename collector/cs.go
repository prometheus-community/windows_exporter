//go:build windows
// +build windows

package collector

import (
	"github.com/prometheus-community/windows_exporter/headers/sysinfoapi"
	"github.com/prometheus-community/windows_exporter/log"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("cs", NewCSCollector)
}

// A CSCollector is a Prometheus collector for WMI metrics
type CSCollector struct {
	PhysicalMemoryBytes *prometheus.Desc
	LogicalProcessors   *prometheus.Desc
	Hostname            *prometheus.Desc
}

// NewCSCollector ...
func NewCSCollector() (Collector, error) {
	const subsystem = "cs"

	return &CSCollector{
		LogicalProcessors: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "logical_processors"),
			"ComputerSystem.NumberOfLogicalProcessors",
			nil,
			nil,
		),
		PhysicalMemoryBytes: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "physical_memory_bytes"),
			"ComputerSystem.TotalPhysicalMemory",
			nil,
			nil,
		),
		Hostname: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "hostname"),
			"Labeled system hostname information as provided by ComputerSystem.DNSHostName and ComputerSystem.Domain",
			[]string{
				"hostname",
				"domain",
				"fqdn"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *CSCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting cs metrics:", desc, err)
		return err
	}
	return nil
}

func (c *CSCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	// Get systeminfo for number of processors
	systemInfo := sysinfoapi.GetSystemInfo()

	// Get memory status for physical memory
	mem, err := sysinfoapi.GlobalMemoryStatusEx()
	if err != nil {
		return nil, err
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
		return nil, err
	}
	domain, err := sysinfoapi.GetComputerName(sysinfoapi.ComputerNameDNSDomain)
	if err != nil {
		return nil, err
	}
	fqdn, err := sysinfoapi.GetComputerName(sysinfoapi.ComputerNameDNSFullyQualified)
	if err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.Hostname,
		prometheus.GaugeValue,
		1.0,
		hostname,
		domain,
		fqdn,
	)

	return nil, nil
}
