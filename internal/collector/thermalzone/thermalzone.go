//go:build windows

package thermalzone

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "thermalzone"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_Counters_ThermalZoneInformation metrics.
type Collector struct {
	config    Config
	miSession *mi.Session
	miQuery   mi.Query

	percentPassiveLimit *prometheus.Desc
	temperature         *prometheus.Desc
	throttleReasons     *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, miSession *mi.Session) error {
	if miSession == nil {
		return errors.New("miSession is nil")
	}

	miQuery, err := mi.NewQuery("SELECT Name, HighPrecisionTemperature, PercentPassiveLimit, ThrottleReasons FROM Win32_PerfRawData_Counters_ThermalZoneInformation")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.miQuery = miQuery
	c.miSession = miSession

	c.temperature = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "temperature_celsius"),
		"(Temperature)",
		[]string{
			"name",
		},
		nil,
	)
	c.percentPassiveLimit = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "percent_passive_limit"),
		"(PercentPassiveLimit)",
		[]string{
			"name",
		},
		nil,
	)
	c.throttleReasons = prometheus.NewDesc(
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
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		return fmt.Errorf("failed collecting thermalzone metrics: %w", err)
	}

	return nil
}

// Win32_PerfRawData_Counters_ThermalZoneInformation docs:
// https://wutils.com/wmi/root/cimv2/win32_perfrawdata_counters_thermalzoneinformation/
type Win32_PerfRawData_Counters_ThermalZoneInformation struct {
	Name                     string `mi:"Name"`
	HighPrecisionTemperature uint32 `mi:"HighPrecisionTemperature"`
	PercentPassiveLimit      uint32 `mi:"PercentPassiveLimit"`
	ThrottleReasons          uint32 `mi:"ThrottleReasons"`
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_Counters_ThermalZoneInformation
	if err := c.miSession.Query(&dst, mi.NamespaceRootCIMv2, c.miQuery); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
	}

	for _, info := range dst {
		// Divide by 10 and subtract 273.15 to convert decikelvin to celsius
		ch <- prometheus.MustNewConstMetric(
			c.temperature,
			prometheus.GaugeValue,
			(float64(info.HighPrecisionTemperature)/10.0)-273.15,
			info.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.percentPassiveLimit,
			prometheus.GaugeValue,
			float64(info.PercentPassiveLimit),
			info.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.throttleReasons,
			prometheus.GaugeValue,
			float64(info.ThrottleReasons),
			info.Name,
		)
	}

	return nil
}
