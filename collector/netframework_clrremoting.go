//go:build windows
// +build windows

package collector

import (
	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	registerCollector("netframework_clrremoting", NewNETFramework_NETCLRRemotingCollector)
}

// A NETFramework_NETCLRRemotingCollector is a Prometheus collector for WMI Win32_PerfRawData_NETFramework_NETCLRRemoting metrics
type NETFramework_NETCLRRemotingCollector struct {
	Channels                  *prometheus.Desc
	ContextBoundClassesLoaded *prometheus.Desc
	ContextBoundObjects       *prometheus.Desc
	ContextProxies            *prometheus.Desc
	Contexts                  *prometheus.Desc
	TotalRemoteCalls          *prometheus.Desc
}

// NewNETFramework_NETCLRRemotingCollector ...
func NewNETFramework_NETCLRRemotingCollector() (Collector, error) {
	const subsystem = "netframework_clrremoting"
	return &NETFramework_NETCLRRemotingCollector{
		Channels: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "channels_total"),
			"Displays the total number of remoting channels registered across all application domains since application started.",
			[]string{"process"},
			nil,
		),
		ContextBoundClassesLoaded: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "context_bound_classes_loaded"),
			"Displays the current number of context-bound classes that are loaded.",
			[]string{"process"},
			nil,
		),
		ContextBoundObjects: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "context_bound_objects_total"),
			"Displays the total number of context-bound objects allocated.",
			[]string{"process"},
			nil,
		),
		ContextProxies: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "context_proxies_total"),
			"Displays the total number of remoting proxy objects in this process since it started.",
			[]string{"process"},
			nil,
		),
		Contexts: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "contexts"),
			"Displays the current number of remoting contexts in the application.",
			[]string{"process"},
			nil,
		),
		TotalRemoteCalls: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "remote_calls_total"),
			"Displays the total number of remote procedure calls invoked since the application started.",
			[]string{"process"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NETFramework_NETCLRRemotingCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Error("failed collecting win32_perfrawdata_netframework_netclrremoting metrics:", desc, err)
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

func (c *NETFramework_NETCLRRemotingCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_NETFramework_NETCLRRemoting
	q := queryAll(&dst)
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	for _, process := range dst {

		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.Channels,
			prometheus.CounterValue,
			float64(process.Channels),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ContextBoundClassesLoaded,
			prometheus.GaugeValue,
			float64(process.ContextBoundClassesLoaded),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ContextBoundObjects,
			prometheus.CounterValue,
			float64(process.ContextBoundObjectsAllocPersec),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ContextProxies,
			prometheus.CounterValue,
			float64(process.ContextProxies),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Contexts,
			prometheus.GaugeValue,
			float64(process.Contexts),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalRemoteCalls,
			prometheus.CounterValue,
			float64(process.TotalRemoteCalls),
			process.Name,
		)
	}

	return nil, nil
}
