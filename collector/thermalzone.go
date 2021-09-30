package collector

import (
	"errors"

	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("thermalzone", NewThermalZoneCollector)
}

// A thermalZoneCollector is a Prometheus collector for WMI Win32_PerfRawData_Counters_ThermalZoneInformation metrics
type thermalZoneCollector struct {
	PercentPassiveLimit *prometheus.Desc
	Temperature         *prometheus.Desc
	ThrottleReasons     *prometheus.Desc
}

// NewThermalZoneCollector ...
func NewThermalZoneCollector() (Collector, error) {
	const subsystem = "thermalzone"
	return &thermalZoneCollector{
		Temperature: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "temperature_celsius"),
			"(Temperature)",
			[]string{
				"name",
			},
			nil,
		),
		PercentPassiveLimit: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "percent_passive_limit"),
			"(PercentPassiveLimit)",
			[]string{
				"name",
			},
			nil,
		),
		ThrottleReasons: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "throttle_reasons"),
			"(ThrottleReasons)",
			[]string{
				"name",
			},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *thermalZoneCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting thermalzone metrics:", desc, err)
		return err
	}
	return nil
}

// Win32_PerfRawData_Counters_ThermalZoneInformation docs:
// https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_thermalzoneinformation/
type Win32_PerfRawData_Counters_ThermalZoneInformation struct {
	Name string

	HighPrecisionTemperature uint32
	PercentPassiveLimit      uint32
	ThrottleReasons          uint32
}

func (c *thermalZoneCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_ThermalZoneInformation
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	// ThermalZone collector has been known to 'successfully' return an empty result.
	if len(dst) == 0 {
		return nil, errors.New("Empty results set for collector")
	}

	for _, info := range dst {
		//Divide by 10 and subtract 273.15 to convert decikelvin to celsius
		ch <- prometheus.MustNewConstMetric(
			c.Temperature,
			prometheus.GaugeValue,
			(float64(info.HighPrecisionTemperature)/10.0)-273.15,
			info.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.PercentPassiveLimit,
			prometheus.GaugeValue,
			float64(info.PercentPassiveLimit),
			info.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ThrottleReasons,
			prometheus.GaugeValue,
			float64(info.ThrottleReasons),
			info.Name,
		)
	}

	return nil, nil
}
