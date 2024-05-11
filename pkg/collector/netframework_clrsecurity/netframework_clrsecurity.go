//go:build windows

package netframework_clrsecurity

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "netframework_clrsecurity"

type Config struct{}

var ConfigDefaults = Config{}

// A collector is a Prometheus collector for WMI Win32_PerfRawData_NETFramework_NETCLRSecurity metrics
type collector struct {
	logger log.Logger

	NumberLinkTimeChecks *prometheus.Desc
	TimeinRTchecks       *prometheus.Desc
	StackWalkDepth       *prometheus.Desc
	TotalRuntimeChecks   *prometheus.Desc
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
	c.NumberLinkTimeChecks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "link_time_checks_total"),
		"Displays the total number of link-time code access security checks since the application started.",
		[]string{"process"},
		nil,
	)
	c.TimeinRTchecks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rt_checks_time_percent"),
		"Displays the percentage of time spent performing runtime code access security checks in the last sample.",
		[]string{"process"},
		nil,
	)
	c.StackWalkDepth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "stack_walk_depth"),
		"Displays the depth of the stack during that last runtime code access security check.",
		[]string{"process"},
		nil,
	)
	c.TotalRuntimeChecks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "runtime_checks_total"),
		"Displays the total number of runtime code access security checks performed since the application started.",
		[]string{"process"},
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting win32_perfrawdata_netframework_netclrsecurity metrics", "err", err)
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

func (c *collector) collect(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRSecurity
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
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

	return nil
}
