//go:build windows

package types

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus/client_golang/prometheus"
)

type CollectorBuilder func(logger log.Logger) Collector
type CollectorBuilderWithFlags func(*kingpin.Application) Collector

// Collector is the interface a collector has to implement.
type Collector interface {
	Build() error
	// GetName get the name of the collector
	GetName() string
	// GetPerfCounter returns the perf counter required by the collector
	GetPerfCounter() ([]string, error)
	// Collect Get new metrics and expose them via prometheus registry.
	Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (err error)
	SetLogger(logger log.Logger)
}

type ScrapeContext struct {
	PerfObjects map[string]*perflib.PerfObject
}
