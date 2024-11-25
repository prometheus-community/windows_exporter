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

package license

import (
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/slc"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "license"

//nolint:gochecknoglobals
var labelMap = map[slc.SL_GENUINE_STATE]string{
	slc.SL_GEN_STATE_IS_GENUINE:      "genuine",
	slc.SL_GEN_STATE_INVALID_LICENSE: "invalid_license",
	slc.SL_GEN_STATE_TAMPERED:        "tampered",
	slc.SL_GEN_STATE_OFFLINE:         "offline",
	slc.SL_GEN_STATE_LAST:            "last",
}

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI Win32_PerfRawData_DNS_DNS metrics.
type Collector struct {
	config Config

	licenseStatus *prometheus.Desc
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
	c.licenseStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "status"),
		"Status of windows license",
		[]string{"state"},
		nil,
	)

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	status, err := slc.SLIsWindowsGenuineLocal()
	if err != nil {
		return err
	}

	for k, v := range labelMap {
		val := 0.0
		if status == k {
			val = 1.0
		}

		ch <- prometheus.MustNewConstMetric(c.licenseStatus, prometheus.GaugeValue, val, v)
	}

	return nil
}
