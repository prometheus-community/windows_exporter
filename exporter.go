package main

import (
	"flag"
	"log"
	"net/http"
	"sync"

	"github.com/martinlindhe/wmi_exporter/collectors"
	"github.com/prometheus/client_golang/prometheus"
)

// WmiExporter wraps all the WMI collectors and provides a single global
// exporter to extracts metrics out of. It also ensures that the collection
// is done in a thread-safe manner, the necessary requirement stated by
// prometheus. It also implements a prometheus.Collector interface in order
// to register it correctly.
type WmiExporter struct {
	mu         sync.Mutex
	collectors []prometheus.Collector
}

// Verify that the WmiExporter implements the prometheus.Collector interface.
var _ prometheus.Collector = &WmiExporter{}

// NewWmiExporter creates an instance to WmiExporter and returns a reference
// to it. We can choose to enable a collector to extract stats out of by adding
// it to the list of collectors.
func NewWmiExporter() *WmiExporter {
	return &WmiExporter{
		collectors: []prometheus.Collector{
			collectors.NewOSCollector(),
			collectors.NewLogicalDiskCollector(),
		},
	}
}

// Describe sends all the descriptors of the collectors included to
// the provided channel.
func (c *WmiExporter) Describe(ch chan<- *prometheus.Desc) {
	for _, cc := range c.collectors {
		cc.Describe(ch)
	}
}

// Collect sends the collected metrics from each of the collectors to
// prometheus. Collect could be called several times concurrently
// and thus its run is protected by a single mutex.
func (c *WmiExporter) Collect(ch chan<- prometheus.Metric) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, cc := range c.collectors {
		cc.Collect(ch)
	}
}

func main() {
	var (
		addr        = flag.String("telemetry.addr", ":9182", "host:port for WMI exporter")
		metricsPath = flag.String("telemetry.path", "/metrics", "URL path for surfacing collected metrics")
	)
	flag.Parse()

	prometheus.MustRegister(NewWmiExporter())

	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, *metricsPath, http.StatusMovedPermanently)
	})

	log.Printf("Starting WMI exporter on %q", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("cannot start WMI exporter: %s", err)
	}
}
