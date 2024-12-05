package types

import "github.com/prometheus-community/windows_exporter/internal/pdh"

type Collector interface {
	Collect() (pdh.CounterValues, error)
	Close()
}
