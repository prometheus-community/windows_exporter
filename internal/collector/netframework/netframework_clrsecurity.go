// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package netframework

import (
	"fmt"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/labelcache"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Collector) describeClrSecurity() {
	c.numberLinkTimeChecks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrSecurity+"_link_time_checks_total"),
		"Displays the total number of link-time code access security checks since the application started.",
		[]string{"process", "process_id"},
		nil,
	)
	c.timeInRTChecks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrSecurity+"_rt_checks_time_ratio"),
		"Displays the ratio (0-1) of time spent performing runtime code access security checks in the last sample.",
		[]string{"process", "process_id"},
		nil,
	)
	c.stackWalkDepth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrSecurity+"_stack_walk_depth"),
		"Displays the depth of the stack during that last runtime code access security check.",
		[]string{"process", "process_id"},
		nil,
	)
	c.totalRuntimeChecks = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrSecurity+"_runtime_checks_total"),
		"Displays the total number of runtime code access security checks performed since the application started.",
		[]string{"process", "process_id"},
		nil,
	)
}

func (c *Collector) buildClrSecurity() error {
	c.describeClrSecurity()

	var err error

	c.perfClrSecurity, err = newPerfCollector[perfDataClrSecurity](c.logger, c.perfQuery, ".NET CLR Security")

	return err
}

type perfDataClrSecurity struct {
	Name                      string
	Occurrence                int
	NumberLinkTimeChecks      float64 `perfdata:"# Link Time Checks"`
	PercentTimeinRTchecks     float64 `perfdata:"% Time in RT checks"`
	PercentTimeinRTchecksBase float64 `perfdata:"% Time in RT checks,secondvalue"`
	StackWalkDepth            float64 `perfdata:"Stack Walk Depth"`
	TotalRuntimeChecks        float64 `perfdata:"Total Runtime Checks"`
}

func (p *perfDataClrSecurity) instance() (*string, int) { return &p.Name, p.Occurrence }

func (c *Collector) collectClrSecurity(ch chan<- prometheus.Metric, scrape *labelcache.Scrape, _ time.Duration) error {
	var dst []perfDataClrSecurity
	if err := c.perfClrSecurity.Collect(&dst); err != nil {
		return fmt.Errorf("failed to collect .NET CLR Security: %w", err)
	}

	dst, processIDs := prepareProcesses(c, dst)

	for i, process := range dst {
		ch <- scrape.MustNewConstMetric(
			c.numberLinkTimeChecks,
			prometheus.CounterValue,
			float64(process.NumberLinkTimeChecks),
			process.Name,
			processIDs[i],
		)

		if process.PercentTimeinRTchecksBase > 0 {
			ch <- scrape.MustNewConstMetric(
				c.timeInRTChecks,
				prometheus.GaugeValue,
				process.PercentTimeinRTchecks/process.PercentTimeinRTchecksBase,
				process.Name,
				processIDs[i],
			)
		}

		ch <- scrape.MustNewConstMetric(
			c.stackWalkDepth,
			prometheus.GaugeValue,
			float64(process.StackWalkDepth),
			process.Name,
			processIDs[i],
		)

		ch <- scrape.MustNewConstMetric(
			c.totalRuntimeChecks,
			prometheus.CounterValue,
			float64(process.TotalRuntimeChecks),
			process.Name,
			processIDs[i],
		)
	}

	return nil
}
