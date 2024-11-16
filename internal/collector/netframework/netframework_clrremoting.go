//go:build windows

package netframework

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Collector) buildClrRemoting() {
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
}

type Win32_PerfRawData_NETFramework_NETCLRRemoting struct {
	Name string `mi:"Name"`

	Channels                       uint32 `mi:"Channels"`
	ContextBoundClassesLoaded      uint32 `mi:"ContextBoundClassesLoaded"`
	ContextBoundObjectsAllocPersec uint32 `mi:"ContextBoundObjectsAllocPersec"`
	ContextProxies                 uint32 `mi:"ContextProxies"`
	Contexts                       uint32 `mi:"Contexts"`
	RemoteCallsPersec              uint32 `mi:"RemoteCallsPersec"`
	TotalRemoteCalls               uint32 `mi:"TotalRemoteCalls"`
}

func (c *Collector) collectClrRemoting(ch chan<- prometheus.Metric) error {
	var dst []Win32_PerfRawData_NETFramework_NETCLRRemoting
	if err := c.miSession.Query(&dst, mi.NamespaceRootCIMv2, utils.Must(mi.NewQuery("SELECT * Win32_PerfRawData_NETFramework_NETCLRRemoting"))); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
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
