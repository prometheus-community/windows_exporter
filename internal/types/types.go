package types

import (
	"github.com/prometheus-community/windows_exporter/internal/perfdata/registry"
)

type ScrapeContext struct {
	PerfObjects map[string]*registry.PerfObject
}
