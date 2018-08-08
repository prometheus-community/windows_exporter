// returns data points from Win32_ComputerSystem
// https://msdn.microsoft.com/en-us/library/aa394102 - Win32_ComputerSystem class

package collector

import (
	"errors"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["cs"] = NewCSCollector
}

// A CSCollector is a Prometheus collector for WMI metrics
type CSCollector struct {
	PhysicalMemoryBytes *prometheus.Desc
	LogicalProcessors   *prometheus.Desc
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
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *CSCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting cs metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_ComputerSystem struct {
	NumberOfLogicalProcessors uint32
	TotalPhysicalMemory       uint64
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

	return nil, nil
}
