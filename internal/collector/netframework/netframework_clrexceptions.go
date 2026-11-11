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

func (c *Collector) describeClrExceptions() {
	c.numberOfExceptionsThrown = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrExceptions+"_exceptions_thrown_total"),
		"Displays the total number of exceptions thrown since the application started. This includes both .NET exceptions and unmanaged exceptions that are converted into .NET exceptions.",
		[]string{"process", "process_id"},
		nil,
	)
	c.numberOfFilters = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrExceptions+"_exceptions_filters_total"),
		"Displays the total number of .NET exception filters executed. An exception filter evaluates regardless of whether an exception is handled.",
		[]string{"process", "process_id"},
		nil,
	)
	c.numberOfFinally = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrExceptions+"_exceptions_finallys_total"),
		"Displays the total number of finally blocks executed. Only the finally blocks executed for an exception are counted; finally blocks on normal code paths are not counted by this counter.",
		[]string{"process", "process_id"},
		nil,
	)
	c.throwToCatchDepth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrExceptions+"_throw_to_catch_depth_total"),
		"Displays the total number of stack frames traversed, from the frame that threw the exception to the frame that handled the exception.",
		[]string{"process", "process_id"},
		nil,
	)
}

func (c *Collector) buildClrExceptions() error {
	c.describeClrExceptions()

	var err error

	c.perfClrExceptions, err = newPerfCollector[perfDataClrExceptions](c.logger, c.perfQuery, ".NET CLR Exceptions")

	return err
}

type perfDataClrExceptions struct {
	Name                    string
	Occurrence              int
	NumberofExcepsThrown    float64 `perfdata:"# of Exceps Thrown"`
	NumberofFiltersPersec   float64 `perfdata:"# of Filters / sec"`
	NumberofFinallysPersec  float64 `perfdata:"# of Finallys / sec"`
	ThrowToCatchDepthPersec float64 `perfdata:"Throw To Catch Depth / sec"`
}

func (p *perfDataClrExceptions) instance() (*string, int) { return &p.Name, p.Occurrence }

func (c *Collector) collectClrExceptions(ch chan<- prometheus.Metric, scrape *labelcache.Scrape, _ time.Duration) error {
	var dst []perfDataClrExceptions
	if err := c.perfClrExceptions.Collect(&dst); err != nil {
		return fmt.Errorf("failed to collect .NET CLR Exceptions: %w", err)
	}

	dst, processIDs := prepareProcesses(c, dst)

	for i, process := range dst {
		ch <- scrape.MustNewConstMetric(
			c.numberOfExceptionsThrown,
			prometheus.CounterValue,
			float64(process.NumberofExcepsThrown),
			process.Name,
			processIDs[i],
		)

		ch <- scrape.MustNewConstMetric(
			c.numberOfFilters,
			prometheus.CounterValue,
			float64(process.NumberofFiltersPersec),
			process.Name,
			processIDs[i],
		)

		ch <- scrape.MustNewConstMetric(
			c.numberOfFinally,
			prometheus.CounterValue,
			float64(process.NumberofFinallysPersec),
			process.Name,
			processIDs[i],
		)

		ch <- scrape.MustNewConstMetric(
			c.throwToCatchDepth,
			prometheus.CounterValue,
			float64(process.ThrowToCatchDepthPersec),
			process.Name,
			processIDs[i],
		)
	}

	return nil
}
