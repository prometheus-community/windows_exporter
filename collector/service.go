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

var (
	allStates = []string{
		"stopped",
		"start pending",
		"stop pending",
		"running",
		"continue pending",
		"pause pending",
		"paused",
		"unknown",
	}
	allStartModes = []string{
		"boot",
		"system",
		"auto",
		"manual",
		"disabled",
	}
)

func (c *serviceCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_Service
	q := wmi.CreateQuery(&dst, "")
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, service := range dst {
		for _, state := range allStates {
			isCurrentState := 0.0
			if state == strings.ToLower(service.State) {
				isCurrentState = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.State,
				prometheus.GaugeValue,
				isCurrentState,
				strings.ToLower(service.Name),
				state,
			)
		}

		for _, startMode := range allStartModes {
			isCurrentStartMode := 0.0
			if startMode == strings.ToLower(service.StartMode) {
				isCurrentStartMode = 1.0
			}
			ch <- prometheus.MustNewConstMetric(
				c.StartMode,
				prometheus.GaugeValue,
				isCurrentStartMode,
				strings.ToLower(service.Name),
				startMode,
			)
		}
	}
	return nil, nil
}
