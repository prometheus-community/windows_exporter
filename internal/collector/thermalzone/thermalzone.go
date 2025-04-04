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

package thermalzone

import (
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "thermalzone"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_Counters_ThermalZoneInformation metrics.
type Collector struct {
	config Config

	perfDataCollector *pdh.Collector
	perfDataObject    []perfDataCounterValues

	percentPassiveLimit *prometheus.Desc
	temperature         *prometheus.Desc
	throttleReasons     *prometheus.Desc
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
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	c.temperature = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "temperature_celsius"),
		"(Temperature)",
		[]string{
			"name",
		},
		nil,
	)
	c.percentPassiveLimit = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "percent_passive_limit"),
		"(PercentPassiveLimit)",
		[]string{
			"name",
		},
		nil,
	)
	c.throttleReasons = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "throttle_reasons"),
		"(ThrottleReasons)",
		[]string{
			"name",
		},
		nil,
	)

	var err error

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "Thermal Zone Information", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Thermal Zone Information collector: %w", err)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect Thermal Zone Information metrics: %w", err)
	}

	for _, data := range c.perfDataObject {
		// Divide by 10 and subtract 273.15 to convert decikelvin to celsius
		ch <- prometheus.MustNewConstMetric(
			c.temperature,
			prometheus.GaugeValue,
			(data.HighPrecisionTemperature/10.0)-273.15,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.percentPassiveLimit,
			prometheus.GaugeValue,
			data.PercentPassiveLimit,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.throttleReasons,
			prometheus.GaugeValue,
			data.ThrottleReasons,
			data.Name,
		)
	}

	return nil
}
