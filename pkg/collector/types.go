package collector

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

type Collectors struct {
	logger log.Logger

	collectors       Map
	perfCounterQuery string
}

type Map map[string]Collector

type (
	Builder                       func(logger log.Logger) Collector
	BuilderWithFlags[C Collector] func(*kingpin.Application) C
)

// Collector interface that a collector has to implement.
type Collector interface {
	Build() error
	// Close closes the collector
	Close() error
	// GetName get the name of the collector
	GetName() string
	// GetPerfCounter returns the perf counter required by the collector
	GetPerfCounter() ([]string, error)
	// Collect Get new metrics and expose them via prometheus registry.
	Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) (err error)
	SetLogger(logger log.Logger)
}
