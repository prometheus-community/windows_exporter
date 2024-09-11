//go:build windows

package netframework_clrremoting

import (
	"errors"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "netframework_clrremoting"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_NETFramework_NETCLRRemoting metrics.
type Collector struct {
	config    Config
	wmiClient *wmi.Client

	channels                  *prometheus.Desc
	contextBoundClassesLoaded *prometheus.Desc
	contextBoundObjects       *prometheus.Desc
	contextProxies            *prometheus.Desc
	contexts                  *prometheus.Desc
	totalRemoteCalls          *prometheus.Desc
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

	c.channels = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "channels_total"),
		"Displays the total number of remoting channels registered across all application domains since application started.",
		[]string{"process"},
		nil,
	)
	c.contextBoundClassesLoaded = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "context_bound_classes_loaded"),
		"Displays the current number of context-bound classes that are loaded.",
		[]string{"process"},
		nil,
	)
	c.contextBoundObjects = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "context_bound_objects_total"),
		"Displays the total number of context-bound objects allocated.",
		[]string{"process"},
		nil,
	)
	c.contextProxies = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "context_proxies_total"),
		"Displays the total number of remoting proxy objects in this process since it started.",
		[]string{"process"},
		nil,
	)
	c.contexts = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "contexts"),
		"Displays the current number of remoting contexts in the application.",
		[]string{"process"},
		nil,
	)
	c.totalRemoteCalls = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "remote_calls_total"),
		"Displays the total number of remote procedure calls invoked since the application started.",
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
		logger.Error("failed collecting win32_perfrawdata_netframework_netclrremoting metrics",
			slog.Any("err", err),
		)

		return err
	}

	return nil
}

type Win32_PerfRawData_NETFramework_NETCLRRemoting struct {
	Name string

	Channels                       uint32
	ContextBoundClassesLoaded      uint32
	ContextBoundObjectsAllocPersec uint32
	ContextProxies                 uint32
	Contexts                       uint32
	RemoteCallsPersec              uint32
	TotalRemoteCalls               uint32
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRRemoting
	if err := c.wmiClient.Query("SELECT * FROM Win32_PerfRawData_NETFramework_NETCLRRemoting", &dst); err != nil {
		return err
	}

	for _, process := range dst {
		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.channels,
			prometheus.CounterValue,
			float64(process.Channels),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.contextBoundClassesLoaded,
			prometheus.GaugeValue,
			float64(process.ContextBoundClassesLoaded),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.contextBoundObjects,
			prometheus.CounterValue,
			float64(process.ContextBoundObjectsAllocPersec),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.contextProxies,
			prometheus.CounterValue,
			float64(process.ContextProxies),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.contexts,
			prometheus.GaugeValue,
			float64(process.Contexts),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.totalRemoteCalls,
			prometheus.CounterValue,
			float64(process.TotalRemoteCalls),
			process.Name,
		)
	}

	return nil
}
