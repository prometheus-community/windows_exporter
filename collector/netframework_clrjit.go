// +build windows

package collector

import (
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	registerCollector("netframework_clrjit", NewNETFrameworkCLRJitCollector, ".NET CLR Jit")
}

// A NETFrameworkCLRJitCollector is a Prometheus collector for Perflib .NET CLR Jit metrics
type NETFrameworkCLRJitCollector struct {
	NumberofMethodsJitted      *prometheus.Desc
	TimeinJit                  *prometheus.Desc
	StandardJitFailures        *prometheus.Desc
	TotalNumberofILBytesJitted *prometheus.Desc

	processWhitelistPattern *regexp.Regexp
	processBlacklistPattern *regexp.Regexp
}

// NewNETFrameworkCLRJitCollector ...
func NewNETFrameworkCLRJitCollector() (Collector, error) {
	const subsystem = "netframework_clrjit"
	commonFlags := GetNETFrameworkFlags()
	return &NETFrameworkCLRJitCollector{
		NumberofMethodsJitted: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "jit_methods_total"),
			"Displays the total number of methods JIT-compiled since the application started. This counter does not include pre-JIT-compiled methods.",
			[]string{"process"},
			nil,
		),
		TimeinJit: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "jit_time_percent"),
			"Displays the percentage of time spent in JIT compilation. This counter is updated at the end of every JIT compilation phase. A JIT compilation phase occurs when a method and its dependencies are compiled.",
			[]string{"process"},
			nil,
		),
		StandardJitFailures: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "jit_standard_failures_total"),
			"Displays the peak number of methods the JIT compiler has failed to compile since the application started. This failure can occur if the MSIL cannot be verified or if there is an internal error in the JIT compiler.",
			[]string{"process"},
			nil,
		),
		TotalNumberofILBytesJitted: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "jit_il_bytes_total"),
			"Displays the total number of Microsoft intermediate language (MSIL) bytes compiled by the just-in-time (JIT) compiler since the application started",
			[]string{"process"},
			nil,
		),
		processWhitelistPattern: commonFlags.whitelistRegexp,
		processBlacklistPattern: commonFlags.blacklistRegexp,
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NETFrameworkCLRJitCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting netframework_clrjit metrics:", desc, err)
		return err
	}
	return nil
}

type netframeworkCLRJit struct {
	Name string

	Frequency_PerfTime         float64 `perflib:"Not Displayed_Base"`
	ILBytesJittedPersec        float64 `perflib:"IL Bytes Jitted / sec"`
	NumberofILBytesJitted      float64 `perflib:"# of IL Bytes Jitted"`
	NumberofMethodsJitted      float64 `perflib:"# of Methods Jitted"`
	PercentTimeinJit           float64 `perflib:"% Time in Jit"`
	StandardJitFailures        float64 `perflib:"Standard Jit Failures"`
	TotalNumberofILBytesJitted float64 `perflib:"Total # of IL Bytes Jitted"`
}

func (c *NETFrameworkCLRJitCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []netframeworkCLRJit

	if err := unmarshalObject(ctx.perfObjects[".NET CLR Jit"], &dst); err != nil {
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
			name = fmt.Sprintf("%s#%d", name, procnum)
			names[name]++
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
			c.NumberofMethodsJitted,
			prometheus.CounterValue,
			process.NumberofMethodsJitted,
			name,
		)

		timeInJit := 0.0
		if process.Frequency_PerfTime != 0 {
			timeInJit = process.PercentTimeinJit / process.Frequency_PerfTime
		}
		ch <- prometheus.MustNewConstMetric(
			c.TimeinJit,
			prometheus.GaugeValue,
			timeInJit,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.StandardJitFailures,
			prometheus.GaugeValue,
			process.StandardJitFailures,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalNumberofILBytesJitted,
			prometheus.CounterValue,
			process.TotalNumberofILBytesJitted,
			name,
		)
	}

	return nil, nil
}
