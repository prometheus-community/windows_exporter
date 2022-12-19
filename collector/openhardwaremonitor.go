package collector

import (
	"strconv"

	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("open_hardware_monitor", NewOpenHardwareMonitorCollector)
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
	const subsystem = "open_hardware_monitor"
	return &SensorCollector{
		Value: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "sensor_value"),
			"Provides the value from an OpenHardwareMonitor sensor.",
			[]string{"parent", "index", "name", "sensor_type"},
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
	Value      float32
	Max        float32
	Min        float32
	Index      uint32
}

func (c *SensorCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Sensor
	q := queryAll(&dst)
	if err := wmi.QueryNamespace(q, &dst, "root/OpenHardwareMonitor"); err != nil {
		return nil, err
	}

	for _, info := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.Value,
			prometheus.GaugeValue,
			float64(info.Value),
			info.Parent,
			strconv.Itoa(int(info.Index)),
			info.Name,
			info.SensorType,
		)
	}

	return nil, nil
}
