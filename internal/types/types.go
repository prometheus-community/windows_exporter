package types

import (
	v1 "github.com/prometheus-community/windows_exporter/internal/perfdata/v1"
)

type ScrapeContext struct {
	PerfObjects map[string]*v1.PerfObject
}
