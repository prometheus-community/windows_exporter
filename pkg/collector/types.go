package collector

import (
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricCollectors struct {
	Collectors       Map
	MISession        *mi.Session
	PerfCounterQuery string
}

type (
	BuilderWithFlags[C Collector] func(*kingpin.Application) C
	Map                           map[string]Collector
)

// Collector interface that a collector has to implement.
type Collector interface {
	Build(logger *slog.Logger, miSession *mi.Session) error
	// Close closes the collector
	Close(logger *slog.Logger) error
	// GetName get the name of the collector
	GetName() string
	// GetPerfCounter returns the perf counter required by the collector
	GetPerfCounter(logger *slog.Logger) ([]string, error)
	// Collect Get new metrics and expose them via prometheus registry.
	Collect(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) (err error)
}
