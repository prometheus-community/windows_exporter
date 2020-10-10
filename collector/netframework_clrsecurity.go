// +build windows

package collector

import (
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	registerCollector("netframework_clrsecurity", NewNETFrameworkCLRSecurityCollector, ".NET CLR Security")
}

// A NETFrameworkCLRSecurityCollector is a Prometheus collector for Perflib .NET CLR Security metrics
type NETFrameworkCLRSecurityCollector struct {
	NumberLinkTimeChecks *prometheus.Desc
	TimeinRTchecks       *prometheus.Desc
	StackWalkDepth       *prometheus.Desc
	TotalRuntimeChecks   *prometheus.Desc

	processWhitelistPattern *regexp.Regexp
	processBlacklistPattern *regexp.Regexp
}

// NewNETFrameworkCLRSecurityCollector ...
func NewNETFrameworkCLRSecurityCollector() (Collector, error) {
	const subsystem = "netframework_clrsecurity"
	commonFlags := GetNETFrameworkFlags()
	return &NETFrameworkCLRSecurityCollector{
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
		processWhitelistPattern: commonFlags.whitelistRegexp,
		processBlacklistPattern: commonFlags.blacklistRegexp,
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NETFrameworkCLRSecurityCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting netframework_clrsecurity metrics:", desc, err)
		return err
	}
	return nil
}

type netframeworkCLRSecurity struct {
	Name string

	Frequency_PerfTime           float64 `perflib:"Not Displayed_Base"`
	NumberLinkTimeChecks         float64 `perflib:"# Link Time Checks"`
	PercentTimeinRTchecks        float64 `perflib:"% Time in RT checks"`
	PercentTimeSigAuthenticating float64 `perflib:"% Time Sig. Authenticating"`
	StackWalkDepth               float64 `perflib:"Stack Walk Depth"`
	TotalRuntimeChecks           float64 `perflib:"Total Runtime Checks"`
}

func (c *NETFrameworkCLRSecurityCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []netframeworkCLRSecurity

	if err := unmarshalObject(ctx.perfObjects[".NET CLR Security"], &dst); err != nil {
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
			c.NumberLinkTimeChecks,
			prometheus.CounterValue,
			process.NumberLinkTimeChecks,
			name,
		)

		timeinRTchecks := 0.0
		if process.Frequency_PerfTime != 0 {
			timeinRTchecks = process.PercentTimeinRTchecks / process.Frequency_PerfTime
		}
		ch <- prometheus.MustNewConstMetric(
			c.TimeinRTchecks,
			prometheus.GaugeValue,
			timeinRTchecks,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.StackWalkDepth,
			prometheus.GaugeValue,
			process.StackWalkDepth,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalRuntimeChecks,
			prometheus.CounterValue,
			process.TotalRuntimeChecks,
			name,
		)
	}

	return nil, nil
}
