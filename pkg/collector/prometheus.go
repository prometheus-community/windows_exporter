//go:build windows

package collector

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/types"
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

	ch <- prometheus.MustNewConstMetric(
		p.snapshotDuration,
		prometheus.GaugeValue,
		time.Since(t).Seconds(),
	)

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
				statusCode: p.execute(name, metricsCollector, ch),
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

func (p *Prometheus) execute(name string, c Collector, ch chan<- prometheus.Metric) collectorStatusCode {
	var (
		err        error
		numMetrics int
		duration   time.Duration
		timeout    atomic.Bool
	)

	// bufCh is a buffer channel to store the metrics
	// This is needed because once timeout is reached, the prometheus registry channel is closed.
	bufCh := make(chan prometheus.Metric, 1000)
	errCh := make(chan error, 1)

	ctx, cancel := context.WithTimeout(context.Background(), p.maxScrapeDuration)
	defer cancel()

	// Execute the collector
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errCh <- fmt.Errorf("panic in collector %s: %v. stack: %s", name, r,
					string(debug.Stack()),
				)
			}

			close(bufCh)
		}()

		errCh <- c.Collect(bufCh)
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer func() {
			// This prevents a panic from race-condition when closing the ch channel too early.
			_ = recover()

			wg.Done()
		}()

		// Pass metrics to the prometheus registry
		// If timeout is reached, the channel is closed.
		// This will cause a panic if we try to write to it.
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-bufCh:
				if !ok {
					return
				}

				if !timeout.Load() {
					ch <- m

					numMetrics++
				}
			}
		}
	}()

	t := time.Now()

	// Wait for the collector to finish or timeout
	select {
	case err = <-errCh:
		wg.Wait() // Wait for the buffer channel to be closed and empty

		duration = time.Since(t)
		ch <- prometheus.MustNewConstMetric(
			p.collectorScrapeDurationDesc,
			prometheus.GaugeValue,
			duration.Seconds(),
			name,
		)
	case <-ctx.Done():
		timeout.Store(true)

		duration = time.Since(t)
		ch <- prometheus.MustNewConstMetric(
			p.collectorScrapeDurationDesc,
			prometheus.GaugeValue,
			duration.Seconds(),
			name,
		)

		p.logger.Warn(fmt.Sprintf("collector %s timeouted after %s, resulting in %d metrics", name, p.maxScrapeDuration, numMetrics))

		go func() {
			// Drain channel in case of premature return to not leak a goroutine.
			//nolint:revive
			for range bufCh {
			}
		}()

		return pending
	}

	if err != nil {
		p.logger.Error(fmt.Sprintf("collector %s failed after %s, resulting in %d metrics", name, duration, numMetrics),
			slog.Any("err", err),
		)

		return failed
	}

	p.logger.Debug(fmt.Sprintf("collector %s succeeded after %s, resulting in %d metrics", name, duration, numMetrics))

	return success
}
