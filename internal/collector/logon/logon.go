//go:build windows

package logon

import (
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/secur32"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "logon"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI metrics.
type Collector struct {
	config Config

	sessionInfo *prometheus.Desc
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

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	c.sessionInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "session_logon_timestamp_seconds"),
		"timestamp of the logon session in seconds.",
		[]string{"id", "username", "domain", "type"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	logonSessions, err := secur32.GetLogonSessions()
	if err != nil {
		return fmt.Errorf("failed to get logon sessions: %w", err)
	}

	for _, session := range logonSessions {
		ch <- prometheus.MustNewConstMetric(
			c.sessionInfo,
			prometheus.GaugeValue,
			float64(session.LogonTime.Unix()),
			session.LogonId.String(), session.UserName, session.LogonDomain, session.LogonType.String(),
		)
	}

	return nil
}
