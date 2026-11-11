// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package collector

import (
	"log/slog"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus/client_golang/prometheus"
)

const DefaultCollectors = "cpu,memory,logical_disk,physical_disk,net,os,service,system"

type Collection struct {
	collectors Map
	startTime  time.Time

	// state is shared with every Collection derived from this one by WithCollectors.
	state *collectionState
	// isView marks a non-owning view returned by WithCollectors.
	isView bool

	scrapeDurationDesc          *prometheus.Desc
	scrapeQueueTimeoutDesc      *prometheus.Desc
	collectorScrapeDurationDesc *prometheus.Desc
	collectorScrapeSuccessDesc  *prometheus.Desc
	collectorScrapeTimeoutDesc  *prometheus.Desc
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
	Collect(ch chan<- prometheus.Metric, maxScrapeDuration time.Duration) (err error)
	// Close closes the collector
	Close() error
}
