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

package httphandler

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Interface guard.
var _ http.Handler = (*MetricsHTTPHandler)(nil)

const (
	// defaultScrapeTimeout applies if the request has no valid X-Prometheus-Scrape-Timeout-Seconds header.
	defaultScrapeTimeout = 10 * time.Second
	// minScrapeTimeout is the smallest time given to the collectors.
	minScrapeTimeout = 10 * time.Millisecond
	// maxScrapeTimeout is the largest time given to the collectors.
	// The exporter's HTTP server ends responses after 5 minutes anyway.
	maxScrapeTimeout = 5 * time.Minute
)

type MetricsHTTPHandler struct {
	metricCollectors *collector.Collection
	// exporterMetricsRegistry is a separate registry for the metrics about
	// the exporter itself.
	exporterMetricsRegistry *prometheus.Registry

	logger        *slog.Logger
	timeoutMargin time.Duration
}

type Options struct {
	DisableExporterMetrics bool
	TimeoutMargin          float64
}

func New(logger *slog.Logger, metricCollectors *collector.Collection, options *Options) *MetricsHTTPHandler {
	if options == nil {
		options = &Options{
			DisableExporterMetrics: false,
			TimeoutMargin:          0.5,
		}
	}

	handler := &MetricsHTTPHandler{
		metricCollectors: metricCollectors,
		logger:           logger,
		timeoutMargin:    toDuration(options.TimeoutMargin),
	}

	if !options.DisableExporterMetrics {
		handler.exporterMetricsRegistry = prometheus.NewRegistry()
		handler.exporterMetricsRegistry.MustRegister(
			collectors.NewBuildInfoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewGoCollector(),
		)
	}

	return handler
}

func (c *MetricsHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := c.logger.With(
		slog.String("remote", r.RemoteAddr),
	)

	// The scrape's deadline starts now, so time spent waiting for an earlier scrape counts against it.
	scrapeTimeout := c.getScrapeTimeout(logger, r)

	handler, err := c.handlerFactory(r.Context(), logger, scrapeTimeout, r.URL.Query()["collect[]"])
	if err != nil {
		logger.WarnContext(r.Context(), "Couldn't create filtered metrics handler",
			slog.Any("err", err),
		)

		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "Couldn't create filtered metrics handler: %s", err)

		return
	}

	handler.ServeHTTP(w, r)
}

// getScrapeTimeout returns the time the collectors get for this request.
// It is the client's X-Prometheus-Scrape-Timeout-Seconds minus the timeout margin,
// but the margin takes at most half of the client's timeout.
// The result is between minScrapeTimeout and maxScrapeTimeout.
func (c *MetricsHTTPHandler) getScrapeTimeout(logger *slog.Logger, r *http.Request) time.Duration {
	timeout := defaultScrapeTimeout

	if v := r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); v != "" {
		seconds, err := strconv.ParseFloat(v, 64)

		switch {
		case math.IsInf(seconds, 1):
			// Includes values too large for a float64, which ParseFloat reports as an error.
			timeout = maxScrapeTimeout
		case err != nil, math.IsNaN(seconds), seconds <= 0:
			logger.WarnContext(r.Context(), fmt.Sprintf("Invalid X-Prometheus-Scrape-Timeout-Seconds: %q. Defaulting timeout to %s", v, defaultScrapeTimeout))
		default:
			timeout = toDuration(seconds)
		}
	}

	timeout = max(timeout-c.timeoutMargin, timeout/2)

	return min(max(timeout, minScrapeTimeout), maxScrapeTimeout)
}

// toDuration converts seconds to a duration between 0 and maxScrapeTimeout.
// NaN and negative values become 0.
func toDuration(seconds float64) time.Duration {
	if math.IsNaN(seconds) || seconds <= 0 {
		return 0
	}

	if seconds >= maxScrapeTimeout.Seconds() {
		return maxScrapeTimeout
	}

	return time.Duration(seconds * float64(time.Second))
}

func (c *MetricsHTTPHandler) handlerFactory(ctx context.Context, logger *slog.Logger, scrapeTimeout time.Duration, requestedCollectors []string) (http.Handler, error) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(version.NewCollector("windows_exporter"))

	collectionHandler, err := c.metricCollectors.NewHandlerWithContext(ctx, scrapeTimeout, logger, requestedCollectors)
	if err != nil {
		return nil, fmt.Errorf("couldn't create collector handler: %w", err)
	}

	if err := reg.Register(collectionHandler); err != nil {
		return nil, fmt.Errorf("couldn't register Prometheus collector: %w", err)
	}

	// Scrapes are serialized by the collection, so the handler doesn't limit requests in flight.
	var regHandler http.Handler
	if c.exporterMetricsRegistry != nil {
		regHandler = promhttp.HandlerFor(
			prometheus.Gatherers{c.exporterMetricsRegistry, reg},
			promhttp.HandlerOpts{
				ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
				ErrorHandling:     promhttp.ContinueOnError,
				Registry:          c.exporterMetricsRegistry,
				EnableOpenMetrics: true,
				ProcessStartTime:  c.metricCollectors.GetStartTime(),
			},
		)

		// Note that we have to use h.exporterMetricsRegistry here to
		// use the same promhttp metrics for all expositions.
		regHandler = promhttp.InstrumentMetricHandler(
			c.exporterMetricsRegistry, regHandler,
		)
	} else {
		regHandler = promhttp.HandlerFor(
			reg,
			promhttp.HandlerOpts{
				ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
				ErrorHandling:     promhttp.ContinueOnError,
				EnableOpenMetrics: true,
				ProcessStartTime:  c.metricCollectors.GetStartTime(),
			},
		)
	}

	return regHandler, nil
}
