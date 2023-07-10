//go:build windows
// +build windows

package collector

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// Base metrics returned by Prometheus

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "exporter", "collector_duration_seconds"),
		"windows_exporter: Duration of a collection.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "exporter", "collector_success"),
		"windows_exporter: Whether the collector was successful.",
		[]string{"collector"},
		nil,
	)
	scrapeTimeoutDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "exporter", "collector_timeout"),
		"windows_exporter: Whether the collector timed out.",
		[]string{"collector"},
		nil,
	)
	snapshotDuration = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "exporter", "perflib_snapshot_duration_seconds"),
		"Duration of perflib snapshot capture",
		nil,
		nil,
	)
)

// Prometheus implements prometheus.Collector for a set of Windows collectors.
type Prometheus struct {
	maxScrapeDuration time.Duration
	collectors        map[string]Collector
	logger            log.Logger
}

// NewPrometheus returns a new Prometheus where the set of collectors must
// return metrics within the given timeout.
func NewPrometheus(timeout time.Duration, cs map[string]Collector, logger log.Logger) *Prometheus {
	return &Prometheus{
		maxScrapeDuration: timeout,
		collectors:        cs,
		logger:            logger,
	}
}

// Describe sends all the descriptors of the collectors included to
// the provided channel.
func (coll *Prometheus) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

type collectorOutcome int

const (
	pending collectorOutcome = iota
	success
	failed
)

// Collect sends the collected metrics from each of the collectors to
// prometheus.
func (coll *Prometheus) Collect(ch chan<- prometheus.Metric) {
	t := time.Now()
	cs := make([]string, 0, len(coll.collectors))
	for name := range coll.collectors {
		cs = append(cs, name)
	}
	scrapeContext, err := PrepareScrapeContext(cs)
	ch <- prometheus.MustNewConstMetric(
		snapshotDuration,
		prometheus.GaugeValue,
		time.Since(t).Seconds(),
	)
	if err != nil {
		ch <- prometheus.NewInvalidMetric(scrapeSuccessDesc, fmt.Errorf("failed to prepare scrape: %v", err))
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(coll.collectors))
	collectorOutcomes := make(map[string]collectorOutcome)
	for name := range coll.collectors {
		collectorOutcomes[name] = pending
	}

	metricsBuffer := make(chan prometheus.Metric)
	l := sync.Mutex{}
	finished := false
	go func() {
		for m := range metricsBuffer {
			l.Lock()
			if !finished {
				ch <- m
			}
			l.Unlock()
		}
	}()

	for name, c := range coll.collectors {
		go func(name string, c Collector) {
			defer wg.Done()
			outcome := execute(name, c, scrapeContext, metricsBuffer, coll.logger)
			l.Lock()
			if !finished {
				collectorOutcomes[name] = outcome
			}
			l.Unlock()
		}(name, c)
	}

	allDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(allDone)
		close(metricsBuffer)
	}()

	// Wait until either all collectors finish, or timeout expires
	select {
	case <-allDone:
	case <-time.After(coll.maxScrapeDuration):
	}

	l.Lock()
	finished = true

	remainingCollectorNames := make([]string, 0)
	for name, outcome := range collectorOutcomes {
		var successValue, timeoutValue float64
		if outcome == pending {
			timeoutValue = 1.0
			remainingCollectorNames = append(remainingCollectorNames, name)
		}
		if outcome == success {
			successValue = 1.0
		}

		ch <- prometheus.MustNewConstMetric(
			scrapeSuccessDesc,
			prometheus.GaugeValue,
			successValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			scrapeTimeoutDesc,
			prometheus.GaugeValue,
			timeoutValue,
			name,
		)
	}

	if len(remainingCollectorNames) > 0 {
		_ = level.Warn(coll.logger).Log("msg", fmt.Sprintf("Collection timed out, still waiting for %v", remainingCollectorNames))
	}

	l.Unlock()
}

func execute(name string, c Collector, ctx *ScrapeContext, ch chan<- prometheus.Metric, logger log.Logger) collectorOutcome {
	t := time.Now()
	err := c.Collect(ctx, ch)
	duration := time.Since(t).Seconds()
	ch <- prometheus.MustNewConstMetric(
		scrapeDurationDesc,
		prometheus.GaugeValue,
		duration,
		name,
	)

	if err != nil {
		_ = level.Error(logger).Log("msg", fmt.Sprintf("collector %s failed after %fs", name, duration), "err", err)
		return failed
	}
	_ = level.Debug(logger).Log("msg", fmt.Sprintf("collector %s succeeded after %fs.", name, duration))
	return success
}
