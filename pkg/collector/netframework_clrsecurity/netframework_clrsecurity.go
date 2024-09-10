//go:build windows

package netframework_clrsecurity

import (
	"errors"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "netframework_clrsecurity"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_NETFramework_NETCLRSecurity metrics.
type Collector struct {
	config    Config
	wmiClient *wmi.Client

	numberLinkTimeChecks *prometheus.Desc
	timeInRTChecks       *prometheus.Desc
	stackWalkDepth       *prometheus.Desc
	totalRuntimeChecks   *prometheus.Desc
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
	c.numberLinkTimeChecks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "link_time_checks_total"),
		"Displays the total number of link-time code access security checks since the application started.",
		[]string{"process"},
		nil,
	)
	c.timeInRTChecks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "rt_checks_time_percent"),
		"Displays the percentage of time spent performing runtime code access security checks in the last sample.",
		[]string{"process"},
		nil,
	)
	c.stackWalkDepth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "stack_walk_depth"),
		"Displays the depth of the stack during that last runtime code access security check.",
		[]string{"process"},
		nil,
	)
	c.totalRuntimeChecks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "runtime_checks_total"),
		"Displays the total number of runtime code access security checks performed since the application started.",
		[]string{"process"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	logger = logger.With(slog.String("collector", Name))
	if err := c.collect(ch); err != nil {
		logger.Error("failed collecting win32_perfrawdata_netframework_netclrsecurity metrics",
			slog.Any("err", err),
		)

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

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRSecurity
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_NETFramework_NETCLRSecurity", &dst); err != nil {
		return err
	}

	for _, process := range dst {
		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.numberLinkTimeChecks,
			prometheus.CounterValue,
			float64(process.NumberLinkTimeChecks),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.timeInRTChecks,
			prometheus.GaugeValue,
			float64(process.PercentTimeinRTchecks)/float64(process.Frequency_PerfTime),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.stackWalkDepth,
			prometheus.GaugeValue,
			float64(process.StackWalkDepth),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.totalRuntimeChecks,
			prometheus.CounterValue,
			float64(process.TotalRuntimeChecks),
			process.Name,
		)
	}

	return nil
}
