// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package collector

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

type collectorStatus struct {
	name       string
	statusCode collectorStatusCode
}

type collectorStatusCode int

const (
	// pending means the collector timed out, or was skipped because an earlier call is still running.
	pending collectorStatusCode = iota
	success
	failed
	// unavailable means the collector was skipped because its Build failed or the collection is closed.
	unavailable
)

// collectAll runs all collectors once. It returns when ctx is done, even if collectors are still running.
// ctx must have a deadline.
func (c *Collection) collectAll(ctx context.Context, ch chan<- prometheus.Metric, logger *slog.Logger) {
	collectorStartTime := time.Now()

	if err := acquireScrapeSlot(ctx); err != nil {
		level := slog.LevelWarn
		if errors.Is(err, context.Canceled) {
			level = slog.LevelDebug
		}

		logger.LogAttrs(ctx, level, "scrape not started, because an earlier scrape was still running",
			slog.Any("err", err),
		)

		for name := range c.collectors {
			c.sendCollectorStatus(ch, collectorStatus{name: name, statusCode: pending})
		}

		c.sendScrapeDuration(ch, collectorStartTime)

		return
	}

	defer releaseScrapeSlot()

	// WaitGroup to wait for all collectors to finish
	wg := sync.WaitGroup{}

	// Using a channel to collect the status of each collector
	// A channel is safe to use concurrently while a map is not
	collectorStatusCh := make(chan collectorStatus, len(c.collectors))

	// Execute all collectors concurrently
	// timeout handling is done in the execute function
	for name, metricsCollector := range c.collectors {
		wg.Go(func() {
			pprof.Do(ctx, pprof.Labels("collector", name), func(ctx context.Context) {
				collectorStatusCh <- collectorStatus{
					name:       name,
					statusCode: c.collectCollector(ctx, ch, logger, name, metricsCollector),
				}
			})
		})
	}

	// Wait for all collectors to finish
	wg.Wait()

	// Close the channel since we are done writing to it
	close(collectorStatusCh)

	for status := range collectorStatusCh {
		c.sendCollectorStatus(ch, status)
	}

	c.sendScrapeDuration(ch, collectorStartTime)
}

func (c *Collection) sendCollectorStatus(ch chan<- prometheus.Metric, status collectorStatus) {
	var successValue, timeoutValue float64
	if status.statusCode == pending {
		timeoutValue = 1.0
	}

	if status.statusCode == success {
		successValue = 1.0
	}

	ch <- prometheus.MustNewConstMetric(
		c.collectorScrapeSuccessDesc,
		prometheus.GaugeValue,
		successValue,
		status.name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.collectorScrapeTimeoutDesc,
		prometheus.GaugeValue,
		timeoutValue,
		status.name,
	)
}

func (c *Collection) sendScrapeDuration(ch chan<- prometheus.Metric, collectorStartTime time.Time) {
	ch <- prometheus.MustNewConstMetric(
		c.scrapeDurationDesc,
		prometheus.GaugeValue,
		time.Since(collectorStartTime).Seconds(),
	)
}

// collectCollector runs one collector and forwards its metrics to ch until ctx is done.
//
// Only one Collect call runs per collector instance. A call that outlives ctx keeps running in the
// background; its late metrics are dropped, and later scrapes skip the collector until the call returns.
func (c *Collection) collectCollector(ctx context.Context, ch chan<- prometheus.Metric, logger *slog.Logger, name string, collector Collector) collectorStatusCode {
	deadline, _ := ctx.Deadline()

	maxScrapeDuration := time.Until(deadline)
	if ctx.Err() != nil || maxScrapeDuration <= 0 {
		return pending
	}

	state := c.state.collector(name)

	switch state.tryCollect() {
	case started:
	case notAvailable:
		logger.LogAttrs(ctx, slog.LevelDebug, fmt.Sprintf("collector %s skipped, because it isn't initialized", name))

		return unavailable
	case busy:
		level := slog.LevelDebug
		if state.logBusy() {
			level = slog.LevelWarn
		}

		logger.LogAttrs(ctx, level, fmt.Sprintf("collector %s skipped, because its call from an earlier scrape is still running", name))

		return pending
	}

	// bufCh is a buffer channel to store the metrics
	// This is needed because once timeout is reached, the prometheus registry channel is closed.
	bufCh := make(chan prometheus.Metric, 1000)
	errCh := make(chan error, 1)

	t := time.Now()

	// execute the collector. The goroutine owns the collector's slot until Collect returns.
	go func() {
		// Release the slot before closing bufCh, so the next scrape never sees this call as running.
		defer close(bufCh)
		defer state.release()

		errCh <- collect(name, collector, bufCh, maxScrapeDuration)
	}()

	numMetrics := 0

	// Pass metrics to the prometheus registry until the collector finishes or the scrape ends.
	// This is the only goroutine that writes to ch, and it stops before collectCollector returns.
	for {
		select {
		case m, ok := <-bufCh:
			if ok {
				ch <- m

				numMetrics++

				continue
			}

			duration := time.Since(t)
			c.sendCollectorDuration(ch, name, duration)

			return logCollectorResult(ctx, logger, name, <-errCh, duration, numMetrics)
		case <-ctx.Done():
			// Drain late metrics, so that Collect doesn't block on a full buffer.
			go func() {
				for range bufCh {
				}
			}()

			duration := time.Since(t)
			c.sendCollectorDuration(ch, name, duration)

			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				logger.LogAttrs(ctx, slog.LevelWarn, fmt.Sprintf("collector %s timeouted after %s, resulting in %d metrics", name, duration, numMetrics))
			} else {
				logger.LogAttrs(ctx, slog.LevelDebug, fmt.Sprintf("collector %s canceled after %s, resulting in %d metrics", name, duration, numMetrics),
					slog.Any("err", ctx.Err()),
				)
			}

			return pending
		}
	}
}

// collect calls Collector.Collect and turns a panic into an error.
func collect(name string, collector Collector, ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in collector %s: %v. stack: %s", name, r,
				string(debug.Stack()),
			)
		}
	}()

	return collector.Collect(ch, maxScrapeDuration)
}

func (c *Collection) sendCollectorDuration(ch chan<- prometheus.Metric, name string, duration time.Duration) {
	ch <- prometheus.MustNewConstMetric(
		c.collectorScrapeDurationDesc,
		prometheus.GaugeValue,
		duration.Seconds(),
		name,
	)
}

func logCollectorResult(ctx context.Context, logger *slog.Logger, name string, err error, duration time.Duration, numMetrics int) collectorStatusCode {
	slogAttrs := make([]slog.Attr, 0)

	result := "succeeded"

	if err != nil {
		if !expectedCollectionError(err) {
			if errors.Is(err, pdh.ErrPerformanceCounterNotInitialized) {
				err = fmt.Errorf("%w. Check application logs from initialization pharse for more information", err)
			}

			logger.LogAttrs(ctx, slog.LevelWarn,
				fmt.Sprintf("collector %s failed after %s, resulting in %d metrics", name, duration, numMetrics),
				slog.Any("err", err),
			)

			return failed
		}

		slogAttrs = append(slogAttrs, slog.Any("err", err))

		result = "succeeded with warnings"
	}

	logger.LogAttrs(ctx, slog.LevelDebug, fmt.Sprintf(
		"collector %s %s after %s, resulting in %d metrics", name, result, duration, numMetrics,
	),
		slogAttrs...,
	)

	return success
}

// expectedCollectionError checks every cause so an expected error cannot hide an independent failure.
func expectedCollectionError(err error) bool {
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		causes := joined.Unwrap()
		for _, cause := range causes {
			if !expectedCollectionError(cause) {
				return false
			}
		}

		return len(causes) != 0
	}

	if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		return expectedCollectionError(wrapped.Unwrap())
	}

	noData := errors.Is(err, pdh.ErrNoData) || errors.Is(err, types.ErrNoData)

	return noData || errors.Is(err, windows.EPT_S_NOT_REGISTERED)
}
