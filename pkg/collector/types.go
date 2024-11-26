// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package collector

import (
	"log/slog"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus/client_golang/prometheus"
)

const DefaultCollectors = "cpu,cs,memory,logical_disk,physical_disk,net,os,service,system"

type Collection struct {
	collectors    Map
	miSession     *mi.Session
	startTime     time.Time
	concurrencyCh chan struct{}

	scrapeDurationDesc          *prometheus.Desc
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
	Collect(ch chan<- prometheus.Metric) (err error)
	// Close closes the collector
	Close() error
}
