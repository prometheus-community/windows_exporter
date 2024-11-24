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

package exchange

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	currentUniqueUsers = "Current Unique Users"
	// requestsPerSec     = "Requests/sec"
)

func (c *Collector) buildOWA() error {
	counters := []string{
		currentUniqueUsers,
		requestsPerSec,
	}

	var err error

	c.perfDataCollectorOWA, err = perfdata.NewCollector("MSExchange OWA", perfdata.InstancesAll, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange OWA collector: %w", err)
	}

	c.currentUniqueUsers = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "owa_current_unique_users"),
		"Number of unique users currently logged on to Outlook Web App",
		nil,
		nil,
	)
	c.owaRequestsPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "owa_requests_total"),
		"Number of requests handled by Outlook Web App per second",
		nil,
		nil,
	)

	return nil
}

func (c *Collector) collectOWA(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorOWA.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange OWA metrics: %w", err)
	}

	if len(perfData) == 0 {
		return errors.New("perflib query for MSExchange OWA returned empty result set")
	}

	for _, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.currentUniqueUsers,
			prometheus.GaugeValue,
			data[currentUniqueUsers].FirstValue,
		)
		ch <- prometheus.MustNewConstMetric(
			c.owaRequestsPerSec,
			prometheus.CounterValue,
			data[requestsPerSec].FirstValue,
		)
	}

	return nil
}
