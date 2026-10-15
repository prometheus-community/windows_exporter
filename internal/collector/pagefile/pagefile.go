// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
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

package pagefile

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/psapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "pagefile"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI metrics.
type Collector struct {
	config Config
	logger *slog.Logger

	perfDataCollector *pdh.Collector[perfDataCounterValues]
	perfDataObject    []perfDataCounterValues

	pagingFreeBytes  *prometheus.Desc
	pagingLimitBytes *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	c.perfDataCollector.Close()

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))

	if err := psapi.Load(); err != nil {
		return err
	}

	c.pagingLimitBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "limit_bytes"),
		"Current size of the paging file in bytes",
		[]string{"file"},
		nil,
	)

	c.pagingFreeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "free_bytes"),
		"Number of bytes that can be mapped into the operating system paging files without causing any other pages to be swapped out",
		[]string{"file"},
		nil,
	)

	var err error

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](c.logger, pdh.CounterTypeRaw, "Paging File", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Paging File collector: %w", err)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect Paging File metrics: %w", err)
	}

	gpi, err := psapi.GetPerformanceInfo()
	if err != nil {
		return err
	}

	c.collectPagingFiles(ch, c.perfDataObject, float64(gpi.PageSize))

	return nil
}

// collectPagingFiles publishes the size and free space of each paging file.
// Both come from "% Usage": its raw value is the number of pages in use and its
// base is the size of the paging file in pages.
func (c *Collector) collectPagingFiles(ch chan<- prometheus.Metric, data []perfDataCounterValues, pageSize float64) {
	for _, file := range data {
		// A paging file without a size has no valid sample. Publishing it would
		// report a size of 0 and a negative free space.
		if file.UsageBase <= 0 || file.Usage < 0 || file.Usage > file.UsageBase {
			c.logger.Debug("skipping paging file without a valid size",
				slog.String("file", file.Name),
				slog.Float64("used_pages", file.Usage),
				slog.Float64("size_pages", file.UsageBase),
			)

			continue
		}

		fileName := strings.ReplaceAll(file.Name, `\??\`, "")

		ch <- prometheus.MustNewConstMetric(
			c.pagingFreeBytes,
			prometheus.GaugeValue,
			(file.UsageBase-file.Usage)*pageSize,
			fileName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.pagingLimitBytes,
			prometheus.GaugeValue,
			file.UsageBase*pageSize,
			fileName,
		)
	}
}
