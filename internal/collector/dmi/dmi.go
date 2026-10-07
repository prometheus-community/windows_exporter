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

package dmi

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/sysinfoapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "dmi"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for DMI (SMBIOS) metrics.
type Collector struct {
	config Config

	dmiInfo *prometheus.Desc
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

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	info, err := sysinfoapi.GetSMBIOSSystemInfo()
	if err != nil {
		return fmt.Errorf("SMBIOS system info unavailable: %w", err)
	}

	// Sanitize label values: firmware strings may contain invalid UTF-8 bytes.
	// Use "�" as replacement to match node_exporter's behavior.
	cleanLabel := func(s string) string {
		return strings.ToValidUTF8(strings.TrimSpace(s), "�")
	}

	desc := prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"A metric with a constant '1' value labeled by product_name, product_serial, "+
			"product_uuid, product_version, system_vendor from SMBIOS Type 1 (System Information).",
		nil,
		prometheus.Labels{
			"product_name":    cleanLabel(info.ProductName),
			"product_serial":  cleanLabel(info.SerialNumber),
			"product_uuid":    strings.ToLower(cleanLabel(info.UUID)),
			"product_version": cleanLabel(info.Version),
			"system_vendor":   cleanLabel(info.Manufacturer),
		},
	)

	// Validate the descriptor by attempting to create a metric.
	// If label values are still invalid, log a warning and skip.
	if _, err = prometheus.NewConstMetric(desc, prometheus.GaugeValue, 1.0); err != nil {
		logger.Warn("SMBIOS labels produced an invalid metric, skipping dmi_info", slog.Any("err", err))

		return nil
	}

	c.dmiInfo = desc

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	if c.dmiInfo != nil {
		ch <- prometheus.MustNewConstMetric(
			c.dmiInfo,
			prometheus.GaugeValue,
			1.0,
		)
	}

	return nil
}
