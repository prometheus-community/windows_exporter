//go:build windows

package thermalzone

import (
	"errors"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "thermalzone"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_Counters_ThermalZoneInformation metrics.
type Collector struct {
	config    Config
	wmiClient *wmi.Client

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

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, wmiClient *wmi.Client) error {
	if wmiClient == nil || wmiClient.SWbemServicesClient == nil {
		return errors.New("wmiClient or SWbemServicesClient is nil")
	}

	c.wmiClient = wmiClient
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
func (c *Collector) Collect(_ *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))
	if err := c.collect(ch); err != nil {
		logger.Error("failed collecting thermalzone metrics",
			slog.Any("err", err),
		)

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

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_Counters_ThermalZoneInformation
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_Counters_ThermalZoneInformation", &dst); err != nil {
		return err
	}

	// ThermalZone collector has been known to 'successfully' return an empty result.
	if len(dst) == 0 {
		return errors.New("empty results set for collector")
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
