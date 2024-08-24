//go:build windows

package collector

import (
	"fmt"
	stdlog "log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (c *Collectors) BuildServeHTTP(logger log.Logger, disableExporterMetrics bool, timeoutMargin float64) http.HandlerFunc {
	collectorFactory := func(timeout time.Duration, requestedCollectors []string) (error, *Prometheus) {
		filteredCollectors := make(map[string]Collector)
		// scrape all enabled collectors if no collector is requested
		if len(requestedCollectors) == 0 {
			filteredCollectors = c.collectors
		}
		for _, name := range requestedCollectors {
			col, exists := c.collectors[name]
			if !exists {
				return fmt.Errorf("unavailable collector: %s", name), nil
			}
			filteredCollectors[name] = col
		}

		filtered := Collectors{
			collectors:       filteredCollectors,
			perfCounterQuery: c.perfCounterQuery,
		}

		return nil, NewPrometheus(timeout, &filtered, logger)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		logger := log.With(logger, "remote", r.RemoteAddr, "correlation_id", uuid.New().String())

		const defaultTimeout = 10.0

		var timeoutSeconds float64
		if v := r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); v != "" {
			var err error
			timeoutSeconds, err = strconv.ParseFloat(v, 64)
			if err != nil {
				_ = level.Warn(logger).Log("msg", fmt.Sprintf("Couldn't parse X-Prometheus-Scrape-Timeout-Seconds: %q. Defaulting timeout to %f", v, defaultTimeout))
			}
		}
		if timeoutSeconds == 0 {
			timeoutSeconds = defaultTimeout
		}
		timeoutSeconds -= timeoutMargin

		reg := prometheus.NewRegistry()
		err, wc := collectorFactory(time.Duration(timeoutSeconds*float64(time.Second)), r.URL.Query()["collect[]"])
		if err != nil {
			_ = level.Warn(logger).Log("msg", "Couldn't create filtered metrics handler", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Couldn't create filtered metrics handler: %s", err))) //nolint:errcheck
			return
		}

		reg.MustRegister(wc)
		if !disableExporterMetrics {
			reg.MustRegister(
				collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
				collectors.NewGoCollector(),
				version.NewCollector("windows_exporter"),
			)
		}

		h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{
			ErrorLog: stdlog.New(log.NewStdlibAdapter(level.Error(logger)), "", stdlog.Lshortfile),
		})
		h.ServeHTTP(w, r)
	}
}
