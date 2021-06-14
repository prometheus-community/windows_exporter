//go:build windows
// +build windows

package main

import (
	//Its important that we do these first so that we can register with the windows service control ASAP to avoid timeouts
	"github.com/prometheus-community/windows_exporter/initiate"
	winlog "github.com/prometheus-community/windows_exporter/log"

	"encoding/json"
	"fmt"
	stdlog "log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/user"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus-community/windows_exporter/collector"
	"github.com/prometheus-community/windows_exporter/config"
	"github.com/prometheus-community/windows_exporter/log/flag"
	"github.com/yusufpapurcu/wmi"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

// Same struct prometheus uses for their /version endpoint.
// Separate copy to avoid pulling all of prometheus as a dependency
type prometheusVersion struct {
	Version   string `json:"version"`
	Revision  string `json:"revision"`
	Branch    string `json:"branch"`
	BuildUser string `json:"buildUser"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
}

const (
	defaultCollectors            = "cpu,cs,logical_disk,physical_disk,net,os,service,system,textfile"
	defaultCollectorsPlaceholder = "[defaults]"
)

func expandEnabledCollectors(enabled string) []string {
	expanded := strings.Replace(enabled, defaultCollectorsPlaceholder, defaultCollectors, -1)
	separated := strings.Split(expanded, ",")
	unique := map[string]bool{}
	for _, s := range separated {
		if s != "" {
			unique[s] = true
		}
	}
	result := make([]string, 0, len(unique))
	for s := range unique {
		result = append(result, s)
	}
	return result
}

func loadCollectors(list string, logger log.Logger) (map[string]collector.Collector, error) {
	collectors := map[string]collector.Collector{}
	enabled := expandEnabledCollectors(list)

	for _, name := range enabled {
		c, err := collector.Build(name, logger)
		if err != nil {
			return nil, err
		}
		collectors[name] = c
	}

	return collectors, nil
}

func initWbem(logger log.Logger) {
	// This initialization prevents a memory leak on WMF 5+. See
	// https://github.com/prometheus-community/windows_exporter/issues/77 and
	// linked issues for details.
	_ = level.Debug(logger).Log("msg", "Initializing SWbemServices")
	s, err := wmi.InitializeSWbemServices(wmi.DefaultClient)
	if err != nil {
		_ = level.Error(logger).Log("err", err)
		os.Exit(1)
	}
	wmi.DefaultClient.AllowMissingFields = true
	wmi.DefaultClient.SWbemServicesClient = s
}

func main() {
	app := kingpin.New("windows_exporter", "A metrics collector for Windows.")
	var (
		configFile = app.Flag(
			"config.file",
			"YAML configuration file to use. Values set in this file will be overridden by CLI flags.",
		).String()
		insecure_skip_verify = app.Flag(
			"config.file.insecure-skip-verify",
			"Skip TLS verification in loading YAML configuration.",
		).Default("false").Bool()
		webConfig   = webflag.AddFlags(app, ":9182")
		metricsPath = app.Flag(
			"telemetry.path",
			"URL path for surfacing collected metrics.",
		).Default("/metrics").String()
		disableExporterMetrics = app.Flag(
			"web.disable-exporter-metrics",
			"Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).",
		).Bool()
		maxRequests = app.Flag(
			"telemetry.max-requests",
			"Maximum number of concurrent requests. 0 to disable.",
		).Default("5").Int()
		enabledCollectors = app.Flag(
			"collectors.enabled",
			"Comma-separated list of collectors to use. Use '[defaults]' as a placeholder for all the collectors enabled by default.").
			Default(defaultCollectors).String()
		printCollectors = app.Flag(
			"collectors.print",
			"If true, print available collectors and exit.",
		).Bool()
		timeoutMargin = app.Flag(
			"scrape.timeout-margin",
			"Seconds to subtract from the timeout allowed by the client. Tune to allow for overhead or high loads.",
		).Default("0.5").Float64()
	)

	winlogConfig := &winlog.Config{}
	flag.AddFlags(app, winlogConfig)

	app.Version(version.Print("windows_exporter"))
	app.HelpFlag.Short('h')

	// Initialize collectors before loading and parsing CLI arguments
	collector.RegisterCollectorsFlags(app)

	// Load values from configuration file(s). Executable flags must first be parsed, in order
	// to load the specified file(s).
	kingpin.MustParse(app.Parse(os.Args[1:]))
	logger, err := winlog.New(winlogConfig)
	if err != nil {
		_ = level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	_ = level.Debug(logger).Log("msg", "Logging has Started")
	if *configFile != "" {
		resolver, err := config.NewResolver(*configFile, logger, *insecure_skip_verify)
		if err != nil {
			_ = level.Error(logger).Log("msg", "could not load config file", "err", err)
			os.Exit(1)
		}
		err = resolver.Bind(app, os.Args[1:])
		if err != nil {
			_ = level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		// NOTE: This is temporary fix for issue #1092, calling kingpin.Parse
		// twice makes slices flags duplicate its value, this clean up
		// the first parse before the second call.
		*webConfig.WebListenAddresses = (*webConfig.WebListenAddresses)[1:]

		// Parse flags once more to include those discovered in configuration file(s).
		kingpin.MustParse(app.Parse(os.Args[1:]))

		logger, err = winlog.New(winlogConfig)
		if err != nil {
			_ = level.Error(logger).Log("err", err)
			os.Exit(1)
		}
	}

	if *printCollectors {
		collectors := collector.Available()
		collectorNames := make(sort.StringSlice, 0, len(collectors))
		for _, n := range collectors {
			collectorNames = append(collectorNames, n)
		}
		collectorNames.Sort()
		fmt.Printf("Available collectors:\n")
		for _, n := range collectorNames {
			fmt.Printf(" - %s\n", n)
		}
		return
	}

	initWbem(logger)

	// Initialize collectors before loading
	collector.RegisterCollectors(logger)

	collectors, err := loadCollectors(*enabledCollectors, logger)
	if err != nil {
		_ = level.Error(logger).Log("msg", "Couldn't load collectors", "err", err)
		os.Exit(1)
	}

	u, err := user.Current()
	if err != nil {
		_ = level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	_ = level.Info(logger).Log("msg", fmt.Sprintf("Running as %v", u.Username))
	if strings.Contains(u.Username, "ContainerAdministrator") || strings.Contains(u.Username, "ContainerUser") {
		_ = level.Warn(logger).Log("msg", "Running as a preconfigured Windows Container user. This may mean you do not have Windows HostProcess containers configured correctly and some functionality will not work as expected.")
	}

	_ = level.Info(logger).Log("msg", fmt.Sprintf("Enabled collectors: %v", strings.Join(keys(collectors), ", ")))

	h := &metricsHandler{
		timeoutMargin:          *timeoutMargin,
		includeExporterMetrics: *disableExporterMetrics,
		collectorFactory: func(timeout time.Duration, requestedCollectors []string) (error, *collector.Prometheus) {
			filteredCollectors := make(map[string]collector.Collector)
			// scrape all enabled collectors if no collector is requested
			if len(requestedCollectors) == 0 {
				filteredCollectors = collectors
			}
			for _, name := range requestedCollectors {
				col, exists := collectors[name]
				if !exists {
					return fmt.Errorf("unavailable collector: %s", name), nil
				}
				filteredCollectors[name] = col
			}
			return nil, collector.NewPrometheus(timeout, filteredCollectors, logger)
		},
		logger: logger,
	}

	http.HandleFunc(*metricsPath, withConcurrencyLimit(*maxRequests, h.ServeHTTP))
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintln(w, `{"status":"ok"}`)
		if err != nil {
			_ = level.Debug(logger).Log("Failed to write to stream", "err", err)
		}
	})
	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		// we can't use "version" directly as it is a package, and not an object that
		// can be serialized.
		err := json.NewEncoder(w).Encode(prometheusVersion{
			Version:   version.Version,
			Revision:  version.Revision,
			Branch:    version.Branch,
			BuildUser: version.BuildUser,
			BuildDate: version.BuildDate,
			GoVersion: version.GoVersion,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("error encoding JSON: %s", err), http.StatusInternalServerError)
		}
	})
	if *metricsPath != "/" && *metricsPath != "" {
		landingConfig := web.LandingConfig{
			Name:        "Windows Exporter",
			Description: "Prometheus Exporter for Windows servers",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
				{
					Address: "/health",
					Text:    "Health Check",
				},
				{
					Address: "/version",
					Text:    "Version Info",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			_ = level.Error(logger).Log("msg", "failed to generate landing page", "err", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	_ = level.Info(logger).Log("msg", "Starting windows_exporter", "version", version.Info())
	_ = level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())
	_ = level.Debug(logger).Log("msg", "Go MAXPROCS", "procs", runtime.GOMAXPROCS(0))

	go func() {
		server := &http.Server{}
		if err := web.ListenAndServe(server, webConfig, logger); err != nil {
			_ = level.Error(logger).Log("msg", "cannot start windows_exporter", "err", err)
			os.Exit(1)
		}
	}()

	for {
		if <-initiate.StopCh {
			_ = level.Info(logger).Log("msg", "Shutting down windows_exporter")
			break
		}
	}
}

func keys(m map[string]collector.Collector) []string {
	ret := make([]string, 0, len(m))
	for key := range m {
		ret = append(ret, key)
	}
	return ret
}

func withConcurrencyLimit(n int, next http.HandlerFunc) http.HandlerFunc {
	if n <= 0 {
		return next
	}

	sem := make(chan struct{}, n)
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("Too many concurrent requests"))
			return
		}
		next(w, r)
	}
}

type metricsHandler struct {
	timeoutMargin          float64
	includeExporterMetrics bool
	collectorFactory       func(timeout time.Duration, requestedCollectors []string) (error, *collector.Prometheus)
	logger                 log.Logger
}

func (mh *metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const defaultTimeout = 10.0

	var timeoutSeconds float64
	if v := r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); v != "" {
		var err error
		timeoutSeconds, err = strconv.ParseFloat(v, 64)
		if err != nil {
			_ = level.Warn(mh.logger).Log("msg", fmt.Sprintf("Couldn't parse X-Prometheus-Scrape-Timeout-Seconds: %q. Defaulting timeout to %f", v, defaultTimeout))
		}
	}
	if timeoutSeconds == 0 {
		timeoutSeconds = defaultTimeout
	}
	timeoutSeconds = timeoutSeconds - mh.timeoutMargin

	reg := prometheus.NewRegistry()
	err, wc := mh.collectorFactory(time.Duration(timeoutSeconds*float64(time.Second)), r.URL.Query()["collect[]"])
	if err != nil {
		_ = level.Warn(mh.logger).Log("msg", "Couldn't create filtered metrics handler", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Couldn't create filtered metrics handler: %s", err))) //nolint:errcheck
		return
	}
	reg.MustRegister(wc)
	if !mh.includeExporterMetrics {
		reg.MustRegister(
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewGoCollector(),
			version.NewCollector("windows_exporter"),
		)
	}

	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorLog: stdlog.New(log.NewStdlibAdapter(level.Error(mh.logger)), "", stdlog.Lshortfile),
	})
	h.ServeHTTP(w, r)
}
