package httphandler

import (
	"fmt"
	stdlog "log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/uuid"
	"github.com/prometheus-community/windows_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Interface guard.
var _ http.Handler = (*MetricsHTTPHandler)(nil)

const defaultScrapeTimeout = 10.0

type MetricsHTTPHandler struct {
	metricCollectors *collector.MetricCollectors
	// exporterMetricsRegistry is a separate registry for the metrics about
	// the exporter itself.
	exporterMetricsRegistry *prometheus.Registry

	logger        log.Logger
	options       Options
	concurrencyCh chan struct{}
}

type Options struct {
	DisableExporterMetrics bool
	TimeoutMargin          float64
	MaxRequests            int
}

func New(logger log.Logger, metricCollectors *collector.MetricCollectors, options *Options) *MetricsHTTPHandler {
	if options == nil {
		options = &Options{
			DisableExporterMetrics: false,
			TimeoutMargin:          0.5,
			MaxRequests:            5,
		}
	}

	handler := &MetricsHTTPHandler{
		metricCollectors: metricCollectors,
		logger:           logger,
		options:          *options,
		concurrencyCh:    make(chan struct{}, options.MaxRequests),
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
	logger := log.With(c.logger, "remote", r.RemoteAddr, "correlation_id", uuid.New().String())

	scrapeTimeout := c.getScrapeTimeout(logger, r)
	handler, err := c.handlerFactory(scrapeTimeout, logger, r.URL.Query()["collect[]"])
	if err != nil {
		_ = level.Warn(logger).Log("msg", "Couldn't create filtered metrics handler", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(fmt.Sprintf("Couldn't create filtered metrics handler: %s", err)))
		return
	}

	handler.ServeHTTP(w, r)
}

func (c *MetricsHTTPHandler) getScrapeTimeout(logger log.Logger, r *http.Request) time.Duration {
	var timeoutSeconds float64

	if v := r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); v != "" {
		var err error
		timeoutSeconds, err = strconv.ParseFloat(v, 64)
		if err != nil {
			_ = level.Warn(logger).Log("msg", fmt.Sprintf("Couldn't parse X-Prometheus-Scrape-Timeout-Seconds: %q. Defaulting timeout to %f", v, defaultScrapeTimeout))
		}
	}

	if timeoutSeconds == 0 {
		timeoutSeconds = defaultScrapeTimeout
	}

	timeoutSeconds -= c.options.TimeoutMargin

	return time.Duration(timeoutSeconds) * time.Second
}

func (c *MetricsHTTPHandler) handlerFactory(scrapeTimeout time.Duration, logger log.Logger, requestedCollectors []string) (http.Handler, error) {
	reg := prometheus.NewRegistry()

	var metricCollectors *collector.MetricCollectors
	if len(requestedCollectors) == 0 {
		metricCollectors = c.metricCollectors
	} else {
		filteredCollectors := make(collector.Map)
		for _, name := range requestedCollectors {
			metricCollector, ok := c.metricCollectors.Collectors[name]
			if !ok {
				return nil, fmt.Errorf("couldn't find collector %s", name)
			}

			filteredCollectors[name] = metricCollector
		}

		metricCollectors = &collector.MetricCollectors{
			Collectors:       filteredCollectors,
			WMIClient:        c.metricCollectors.WMIClient,
			PerfCounterQuery: c.metricCollectors.PerfCounterQuery,
		}
	}

	reg.MustRegister(version.NewCollector("windows_exporter"))
	if err := reg.Register(metricCollectors.NewPrometheusCollector(scrapeTimeout, c.logger)); err != nil {
		return nil, fmt.Errorf("couldn't register Prometheus collector: %w", err)
	}

	var handler http.Handler
	if c.exporterMetricsRegistry != nil {
		handler = promhttp.HandlerFor(
			prometheus.Gatherers{c.exporterMetricsRegistry, reg},
			promhttp.HandlerOpts{
				ErrorLog:            stdlog.New(log.NewStdlibAdapter(level.Error(logger)), "", stdlog.Lshortfile),
				ErrorHandling:       promhttp.ContinueOnError,
				MaxRequestsInFlight: c.options.MaxRequests,
				Registry:            c.exporterMetricsRegistry,
			},
		)

		// Note that we have to use h.exporterMetricsRegistry here to
		// use the same promhttp metrics for all expositions.
		handler = promhttp.InstrumentMetricHandler(
			c.exporterMetricsRegistry, handler,
		)
	} else {
		handler = promhttp.HandlerFor(
			reg,
			promhttp.HandlerOpts{
				ErrorLog:            stdlog.New(log.NewStdlibAdapter(level.Error(logger)), "", 0),
				ErrorHandling:       promhttp.ContinueOnError,
				MaxRequestsInFlight: c.options.MaxRequests,
			},
		)
	}

	return c.withConcurrencyLimit(handler.ServeHTTP), nil
}

func (c *MetricsHTTPHandler) withConcurrencyLimit(next http.HandlerFunc) http.HandlerFunc {
	if c.options.MaxRequests <= 0 {
		return next
	}

	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case c.concurrencyCh <- struct{}{}:
			defer func() { <-c.concurrencyCh }()
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("Too many concurrent requests"))
			return
		}

		next(w, r)
	}
}
