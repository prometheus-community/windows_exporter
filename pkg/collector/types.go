package collector

import (
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus/client_golang/prometheus"
)

const DefaultCollectors = "cpu,cs,memory,logical_disk,physical_disk,net,os,service,system"

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
	// GetName get the name of the collector
	GetName() string
	// Build build the collector
	Build(logger *slog.Logger, miSession *mi.Session) error
	// Collect Get new metrics and expose them via prometheus registry.
	Collect(ch chan<- prometheus.Metric) (err error)
	// Close closes the collector
	Close() error
}
