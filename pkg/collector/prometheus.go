//go:build windows

package collector

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

// Interface guard.
var _ prometheus.Collector = (*Prometheus)(nil)

// Prometheus implements prometheus.Collector for a set of Windows MetricCollectors.
type Prometheus struct {
	maxScrapeDuration time.Duration
	logger            *slog.Logger
	metricCollectors  *MetricCollectors

	// Base metrics returned by Prometheus
	scrapeDurationDesc          *prometheus.Desc
	collectorScrapeDurationDesc *prometheus.Desc
	collectorScrapeSuccessDesc  *prometheus.Desc
	collectorScrapeTimeoutDesc  *prometheus.Desc
	snapshotDuration            *prometheus.Desc
}

type collectorStatus struct {
	name       string
	statusCode collectorStatusCode
}

type collectorStatusCode int

const (
	pending collectorStatusCode = iota
	success
	failed
)

// NewPrometheusCollector returns a new Prometheus where the set of MetricCollectors must
// return metrics within the given timeout.
func (c *MetricCollectors) NewPrometheusCollector(timeout time.Duration, logger *slog.Logger) *Prometheus {
	return &Prometheus{
		maxScrapeDuration: timeout,
		metricCollectors:  c,
		logger:            logger,
		scrapeDurationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "exporter", "scrape_duration_seconds"),
			"windows_exporter: Total scrape duration.",
			nil,
			nil,
		),
		collectorScrapeDurationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "exporter", "collector_duration_seconds"),
			"windows_exporter: Duration of a collection.",
			[]string{"collector"},
			nil,
		),
		collectorScrapeSuccessDesc: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "exporter", "collector_success"),
			"windows_exporter: Whether the collector was successful.",
			[]string{"collector"},
			nil,
		),
		collectorScrapeTimeoutDesc: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "exporter", "collector_timeout"),
			"windows_exporter: Whether the collector timed out.",
			[]string{"collector"},
			nil,
		),
		snapshotDuration: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "exporter", "perflib_snapshot_duration_seconds"),
			"Duration of perflib snapshot capture",
			nil,
			nil,
		),
	}
}

func (p *Prometheus) Describe(_ chan<- *prometheus.Desc) {}

// Collect sends the collected metrics from each of the MetricCollectors to
// prometheus.
func (p *Prometheus) Collect(ch chan<- prometheus.Metric) {
	t := time.Now()

	// Scrape Performance Counters for all collectors
	scrapeContext, err := p.metricCollectors.PrepareScrapeContext()

	ch <- prometheus.MustNewConstMetric(
		p.snapshotDuration,
		prometheus.GaugeValue,
		time.Since(t).Seconds(),
	)

	if err != nil {
		ch <- prometheus.NewInvalidMetric(p.collectorScrapeSuccessDesc, fmt.Errorf("failed to prepare scrape: %w", err))

		return
	}

	// WaitGroup to wait for all collectors to finish
	wg := sync.WaitGroup{}
	wg.Add(len(p.metricCollectors.Collectors))

	// Using a channel to collect the status of each collector
	// A channel is safe to use concurrently while a map is not
	collectorStatusCh := make(chan collectorStatus, len(p.metricCollectors.Collectors))

	// Execute all collectors concurrently
	// timeout handling is done in the execute function
	for name, metricsCollector := range p.metricCollectors.Collectors {
		go func(name string, metricsCollector Collector) {
			defer wg.Done()

			collectorStatusCh <- collectorStatus{
				name:       name,
				statusCode: p.execute(name, metricsCollector, scrapeContext, ch),
			}
		}(name, metricsCollector)
	}

	// Wait for all collectors to finish
	wg.Wait()

	// Close the channel since we are done writing to it
	close(collectorStatusCh)

	for status := range collectorStatusCh {
		var successValue, timeoutValue float64
		if status.statusCode == pending {
			timeoutValue = 1.0
		}

		if status.statusCode == success {
			successValue = 1.0
		}

		ch <- prometheus.MustNewConstMetric(
			p.collectorScrapeSuccessDesc,
			prometheus.GaugeValue,
			successValue,
			status.name,
		)

		ch <- prometheus.MustNewConstMetric(
			p.collectorScrapeTimeoutDesc,
			prometheus.GaugeValue,
			timeoutValue,
			status.name,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		p.scrapeDurationDesc,
		prometheus.GaugeValue,
		time.Since(t).Seconds(),
	)
}

func (p *Prometheus) execute(name string, c Collector, ctx *types.ScrapeContext, ch chan<- prometheus.Metric) collectorStatusCode {
	var (
		err      error
		duration time.Duration
		timeout  atomic.Bool
	)

	// bufCh is a buffer channel to store the metrics
	// This is needed because once timeout is reached, the prometheus registry channel is closed.
	bufCh := make(chan prometheus.Metric, 10)
	errCh := make(chan error, 1)

	// Execute the collector
	go func() {
		errCh <- c.Collect(ctx, p.logger, bufCh)

		close(bufCh)
	}()

	go func() {
		defer func() {
			// This prevents a panic from race-condition when closing the ch channel too early.
			_ = recover()
		}()

		// Pass metrics to the prometheus registry
		// If timeout is reached, the channel is closed.
		// This will cause a panic if we try to write to it.
		for m := range bufCh {
			if !timeout.Load() {
				ch <- m
			}
		}
	}()

	t := time.Now()

	// Wait for the collector to finish or timeout
	select {
	case err = <-errCh:
		duration = time.Since(t)
		ch <- prometheus.MustNewConstMetric(
			p.collectorScrapeDurationDesc,
			prometheus.GaugeValue,
			duration.Seconds(),
			name,
		)
	case <-time.After(p.maxScrapeDuration):
		timeout.Store(true)

		duration = time.Since(t)
		ch <- prometheus.MustNewConstMetric(
			p.collectorScrapeDurationDesc,
			prometheus.GaugeValue,
			duration.Seconds(),
			name,
		)

		p.logger.Warn(fmt.Sprintf("collector %s timeouted after %s", name, p.maxScrapeDuration))

		return pending
	}

	if err != nil {
		p.logger.Error(fmt.Sprintf("collector %s failed after %s", name, p.maxScrapeDuration),
			slog.Any("err", err),
		)

		return failed
	}

	p.logger.Error(fmt.Sprintf("collector %s succeeded after %s", name, p.maxScrapeDuration))

	return success
}
