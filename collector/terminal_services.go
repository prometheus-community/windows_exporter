// +build windows

package collector

import (
	"errors"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	registerCollector("terminal_services", NewTerminalServicesCollector)
}

// A TerminalServicesCollector is a Prometheus collector for WMI
// Win32_PerfRawData_LocalSessionManager_TerminalServices &  Win32_PerfRawData_TermService_TerminalServicesSession  metrics
type TerminalServicesCollector struct {
	Local_session_count *prometheus.Desc
}

// NewTerminalServicesCollector ...
func NewTerminalServicesCollector() (Collector, error) {
	const subsystem = "terminal_services"
	return &TerminalServicesCollector{
		Local_session_count: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "local_session_count"),
			"Number of Terminal Services sessions",
			[]string{"session"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *TerminalServicesCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collectTSSessionCount(ch); err != nil {
		log.Error("failed collecting terminal services session count metrics:", desc, err)
		return err
	}
	return nil
}

type Win32_PerfRawData_LocalSessionManager_TerminalServices struct {
	ActiveSessions   uint32
	InactiveSessions uint32
	TotalSessions    uint32
}

func (c *TerminalServicesCollector) collectTSSessionCount(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_LocalSessionManager_TerminalServices
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	if len(dst) == 0 {
		return nil, errors.New("WMI query returned empty result set")
	}

	ch <- prometheus.MustNewConstMetric(
		c.Local_session_count,
		prometheus.GaugeValue,
		float64(dst[0].ActiveSessions),
		"active",
	)

	ch <- prometheus.MustNewConstMetric(
		c.Local_session_count,
		prometheus.GaugeValue,
		float64(dst[0].InactiveSessions),
		"inactive",
	)

	ch <- prometheus.MustNewConstMetric(
		c.Local_session_count,
		prometheus.GaugeValue,
		float64(dst[0].TotalSessions),
		"total",
	)

	return nil, nil
}
