//go:build windows

package license

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus-community/windows_exporter/pkg/headers/slc"
	"github.com/prometheus-community/windows_exporter/pkg/types"
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

// A collector is a Prometheus collector for WMI Win32_PerfRawData_DNS_DNS metrics
type collector struct {
	logger log.Logger

	LicenseStatus *prometheus.Desc
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
	c.LicenseStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "status"),
		"Status of windows license",
		[]string{"state"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting license metrics", "err", err)
		return err
	}
	return nil
}

func (c *collector) collect(ch chan<- prometheus.Metric) error {
	status, err := slc.SLIsWindowsGenuineLocal()
	if err != nil {
		return err
	}

	for k, v := range labelMap {
		val := 0.0
		if status == k {
			val = 1.0
		}

		ch <- prometheus.MustNewConstMetric(c.LicenseStatus, prometheus.GaugeValue, val, v)
	}

	return nil
}
