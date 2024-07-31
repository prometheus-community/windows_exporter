package types

import "github.com/prometheus-community/windows_exporter/pkg/perflib"

type ScrapeContext struct {
	PerfObjects map[string]*perflib.PerfObject
}
