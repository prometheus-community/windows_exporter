package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["thermalzone"] = NewThermalZoneCollector
}

// A ThermalZoneCollector is a Prometheus collector for WMI Win32_PerfRawData_Counters_ThermalZoneInformation metrics
type ThermalZoneCollector struct {
	HighPrecisionTemperature *prometheus.Desc
	PercentPassiveLimit      *prometheus.Desc
	Temperature              *prometheus.Desc
	ThrottleReasons          *prometheus.Desc
}

// NewThermalZoneCollector ...
func NewThermalZoneCollector() (Collector, error) {
	const subsystem = "thermalzone"
	return &ThermalZoneCollector{
		HighPrecisionTemperature: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "high_precision_temperature"),
			"(HighPrecisionTemperature)",
			nil,
			nil,
		),
		PercentPassiveLimit: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "percent_passive_limit"),
			"(PercentPassiveLimit)",
			nil,
			nil,
		),
		Temperature: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "temperature"),
			"(Temperature)",
			nil,
			nil,
		),
		ThrottleReasons: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "throttle_reasons"),
			"(ThrottleReasons)",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *ThermalZoneCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting thermalzone metrics:", desc, err)
		return err
	}
	return nil
}

// Win32_PerfRawData_Counters_ThermalZoneInformation docs:
// - <add link to documentation here>
type Win32_PerfRawData_Counters_ThermalZoneInformation struct {
	Name string

	HighPrecisionTemperature uint32
	PercentPassiveLimit      uint32
	Temperature              uint32
	ThrottleReasons          uint32
}

func (c *ThermalZoneCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_ThermalZoneInformation
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.HighPrecisionTemperature,
		prometheus.GaugeValue,
		float64(dst[0].HighPrecisionTemperature),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentPassiveLimit,
		prometheus.GaugeValue,
		float64(dst[0].PercentPassiveLimit),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Temperature,
		prometheus.GaugeValue,
		float64(dst[0].Temperature),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ThrottleReasons,
		prometheus.GaugeValue,
		float64(dst[0].ThrottleReasons),
	)

	return nil, nil
}
