//go:build windows

package collector

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus implements prometheus.Collector for a set of Windows Collectors.
type Prometheus struct {
	maxScrapeDuration time.Duration
	collectors        *Collectors
	logger            log.Logger

	// Base metrics returned by Prometheus
	scrapeDurationDesc *prometheus.Desc
	scrapeSuccessDesc  *prometheus.Desc
	scrapeTimeoutDesc  *prometheus.Desc
	snapshotDuration   *prometheus.Desc
}

// NewPrometheus returns a new Prometheus where the set of Collectors must
// return metrics within the given timeout.
func NewPrometheus(timeout time.Duration, cs *Collectors, logger log.Logger) *Prometheus {
	return &Prometheus{
		maxScrapeDuration: timeout,
		collectors:        cs,
		logger:            logger,
		scrapeDurationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "exporter", "collector_duration_seconds"),
			"windows_exporter: Duration of a collection.",
			[]string{"collector"},
			nil,
		),
		scrapeSuccessDesc: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "exporter", "collector_success"),
			"windows_exporter: Whether the collector was successful.",
			[]string{"collector"},
			nil,
		),
		scrapeTimeoutDesc: prometheus.NewDesc(
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

// Describe sends all the descriptors of the Collectors included to
// the provided channel.
func (coll *Prometheus) Describe(ch chan<- *prometheus.Desc) {
	ch <- coll.scrapeDurationDesc
	ch <- coll.scrapeSuccessDesc
}

type collectorOutcome int

const (
	pending collectorOutcome = iota
	success
	failed
)

// Collect sends the collected metrics from each of the Collectors to
// prometheus.
func (coll *Prometheus) Collect(ch chan<- prometheus.Metric) {
	t := time.Now()

	scrapeContext, err := coll.collectors.PrepareScrapeContext()
	ch <- prometheus.MustNewConstMetric(
		coll.snapshotDuration,
		prometheus.GaugeValue,
		time.Since(t).Seconds(),
	)
	if err != nil {
		ch <- prometheus.NewInvalidMetric(coll.scrapeSuccessDesc, fmt.Errorf("failed to prepare scrape: %v", err))
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(len(coll.collectors.collectors))
	collectorOutcomes := make(map[string]collectorOutcome)
	for name := range coll.collectors.collectors {
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

	for name, c := range coll.collectors.collectors {
		go func(name string, c Collector) {
			defer wg.Done()
			outcome := coll.execute(name, c, scrapeContext, metricsBuffer)
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

	// Wait until either all Collectors finish, or timeout expires
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
			coll.scrapeSuccessDesc,
			prometheus.GaugeValue,
			successValue,
			name,
		)
		ch <- prometheus.MustNewConstMetric(
			coll.scrapeTimeoutDesc,
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

func (coll *Prometheus) execute(name string, c Collector, ctx *types.ScrapeContext, ch chan<- prometheus.Metric) collectorOutcome {
	t := time.Now()
	err := c.Collect(ctx, ch)
	duration := time.Since(t).Seconds()
	ch <- prometheus.MustNewConstMetric(
		coll.scrapeDurationDesc,
		prometheus.GaugeValue,
		duration,
		name,
	)

	if err != nil {
		_ = level.Error(coll.logger).Log("msg", fmt.Sprintf("collector %s failed after %fs", name, duration), "err", err)
		return failed
	}
	_ = level.Debug(coll.logger).Log("msg", fmt.Sprintf("collector %s succeeded after %fs.", name, duration))
	return success
}
