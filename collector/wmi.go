package collector

import "github.com/prometheus/client_golang/prometheus"

// ...
const (
	Namespace = "wmi"
)

// Factories ...
var Factories = make(map[string]func() (Collector, error))

// Collector is the interface a collector has to implement.
type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Collect(ch chan<- prometheus.Metric) (err error)
}
