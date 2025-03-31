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

package pagefile

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

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

	perfDataCollector *pdh.Collector
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

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	c.pagingLimitBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "limit_bytes"),
		"Number of bytes that can be stored in the operating system paging files. 0 (zero) indicates that there are no paging files",
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

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "Paging File", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Paging File collector: %w", err)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect Paging File metrics: %w", err)
	}

	gpi, err := psapi.GetPerformanceInfo()
	if err != nil {
		return err
	}

	for _, data := range c.perfDataObject {
		fileString := strings.ReplaceAll(data.Name, `\??\`, "")
		file, err := os.Stat(fileString)

		var fileSize float64

		// For unknown reasons, Windows doesn't always create a page file. Continue collection rather than aborting.
		if err == nil {
			fileSize = float64(file.Size())
		}

		ch <- prometheus.MustNewConstMetric(
			c.pagingFreeBytes,
			prometheus.GaugeValue,
			fileSize-(data.Usage*float64(gpi.PageSize)),
			fileString,
		)

		ch <- prometheus.MustNewConstMetric(
			c.pagingLimitBytes,
			prometheus.GaugeValue,
			fileSize,
			fileString,
		)
	}

	return nil
}
