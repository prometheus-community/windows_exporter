// +build windows

package collector

import (
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	registerCollector("netframework_clrexceptions", NewNETFrameworkCLRExceptionsCollector, ".NET CLR Exceptions")
}

// A NETFrameworkCLRExceptionsCollector is a Prometheus collector for Perflib .NET CLR Exceptions metrics
type NETFrameworkCLRExceptionsCollector struct {
	NumberofExcepsThrown *prometheus.Desc
	NumberofFilters      *prometheus.Desc
	NumberofFinallys     *prometheus.Desc
	ThrowToCatchDepth    *prometheus.Desc

	processWhitelistPattern *regexp.Regexp
	processBlacklistPattern *regexp.Regexp
}

// NewNETFrameworkCLRExceptionsCollector ...
func NewNETFrameworkCLRExceptionsCollector() (Collector, error) {
	const subsystem = "netframework_clrexceptions"
	commonFlags := GetNETFrameworkFlags()
	return &NETFrameworkCLRExceptionsCollector{
		NumberofExcepsThrown: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "exceptions_thrown_total"),
			"Displays the total number of exceptions thrown since the application started. This includes both .NET exceptions and unmanaged exceptions that are converted into .NET exceptions.",
			[]string{"process"},
			nil,
		),
		NumberofFilters: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "exceptions_filters_total"),
			"Displays the total number of .NET exception filters executed. An exception filter evaluates regardless of whether an exception is handled.",
			[]string{"process"},
			nil,
		),
		NumberofFinallys: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "exceptions_finallys_total"),
			"Displays the total number of finally blocks executed. Only the finally blocks executed for an exception are counted; finally blocks on normal code paths are not counted by this counter.",
			[]string{"process"},
			nil,
		),
		ThrowToCatchDepth: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "throw_to_catch_depth_total"),
			"Displays the total number of stack frames traversed, from the frame that threw the exception to the frame that handled the exception.",
			[]string{"process"},
			nil,
		),
		processWhitelistPattern: commonFlags.whitelistRegexp,
		processBlacklistPattern: commonFlags.blacklistRegexp,
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *NETFrameworkCLRExceptionsCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting netframework_clrexceptions metrics:", desc, err)
		return err
	}
	return nil
}

type netframeworkCLRExceptions struct {
	Name string

	NumberofExcepsThrown       float64 `perflib:"# of Exceps Thrown"`
	NumberofExcepsThrownPersec float64 `perflib:"# of Exceps Thrown / sec"`
	NumberofFiltersPersec      float64 `perflib:"# of Filters / sec"`
	NumberofFinallysPersec     float64 `perflib:"# of Finallys / sec"`
	ThrowToCatchDepthPersec    float64 `perflib:"Throw To Catch Depth / sec"`
}

func (c *NETFrameworkCLRExceptionsCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []netframeworkCLRExceptions

	if err := unmarshalObject(ctx.perfObjects[".NET CLR Exceptions"], &dst); err != nil {
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
			c.NumberofExcepsThrown,
			prometheus.CounterValue,
			process.NumberofExcepsThrown,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofFilters,
			prometheus.CounterValue,
			process.NumberofFiltersPersec,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.NumberofFinallys,
			prometheus.CounterValue,
			process.NumberofFinallysPersec,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ThrowToCatchDepth,
			prometheus.CounterValue,
			process.ThrowToCatchDepthPersec,
			name,
		)
	}

	return nil, nil
}
