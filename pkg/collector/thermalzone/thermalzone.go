//go:build windows

package thermalzone

import (
	"errors"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "thermalzone"

type Config struct{}

var ConfigDefaults = Config{}

// A collector is a Prometheus collector for WMI Win32_PerfRawData_Counters_ThermalZoneInformation metrics
type collector struct {
	logger log.Logger

	PercentPassiveLimit *prometheus.Desc
	Temperature         *prometheus.Desc
	ThrottleReasons     *prometheus.Desc
}

func New(logger log.Logger, _ *Config) types.Collector {
	c := &collector{}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(_ *kingpin.Application) types.Collector {
	return &collector{}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *collector) Build() error {
	c.Temperature = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "temperature_celsius"),
		"(Temperature)",
		[]string{
			"name",
		},
		nil,
	)
	c.PercentPassiveLimit = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "percent_passive_limit"),
		"(PercentPassiveLimit)",
		[]string{
			"name",
		},
		nil,
	)
	c.ThrottleReasons = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "throttle_reasons"),
		"(ThrottleReasons)",
		[]string{
			"name",
		},
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting thermalzone metrics", "desc", desc, "err", err)
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

func (c *collector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_Counters_ThermalZoneInformation
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	// ThermalZone collector has been known to 'successfully' return an empty result.
	if len(dst) == 0 {
		return nil, errors.New("Empty results set for collector")
	}

	for _, info := range dst {
		// Divide by 10 and subtract 273.15 to convert decikelvin to celsius
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
