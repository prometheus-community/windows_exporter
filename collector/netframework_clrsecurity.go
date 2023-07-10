//go:build windows
// +build windows

package collector

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

// A NETFramework_NETCLRSecurityCollector is a Prometheus collector for WMI Win32_PerfRawData_NETFramework_NETCLRSecurity metrics
type NETFramework_NETCLRSecurityCollector struct {
	logger log.Logger

	NumberLinkTimeChecks *prometheus.Desc
	TimeinRTchecks       *prometheus.Desc
	StackWalkDepth       *prometheus.Desc
	TotalRuntimeChecks   *prometheus.Desc
}

// newNETFramework_NETCLRSecurityCollector ...
func newNETFramework_NETCLRSecurityCollector(logger log.Logger) (Collector, error) {
	const subsystem = "netframework_clrsecurity"
	return &NETFramework_NETCLRSecurityCollector{
		logger: log.With(logger, "collector", subsystem),
		NumberLinkTimeChecks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "link_time_checks_total"),
			"Displays the total number of link-time code access security checks since the application started.",
			[]string{"process"},
			nil,
		),
		TimeinRTchecks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "rt_checks_time_percent"),
			"Displays the percentage of time spent performing runtime code access security checks in the last sample.",
			[]string{"process"},
			nil,
		),
		StackWalkDepth: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "stack_walk_depth"),
			"Displays the depth of the stack during that last runtime code access security check.",
			[]string{"process"},
			nil,
		),
		TotalRuntimeChecks: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "runtime_checks_total"),
			"Displays the total number of runtime code access security checks performed since the application started.",
			[]string{"process"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NETFramework_NETCLRSecurityCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("failed collecting win32_perfrawdata_netframework_netclrsecurity metrics", "desc", desc, "err", err)
		return err
	}
	return nil
}

type Win32_PerfRawData_NETFramework_NETCLRSecurity struct {
	Name string

	Frequency_PerfTime           uint32
	NumberLinkTimeChecks         uint32
	PercentTimeinRTchecks        uint32
	PercentTimeSigAuthenticating uint64
	StackWalkDepth               uint32
	TotalRuntimeChecks           uint32
}

func (c *NETFramework_NETCLRSecurityCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_NETFramework_NETCLRSecurity
	q := queryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, process := range dst {

		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.NumberLinkTimeChecks,
			prometheus.CounterValue,
			float64(process.NumberLinkTimeChecks),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TimeinRTchecks,
			prometheus.GaugeValue,
			float64(process.PercentTimeinRTchecks)/float64(process.Frequency_PerfTime),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.StackWalkDepth,
			prometheus.GaugeValue,
			float64(process.StackWalkDepth),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalRuntimeChecks,
			prometheus.CounterValue,
			float64(process.TotalRuntimeChecks),
			process.Name,
		)
	}

	return nil, nil
}
