package perfdata

import (
	"errors"

	"github.com/prometheus-community/windows_exporter/internal/perfdata/perftypes"
	"github.com/prometheus-community/windows_exporter/internal/perfdata/v1"
	"github.com/prometheus-community/windows_exporter/internal/perfdata/v2"
)

type Collector interface {
	Describe() map[string]string
	Collect() (map[string]map[string]perftypes.CounterValues, error)
	Close()
}

type Engine int

const (
	_ Engine = iota
	V1
	V2
)

var ErrUnknownEngine = errors.New("unknown engine")
var AllInstances = []string{"*"}

func NewCollector(engine Engine, object string, instances []string, counters []string) (Collector, error) {
	switch engine {
	case V1:
		return v1.NewCollector(object, instances, counters)
	case V2:
		return v2.NewCollector(object, instances, counters)
	default:
		return nil, ErrUnknownEngine
	}
}
