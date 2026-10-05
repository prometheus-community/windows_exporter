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

package hyperv

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// wmiHealthQueryTimeout bounds the WMI health query, so a hung WMI service
// is reported as unhealthy instead of blocking the scrape.
const wmiHealthQueryTimeout = 5 * time.Second

// collectorWMIHealth Hyper-V WMI health metrics
type collectorWMIHealth struct {
	miQuery mi.Query

	wmiHealth *prometheus.Desc
}

// msvmComputerSystem is the Msvm_ComputerSystem WMI class.
// Only Name is needed to verify the Hyper-V WMI provider is working.
type msvmComputerSystem struct {
	Name string `mi:"Name"`
}

func (c *Collector) buildWMIHealth() error {
	if c.miSession == nil {
		return mi.ErrNotInitialized
	}

	miQuery, err := mi.NewQuery("SELECT Name FROM Msvm_ComputerSystem WHERE Caption = 'Hosting Computer System'")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.miQuery = miQuery

	c.wmiHealth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "wmi_health"),
		"Hyper-V WMI health status. 1 if the Hyper-V WMI namespace is responding, 0 if it is broken or not responding",
		nil,
		nil,
	)

	return nil
}

// collectWMIHealth reports 1 if a minimal query against the Hyper-V WMI
// namespace succeeds and 0 otherwise.
func (c *Collector) collectWMIHealth(ch chan<- prometheus.Metric) error {
	var dst []msvmComputerSystem

	healthValue := 1.0

	if err := c.miSession.Query(&dst, mi.NamespaceRootVirtualizationV2, c.miQuery, wmiHealthQueryTimeout); err != nil {
		c.logger.Debug("Hyper-V WMI health query failed", slog.Any("err", err))

		healthValue = 0.0
	}

	ch <- prometheus.MustNewConstMetric(
		c.wmiHealth,
		prometheus.GaugeValue,
		healthValue,
	)

	return nil
}
