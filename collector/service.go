// returns data points from Win32_Service
// https://msdn.microsoft.com/en-us/library/aa394418(v=vs.85).aspx - Win32_Service class
package collector

import (
	"log"
	"strings"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	Factories["service"] = NewserviceCollector
}

// A serviceCollector is a Prometheus collector for WMI Win32_Service metrics
type serviceCollector struct {
	State     *prometheus.Desc
	StartMode *prometheus.Desc
}

// NewserviceCollector ...
func NewserviceCollector() (Collector, error) {
	const subsystem = "service"
	return &serviceCollector{
		State: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "state"),
			"The state of the service (State)",
			[]string{"name", "state"},
			nil,
		),
		StartMode: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "start_mode"),
			"The start mode of the service (StartMode)",
			[]string{"name", "start_mode"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *serviceCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Println("[ERROR] failed collecting service metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_Service struct {
	Name      string
	State     string
	StartMode string
}

func (c *serviceCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_Service
	q := wmi.CreateQuery(&dst, "")
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, service := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.State,
			prometheus.GaugeValue,
			1.0,
			strings.ToLower(service.Name),
			strings.ToLower(service.State),
		)

		ch <- prometheus.MustNewConstMetric(
			c.StartMode,
			prometheus.GaugeValue,
			1.0,
			strings.ToLower(service.Name),
			strings.ToLower(service.StartMode),
		)
	}
	return nil, nil
}
