package collector

import (
	"github.com/leoluk/perflib_exporter/perflib"
	"github.com/prometheus/client_golang/prometheus"
)

// ...
const (
	// TODO: Make package-local
	Namespace = "wmi"

	// Conversion factors
	ticksToSecondsScaleFactor = 1 / 1e7
	windowsEpoch              = 116444736000000000
)

// Factories ...
var Factories = make(map[string]func() (Collector, error))

// Collector is the interface a collector has to implement.
type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (err error)
}

type ScrapeContext struct {
	perfObjects map[string]*perflib.PerfObject
}

// PrepareScrapeContext creates a ScrapeContext to be used during a single scrape
func PrepareScrapeContext() (*ScrapeContext, error) {
	objs, err := getPerflibSnapshot()
	if err != nil {
		return nil, err
	}

	return &ScrapeContext{objs}, nil
}
