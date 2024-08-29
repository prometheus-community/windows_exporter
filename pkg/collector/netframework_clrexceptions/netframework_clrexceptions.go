//go:build windows

package netframework_clrexceptions

import (
	"errors"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "netframework_clrexceptions"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_NETFramework_NETCLRExceptions metrics.
type Collector struct {
	config    Config
	wmiClient *wmi.Client

	numberOfExceptionsThrown *prometheus.Desc
	numberOfFilters          *prometheus.Desc
	numberOfFinally          *prometheus.Desc
	throwToCatchDepth        *prometheus.Desc
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

func (c *Collector) Build(_ log.Logger, wmiClient *wmi.Client) error {
	if wmiClient == nil || wmiClient.SWbemServicesClient == nil {
		return errors.New("wmiClient or SWbemServicesClient is nil")
	}

	c.wmiClient = wmiClient
	c.numberOfExceptionsThrown = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "exceptions_thrown_total"),
		"Displays the total number of exceptions thrown since the application started. This includes both .NET exceptions and unmanaged exceptions that are converted into .NET exceptions.",
		[]string{"process"},
		nil,
	)
	c.numberOfFilters = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "exceptions_filters_total"),
		"Displays the total number of .NET exception filters executed. An exception filter evaluates regardless of whether an exception is handled.",
		[]string{"process"},
		nil,
	)
	c.numberOfFinally = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "exceptions_finallys_total"),
		"Displays the total number of finally blocks executed. Only the finally blocks executed for an exception are counted; finally blocks on normal code paths are not counted by this counter.",
		[]string{"process"},
		nil,
	)
	c.throwToCatchDepth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "throw_to_catch_depth_total"),
		"Displays the total number of stack frames traversed, from the frame that threw the exception to the frame that handled the exception.",
		[]string{"process"},
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	if err := c.collect(ch); err != nil {
		_ = level.Error(logger).Log("msg", "failed collecting win32_perfrawdata_netframework_netclrexceptions metrics", "err", err)
		return err
	}
	return nil
}

type Win32_PerfRawData_NETFramework_NETCLRExceptions struct {
	Name string

	NumberofExcepsThrown       uint32
	NumberofExcepsThrownPersec uint32
	NumberofFiltersPersec      uint32
	NumberofFinallysPersec     uint32
	ThrowToCatchDepthPersec    uint32
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRExceptions
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_NETFramework_NETCLRExceptions", &dst); err != nil {
		return err
	}

	for _, process := range dst {
		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.numberOfExceptionsThrown,
			prometheus.CounterValue,
			float64(process.NumberofExcepsThrown),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfFilters,
			prometheus.CounterValue,
			float64(process.NumberofFiltersPersec),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfFinally,
			prometheus.CounterValue,
			float64(process.NumberofFinallysPersec),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.throwToCatchDepth,
			prometheus.CounterValue,
			float64(process.ThrowToCatchDepthPersec),
			process.Name,
		)
	}

	return nil
}
