package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/windows/svc"

	"github.com/martinlindhe/wmi_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

// WmiCollector implements the prometheus.Collector interface.
type WmiCollector struct {
	collectors map[string]collector.Collector
}

const (
	defaultCollectors = "cpu,logical_disk,os"
	serviceName       = "wmi_exporter"
)

var (
	scrapeDurations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: collector.Namespace,
			Subsystem: "exporter",
			Name:      "scrape_duration_seconds",
			Help:      "wmi_exporter: Duration of a scrape job.",
		},
		[]string{"collector", "result"},
	)
)

// Describe sends all the descriptors of the collectors included to
// the provided channel.
func (coll WmiCollector) Describe(ch chan<- *prometheus.Desc) {
	scrapeDurations.Describe(ch)
}

// Collect sends the collected metrics from each of the collectors to
// prometheus. Collect could be called several times concurrently
// and thus its run is protected by a single mutex.
func (coll WmiCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(coll.collectors))
	for name, c := range coll.collectors {
		go func(name string, c collector.Collector) {
			execute(name, c, ch)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
	scrapeDurations.Collect(ch)
}

func filterAvailableCollectors(collectors string) string {
	var availableCollectors []string
	for _, c := range strings.Split(collectors, ",") {
		_, ok := collector.Factories[c]
		if ok {
			availableCollectors = append(availableCollectors, c)
		}
	}
	return strings.Join(availableCollectors, ",")
}

func execute(name string, c collector.Collector, ch chan<- prometheus.Metric) {
	begin := time.Now()
	err := c.Collect(ch)
	duration := time.Since(begin)
	var result string

	if err != nil {
		log.Errorf("ERROR: %s collector failed after %fs: %s", name, duration.Seconds(), err)
		result = "error"
	} else {
		log.Debugf("OK: %s collector succeeded after %fs.", name, duration.Seconds())
		result = "success"
	}
	scrapeDurations.WithLabelValues(name, result).Observe(duration.Seconds())
}

func loadCollectors(list string) (map[string]collector.Collector, error) {
	collectors := map[string]collector.Collector{}
	for _, name := range strings.Split(list, ",") {
		fn, ok := collector.Factories[name]
		if !ok {
			return nil, fmt.Errorf("collector '%s' not available", name)
		}
		c, err := fn()
		if err != nil {
			return nil, err
		}
		collectors[name] = c
	}
	return collectors, nil
}

func init() {
	prometheus.MustRegister(version.NewCollector("wmi_exporter"))
}

func main() {
	var (
		showVersion       = flag.Bool("version", false, "Print version information.")
		listenAddress     = flag.String("telemetry.addr", ":9182", "host:port for WMI exporter.")
		metricsPath       = flag.String("telemetry.path", "/metrics", "URL path for surfacing collected metrics.")
		enabledCollectors = flag.String("collectors.enabled", filterAvailableCollectors(defaultCollectors), "Comma-separated list of collectors to use.")
		printCollectors   = flag.Bool("collectors.print", false, "If true, print available collectors and exit.")
	)
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("wmi_exporter"))
		os.Exit(0)
	}

	if *printCollectors {
		collectorNames := make(sort.StringSlice, 0, len(collector.Factories))
		for n := range collector.Factories {
			collectorNames = append(collectorNames, n)
		}
		collectorNames.Sort()
		fmt.Printf("Available collectors:\n")
		for _, n := range collectorNames {
			fmt.Printf(" - %s\n", n)
		}
		return
	}

	isInteractive, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatal(err)
	}

	stopCh := make(chan bool)
	if !isInteractive {
		go svc.Run(serviceName, &wmiExporterService{stopCh: stopCh})
	}

	collectors, err := loadCollectors(*enabledCollectors)
	if err != nil {
		log.Fatalf("Couldn't load collectors: %s", err)
	}

	log.Infof("Enabled collectors: %v", strings.Join(keys(collectors), ", "))

	nodeCollector := WmiCollector{collectors: collectors}
	prometheus.MustRegister(nodeCollector)

	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, *metricsPath, http.StatusMovedPermanently)
	})

	log.Infoln("Starting WMI exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	go func() {
		log.Infoln("Starting server on", *listenAddress)
		if err := http.ListenAndServe(*listenAddress, nil); err != nil {
			log.Fatalf("cannot start WMI exporter: %s", err)
		}
	}()

	for {
		if <-stopCh {
			log.Info("Shutting down WMI exporter")
			break
		}
	}
}

func keys(m map[string]collector.Collector) []string {
	ret := make([]string, 0, len(m))
	for key, _ := range m {
		ret = append(ret, key)
	}
	return ret
}

type wmiExporterService struct {
	stopCh chan<- bool
}

func (s *wmiExporterService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
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
