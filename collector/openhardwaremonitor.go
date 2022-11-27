package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("openhardwaremonitor", NewOpenHardwareMonitorCollector)
}

// A openHardwareMonitorCollector is a Prometheus collector for WMI Sensor metrics
type SensorCollector struct {
	SensorType *prometheus.Desc
	Identifier *prometheus.Desc
	Parent     *prometheus.Desc
	Name       *prometheus.Desc
	Value      *prometheus.Desc
	Max        *prometheus.Desc
	Min        *prometheus.Desc
	Index      *prometheus.Desc
}

// NewOpenHardwareMonitorCollector ...
func NewOpenHardwareMonitorCollector() (Collector, error) {
	const subsystem = "openHardwareMonitor"
	return &SensorCollector{
		Value: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "value"),
			"(Value)",
			[]string{"identifier", "name", "sensor_type"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *SensorCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting openHardwareMonitor metrics:", desc, err)
		return err
	}
	return nil
}

// OpenHardwareMonitor Sensor:
type Sensor struct {
	SensorType string
	Identifier string
	Parent     string
	Name       string
	Value      float64
	Max        float64
	Min        float64
	Index      uint32
}

func (c *SensorCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Sensor
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, info := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.Value,
			prometheus.GaugeValue,
			float64(info.Value),
			info.Identifier,
			info.Name,
			info.SensorType,
		)
	}

	return nil, nil
}
