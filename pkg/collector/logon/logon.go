//go:build windows

package logon

import (
	"errors"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "logon"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI metrics.
type Collector struct {
	config    Config
	wmiClient *wmi.Client

	logonType *prometheus.Desc
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
	c.logonType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "logon_type"),
		"Number of active logon sessions (LogonSession.LogonType)",
		[]string{"status"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))
	if err := c.collect(ch); err != nil {
		logger.Error("failed collecting user metrics",
			slog.Any("err", err),
		)

		return err
	}

	return nil
}

// Win32_LogonSession docs:
// - https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-logonsession
type Win32_LogonSession struct {
	LogonType uint32
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []Win32_LogonSession
	if err := c.wmiClient.Query("SELECT * FROM Win32_LogonSession", &dst); err != nil {
		return err
	}

	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
	}

	// Init counters
	system := 0
	interactive := 0
	network := 0
	batch := 0
	service := 0
	proxy := 0
	unlock := 0
	networkcleartext := 0
	newcredentials := 0
	remoteinteractive := 0
	cachedinteractive := 0
	cachedremoteinteractive := 0
	cachedunlock := 0

	for _, entry := range dst {
		switch entry.LogonType {
		case 0:
			system++
		case 2:
			interactive++
		case 3:
			network++
		case 4:
			batch++
		case 5:
			service++
		case 6:
			proxy++
		case 7:
			unlock++
		case 8:
			networkcleartext++
		case 9:
			newcredentials++
		case 10:
			remoteinteractive++
		case 11:
			cachedinteractive++
		case 12:
			cachedremoteinteractive++
		case 13:
			cachedunlock++
		}
	}

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(system),
		"system",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(interactive),
		"interactive",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(network),
		"network",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(batch),
		"batch",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(service),
		"service",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(proxy),
		"proxy",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(unlock),
		"unlock",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(networkcleartext),
		"network_clear_text",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(newcredentials),
		"new_credentials",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(remoteinteractive),
		"remote_interactive",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(cachedinteractive),
		"cached_interactive",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(remoteinteractive),
		"cached_remote_interactive",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(cachedunlock),
		"cached_unlock",
	)

	return nil
}
