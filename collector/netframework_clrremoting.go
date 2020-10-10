// +build windows

package collector

import (
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	registerCollector("netframework_clrremoting", NewNETFrameworkCLRRemotingCollector, ".NET CLR Remoting")
}

// A NETFrameworkCLRRemotingCollector is a Prometheus collector for Perflib .NET CLR Remoting metrics
type NETFrameworkCLRRemotingCollector struct {
	Channels                  *prometheus.Desc
	ContextBoundClassesLoaded *prometheus.Desc
	ContextBoundObjects       *prometheus.Desc
	ContextProxies            *prometheus.Desc
	Contexts                  *prometheus.Desc
	TotalRemoteCalls          *prometheus.Desc

	processWhitelistPattern *regexp.Regexp
	processBlacklistPattern *regexp.Regexp
}

// NewNETFrameworkCLRRemotingCollector ...
func NewNETFrameworkCLRRemotingCollector() (Collector, error) {
	const subsystem = "netframework_clrremoting"
	commonFlags := GetNETFrameworkFlags()
	return &NETFrameworkCLRRemotingCollector{
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
		processWhitelistPattern: commonFlags.whitelistRegexp,
		processBlacklistPattern: commonFlags.blacklistRegexp,
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NETFrameworkCLRRemotingCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting netframework_clrremoting metrics:", desc, err)
		return err
	}
	return nil
}

type netframeworkCLRRemoting struct {
	Name string

	Channels                       float64 `perflib:"Channels"`
	ContextBoundClassesLoaded      float64 `perflib:"Context-Bound Classes Loaded"`
	ContextBoundObjectsAllocPersec float64 `perflib:"Context-Bound Objects Alloc / sec"`
	ContextProxies                 float64 `perflib:"Context Proxies"`
	Contexts                       float64 `perflib:"Contexts"`
	RemoteCallsPersec              float64 `perflib:"Remote Calls/sec"`
	TotalRemoteCalls               float64 `perflib:"Total Remote Calls"`
}

func (c *NETFrameworkCLRRemotingCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []netframeworkCLRRemoting

	if err := unmarshalObject(ctx.perfObjects[".NET CLR Remoting"], &dst); err != nil {
		return nil, err
	}

	var names = make(map[string]int, len(dst))
	for _, process := range dst {

		if process.Name == "_Global_" {
			continue
		}

		// Append "#1", "#2", etc., to process names to disambiguate duplicates.
		name := process.Name
		procnum, exists := names[name]
		if exists {
			names[name]++
			name = fmt.Sprintf("%s#%d", name, procnum)
		} else {
			names[name] = 1
		}

		// The pattern matching against the whitelist and blacklist has to occur
		// after appending #N above to be consistent with other collectors.
		if c.processBlacklistPattern.MatchString(name) ||
			!c.processWhitelistPattern.MatchString(name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.Channels,
			prometheus.CounterValue,
			process.Channels,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ContextBoundClassesLoaded,
			prometheus.GaugeValue,
			process.ContextBoundClassesLoaded,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ContextBoundObjects,
			prometheus.CounterValue,
			process.ContextBoundObjectsAllocPersec,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ContextProxies,
			prometheus.CounterValue,
			process.ContextProxies,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Contexts,
			prometheus.GaugeValue,
			process.Contexts,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalRemoteCalls,
			prometheus.CounterValue,
			process.TotalRemoteCalls,
			name,
		)
	}

	return nil, nil
}
