//go:build windows

package license

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/headers/slc"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "license"

var labelMap = map[slc.SL_GENUINE_STATE]string{
	slc.SL_GEN_STATE_IS_GENUINE:      "genuine",
	slc.SL_GEN_STATE_INVALID_LICENSE: "invalid_license",
	slc.SL_GEN_STATE_TAMPERED:        "tampered",
	slc.SL_GEN_STATE_OFFLINE:         "offline",
	slc.SL_GEN_STATE_LAST:            "last",
}

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_DNS_DNS metrics.
type Collector struct {
	config Config

	licenseStatus *prometheus.Desc
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

func (c *Collector) GetPerfCounter(_ log.Logger) ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(_ log.Logger) error {
	c.licenseStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "status"),
		"Status of windows license",
		[]string{"state"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	if err := c.collect(ch); err != nil {
		_ = level.Error(logger).Log("msg", "failed collecting license metrics", "err", err)
		return err
	}
	return nil
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	status, err := slc.SLIsWindowsGenuineLocal()
	if err != nil {
		return err
	}

	for k, v := range labelMap {
		val := 0.0
		if status == k {
			val = 1.0
		}

		ch <- prometheus.MustNewConstMetric(c.licenseStatus, prometheus.GaugeValue, val, v)
	}

	return nil
}
