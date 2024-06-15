//go:build windows

package netframework_clrinterop

import (
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "netframework_clrinterop"

type Config struct{}

var ConfigDefaults = Config{}

// A collector is a Prometheus collector for WMI Win32_PerfRawData_NETFramework_NETCLRInterop metrics
type collector struct {
	logger log.Logger

	NumberofCCWs        *prometheus.Desc
	Numberofmarshalling *prometheus.Desc
	NumberofStubs       *prometheus.Desc
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
	c.NumberofCCWs = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "com_callable_wrappers_total"),
		"Displays the current number of COM callable wrappers (CCWs). A CCW is a proxy for a managed object being referenced from an unmanaged COM client.",
		[]string{"process"},
		nil,
	)
	c.Numberofmarshalling = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "interop_marshalling_total"),
		"Displays the total number of times arguments and return values have been marshaled from managed to unmanaged code, and vice versa, since the application started.",
		[]string{"process"},
		nil,
	)
	c.NumberofStubs = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "interop_stubs_created_total"),
		"Displays the current number of stubs created by the common language runtime. Stubs are responsible for marshaling arguments and return values from managed to unmanaged code, and vice versa, during a COM interop call or a platform invoke call.",
		[]string{"process"},
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting win32_perfrawdata_netframework_netclrinterop metrics", "err", err)
		return err
	}
	return nil
}

type Win32_PerfRawData_NETFramework_NETCLRInterop struct {
	Name string

	NumberofCCWs             uint32
	Numberofmarshalling      uint32
	NumberofStubs            uint32
	NumberofTLBexportsPersec uint32
	NumberofTLBimportsPersec uint32
}

func (c *collector) collect(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRInterop
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}

	for _, process := range dst {

		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.NumberofCCWs,
			prometheus.CounterValue,
			float64(process.NumberofCCWs),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Numberofmarshalling,
			prometheus.CounterValue,
			float64(process.Numberofmarshalling),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofStubs,
			prometheus.CounterValue,
			float64(process.NumberofStubs),
			process.Name,
		)
	}

	return nil
}
