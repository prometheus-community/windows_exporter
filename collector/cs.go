// +build windows

package collector

import (
	"errors"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
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

// Win32_ComputerSystem docs:
// - https://msdn.microsoft.com/en-us/library/aa394102
type Win32_ComputerSystem struct {
	NumberOfLogicalProcessors uint32
	TotalPhysicalMemory       uint64
	DNSHostname               string
	Domain                    string
	Workgroup                 string
}

func (c *CSCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_ComputerSystem
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.LogicalProcessors,
		prometheus.GaugeValue,
		float64(dst[0].NumberOfLogicalProcessors),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PhysicalMemoryBytes,
		prometheus.GaugeValue,
		float64(dst[0].TotalPhysicalMemory),
	)

	var fqdn string
	if dst[0].Domain != dst[0].Workgroup {
		fqdn = dst[0].DNSHostname + "." + dst[0].Domain
	} else {
		fqdn = dst[0].DNSHostname
	}

	ch <- prometheus.MustNewConstMetric(
		c.Hostname,
		prometheus.GaugeValue,
		1.0,
		dst[0].DNSHostname,
		dst[0].Domain,
		fqdn,
	)

	return nil, nil
}
