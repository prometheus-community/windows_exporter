// +build windows

package collector

import (
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	registerCollector("netframework_clrinterop", NewNETFrameworkCLRInteropCollector, ".NET CLR Interop")
}

// A NETFrameworkCLRInteropCollector is a Prometheus collector for Perflib .NET CLR Interop metrics
type NETFrameworkCLRInteropCollector struct {
	NumberofCCWs        *prometheus.Desc
	Numberofmarshalling *prometheus.Desc
	NumberofStubs       *prometheus.Desc

	processWhitelistPattern *regexp.Regexp
	processBlacklistPattern *regexp.Regexp
}

// NewNETFrameworkCLRInteropCollector ...
func NewNETFrameworkCLRInteropCollector() (Collector, error) {
	const subsystem = "netframework_clrinterop"
	commonFlags := GetNETFrameworkFlags()
	return &NETFrameworkCLRInteropCollector{
		NumberofCCWs: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "com_callable_wrappers_total"),
			"Displays the current number of COM callable wrappers (CCWs). A CCW is a proxy for a managed object being referenced from an unmanaged COM client.",
			[]string{"process"},
			nil,
		),
		Numberofmarshalling: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "interop_marshalling_total"),
			"Displays the total number of times arguments and return values have been marshaled from managed to unmanaged code, and vice versa, since the application started.",
			[]string{"process"},
			nil,
		),
		NumberofStubs: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "interop_stubs_created_total"),
			"Displays the current number of stubs created by the common language runtime. Stubs are responsible for marshaling arguments and return values from managed to unmanaged code, and vice versa, during a COM interop call or a platform invoke call.",
			[]string{"process"},
			nil,
		),
		processWhitelistPattern: commonFlags.whitelistRegexp,
		processBlacklistPattern: commonFlags.blacklistRegexp,
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NETFrameworkCLRInteropCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting netframework_clrinterop metrics:", desc, err)
		return err
	}
	return nil
}

type netframeworkCLRInterop struct {
	Name string

	NumberofCCWs             float64 `perflib:"# of CCWs"`
	Numberofmarshalling      float64 `perflib:"# of marshalling"`
	NumberofStubs            float64 `perflib:"# of Stubs"`
	NumberofTLBexportsPersec float64 `perflib:"# of TLB exports / sec"`
	NumberofTLBimportsPersec float64 `perflib:"# of TLB imports / sec"`
}

func (c *NETFrameworkCLRInteropCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []netframeworkCLRInterop

	if err := unmarshalObject(ctx.perfObjects[".NET CLR Interop"], &dst); err != nil {
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
			c.NumberofCCWs,
			prometheus.CounterValue,
			process.NumberofCCWs,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Numberofmarshalling,
			prometheus.CounterValue,
			process.Numberofmarshalling,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofStubs,
			prometheus.CounterValue,
			process.NumberofStubs,
			name,
		)
	}

	return nil, nil
}
