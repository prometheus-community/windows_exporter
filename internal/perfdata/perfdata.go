package perfdata

import (
	"errors"

	"github.com/prometheus-community/windows_exporter/internal/perfdata/pdh"
	"github.com/prometheus-community/windows_exporter/internal/perfdata/perftypes"
	"github.com/prometheus-community/windows_exporter/internal/perfdata/registry"
)

type Collector interface {
	Describe() map[string]string
	Collect() (map[string]map[string]perftypes.CounterValues, error)
	Close()
}

type Engine int

const (
	_ Engine = iota
	PDH
	Registry
)

var ErrUnknownEngine = errors.New("unknown engine")
var AllInstances = []string{"*"}

func NewCollector(engine Engine, object string, instances []string, counters []string) (Collector, error) {
	switch engine {
	case PDH:
		return pdh.NewCollector(object, instances, counters)
	case Registry:
		return registry.NewCollector(object, instances, counters)
	default:
		return nil, ErrUnknownEngine
	}
}
