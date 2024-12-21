// Copyright 2024 The Prometheus Authors
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

package collector

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

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

func (c *Collection) collectAll(ch chan<- prometheus.Metric, logger *slog.Logger, maxScrapeDuration time.Duration) {
	collectorStartTime := time.Now()

	// WaitGroup to wait for all collectors to finish
	wg := sync.WaitGroup{}
	wg.Add(len(c.collectors))

	// Using a channel to collect the status of each collector
	// A channel is safe to use concurrently while a map is not
	collectorStatusCh := make(chan collectorStatus, len(c.collectors))

	// Execute all collectors concurrently
	// timeout handling is done in the execute function
	for name, metricsCollector := range c.collectors {
		go func(name string, metricsCollector Collector) {
			defer wg.Done()

			collectorStatusCh <- collectorStatus{
				name:       name,
				statusCode: c.collectCollector(ch, logger, name, metricsCollector, maxScrapeDuration),
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

	ch <- prometheus.MustNewConstMetric(
		c.scrapeDurationDesc,
		prometheus.GaugeValue,
		time.Since(collectorStartTime).Seconds(),
	)
}

func (c *Collection) collectCollector(ch chan<- prometheus.Metric, logger *slog.Logger, name string, collector Collector, maxScrapeDuration time.Duration) collectorStatusCode {
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

	ctx, cancel := context.WithTimeout(context.Background(), maxScrapeDuration)
	defer cancel()

	// execute the collector
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errCh <- fmt.Errorf("panic in collector %s: %v. stack: %s", name, r,
					string(debug.Stack()),
				)
			}

			close(bufCh)
		}()

		errCh <- collector.Collect(bufCh)
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
			c.collectorScrapeDurationDesc,
			prometheus.GaugeValue,
			duration.Seconds(),
			name,
		)
	case <-ctx.Done():
		timeout.Store(true)

		duration = time.Since(t)
		ch <- prometheus.MustNewConstMetric(
			c.collectorScrapeDurationDesc,
			prometheus.GaugeValue,
			duration.Seconds(),
			name,
		)

		logger.LogAttrs(ctx, slog.LevelWarn, fmt.Sprintf("collector %s timeouted after %s, resulting in %d metrics", name, maxScrapeDuration, numMetrics))

		go func() {
			// Drain channel in case of premature return to not leak a goroutine.
			//nolint:revive
			for range bufCh {
			}
		}()

		return pending
	}

	if err != nil && !errors.Is(err, pdh.ErrNoData) && !errors.Is(err, types.ErrNoData) {
		if errors.Is(err, pdh.ErrPerformanceCounterNotInitialized) || errors.Is(err, mi.MI_RESULT_INVALID_NAMESPACE) {
			err = fmt.Errorf("%w. Check application logs from initialization pharse for more information", err)
		}

		logger.LogAttrs(ctx, slog.LevelWarn,
			fmt.Sprintf("collector %s failed after %s, resulting in %d metrics", name, duration, numMetrics),
			slog.Any("err", err),
		)

		return failed
	}

	logger.LogAttrs(ctx, slog.LevelDebug, fmt.Sprintf("collector %s succeeded after %s, resulting in %d metrics", name, duration, numMetrics))

	return success
}
