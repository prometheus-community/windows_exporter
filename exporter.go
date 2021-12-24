//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/windows/svc"

	"github.com/StackExchange/wmi"
	"github.com/prometheus-community/windows_exporter/collector"
	"github.com/prometheus-community/windows_exporter/config"
	"github.com/prometheus-community/windows_exporter/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"gopkg.in/alecthomas/kingpin.v2"
)

type windowsCollector struct {
	maxScrapeDuration time.Duration
	collectors        map[string]collector.Collector
}

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
	defaultCollectors            = "cpu,cs,logical_disk,net,os,service,system,textfile"
	defaultCollectorsPlaceholder = "[defaults]"
	serviceName                  = "windows_exporter"
)

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, "exporter", "collector_duration_seconds"),
		"windows_exporter: Duration of a collection.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, "exporter", "collector_success"),
		"windows_exporter: Whether the collector was successful.",
		[]string{"collector"},
		nil,
	)
	scrapeTimeoutDesc = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, "exporter", "collector_timeout"),
		"windows_exporter: Whether the collector timed out.",
		[]string{"collector"},
		nil,
	)
	snapshotDuration = prometheus.NewDesc(
		prometheus.BuildFQName(collector.Namespace, "exporter", "perflib_snapshot_duration_seconds"),
		"Duration of perflib snapshot capture",
		nil,
		nil,
	)
)

// Describe sends all the descriptors of the collectors included to
// the provided channel.
func (coll windowsCollector) Describe(ch chan<- *prometheus.Desc) {
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
func (coll windowsCollector) Collect(ch chan<- prometheus.Metric) {
	t := time.Now()
	cs := make([]string, 0, len(coll.collectors))
	for name := range coll.collectors {
		cs = append(cs, name)
	}
	scrapeContext, err := collector.PrepareScrapeContext(cs)
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
		go func(name string, c collector.Collector) {
			defer wg.Done()
			outcome := execute(name, c, scrapeContext, metricsBuffer)
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
		log.Warn("Collection timed out, still waiting for ", remainingCollectorNames)
	}

	l.Unlock()
}

func execute(name string, c collector.Collector, ctx *collector.ScrapeContext, ch chan<- prometheus.Metric) collectorOutcome {
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
		log.Errorf("collector %s failed after %fs: %s", name, duration, err)
		return failed
	}
	log.Debugf("collector %s succeeded after %fs.", name, duration)
	return success
}

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

func loadCollectors(list string) (map[string]collector.Collector, error) {
	collectors := map[string]collector.Collector{}
	enabled := expandEnabledCollectors(list)

	for _, name := range enabled {
		c, err := collector.Build(name)
		if err != nil {
			return nil, err
		}
		collectors[name] = c
	}

	return collectors, nil
}

func initWbem() {
	// This initialization prevents a memory leak on WMF 5+. See
	// https://github.com/prometheus-community/windows_exporter/issues/77 and
	// linked issues for details.
	log.Debugf("Initializing SWbemServices")
	s, err := wmi.InitializeSWbemServices(wmi.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}
	wmi.DefaultClient.AllowMissingFields = true
	wmi.DefaultClient.SWbemServicesClient = s
}

func main() {
	var (
		configFile = kingpin.Flag(
			"config.file",
			"YAML configuration file to use. Values set in this file will be overridden by CLI flags.",
		).String()
		webConfig     = webflag.AddFlags(kingpin.CommandLine)
		listenAddress = kingpin.Flag(
			"telemetry.addr",
			"host:port for exporter.",
		).Default(":9182").String()
		metricsPath = kingpin.Flag(
			"telemetry.path",
			"URL path for surfacing collected metrics.",
		).Default("/metrics").String()
		maxRequests = kingpin.Flag(
			"telemetry.max-requests",
			"Maximum number of concurrent requests. 0 to disable.",
		).Default("5").Int()
		enabledCollectors = kingpin.Flag(
			"collectors.enabled",
			"Comma-separated list of collectors to use. Use '[defaults]' as a placeholder for all the collectors enabled by default.").
			Default(defaultCollectors).String()
		printCollectors = kingpin.Flag(
			"collectors.print",
			"If true, print available collectors and exit.",
		).Bool()
		timeoutMargin = kingpin.Flag(
			"scrape.timeout-margin",
			"Seconds to subtract from the timeout allowed by the client. Tune to allow for overhead or high loads.",
		).Default("0.5").Float64()
	)

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("windows_exporter"))
	kingpin.HelpFlag.Short('h')

	// Load values from configuration file(s). Executable flags must first be parsed, in order
	// to load the specified file(s).
	kingpin.Parse()

	if *configFile != "" {
		resolver, err := config.NewResolver(*configFile)
		if err != nil {
			log.Fatalf("could not load config file: %v\n", err)
		}
		err = resolver.Bind(kingpin.CommandLine, os.Args[1:])
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		// Parse flags once more to include those discovered in configuration file(s).
		kingpin.Parse()
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

	initWbem()

	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatal(err)
	}

	stopCh := make(chan bool)
	if isService {
		go func() {
			err = svc.Run(serviceName, &windowsExporterService{stopCh: stopCh})
			if err != nil {
				log.Errorf("Failed to start service: %v", err)
			}
		}()
	}

	collectors, err := loadCollectors(*enabledCollectors)
	if err != nil {
		log.Fatalf("Couldn't load collectors: %s", err)
	}

	log.Infof("Enabled collectors: %v", strings.Join(keys(collectors), ", "))

	h := &metricsHandler{
		timeoutMargin: *timeoutMargin,
		collectorFactory: func(timeout time.Duration, requestedCollectors []string) (error, *windowsCollector) {
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
			return nil, &windowsCollector{
				collectors:        filteredCollectors,
				maxScrapeDuration: timeout,
			}
		},
	}

	http.HandleFunc(*metricsPath, withConcurrencyLimit(*maxRequests, h.ServeHTTP))
	http.HandleFunc("/health", healthCheck)
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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
<head><title>windows_exporter</title></head>
<body>
<h1>windows_exporter</h1>
<p><a href="` + *metricsPath + `">Metrics</a></p>
<p><i>` + version.Info() + `</i></p>
</body>
</html>`))
	})

	log.Infoln("Starting windows_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	go func() {
		log.Infoln("Starting server on", *listenAddress)
		server := &http.Server{Addr: *listenAddress}
		if err := web.ListenAndServe(server, *webConfig, log.NewToolkitAdapter()); err != nil {
			log.Fatalf("cannot start windows_exporter: %s", err)
		}
	}()

	for {
		if <-stopCh {
			log.Info("Shutting down windows_exporter")
			break
		}
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := fmt.Fprintln(w, `{"status":"ok"}`)
	if err != nil {
		log.Debugf("Failed to write to stream: %v", err)
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

type windowsExporterService struct {
	stopCh chan<- bool
}

func (s *windowsExporterService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				s.stopCh <- true
				break loop
			default:
				log.Error(fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

type metricsHandler struct {
	timeoutMargin    float64
	collectorFactory func(timeout time.Duration, requestedCollectors []string) (error, *windowsCollector)
}

func (mh *metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const defaultTimeout = 10.0

	var timeoutSeconds float64
	if v := r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); v != "" {
		var err error
		timeoutSeconds, err = strconv.ParseFloat(v, 64)
		if err != nil {
			log.Warnf("Couldn't parse X-Prometheus-Scrape-Timeout-Seconds: %q. Defaulting timeout to %f", v, defaultTimeout)
		}
	}
	if timeoutSeconds == 0 {
		timeoutSeconds = defaultTimeout
	}
	timeoutSeconds = timeoutSeconds - mh.timeoutMargin

	reg := prometheus.NewRegistry()
	err, wc := mh.collectorFactory(time.Duration(timeoutSeconds*float64(time.Second)), r.URL.Query()["collect[]"])
	if err != nil {
		log.Warnln("Couldn't create filtered metrics handler: ", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Couldn't create filtered metrics handler: %s", err))) //nolint:errcheck
		return
	}
	reg.MustRegister(wc)
	reg.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
		version.NewCollector("windows_exporter"),
	)

	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}
