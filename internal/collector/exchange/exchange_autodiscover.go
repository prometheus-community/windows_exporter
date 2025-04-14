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
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorAutoDiscover struct {
	perfDataCollectorAutoDiscover *pdh.Collector
	perfDataObjectAutoDiscover    []perfDataCounterValuesAutoDiscover

	autoDiscoverRequestsPerSec *prometheus.Desc
}

type perfDataCounterValuesAutoDiscover struct {
	RequestsPerSec float64 `perfdata:"Requests/sec"`
}

func (c *Collector) buildAutoDiscover() error {
	var err error

	c.perfDataCollectorAutoDiscover, err = pdh.NewCollector[perfDataCounterValuesAutoDiscover](pdh.CounterTypeRaw, "MSExchangeAutodiscover", nil)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange Autodiscover collector: %w", err)
	}

	c.autoDiscoverRequestsPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "autodiscover_requests_total"),
		"Number of autodiscover service requests processed each second",
		nil,
		nil,
	)

	return nil
}

func (c *Collector) collectAutoDiscover(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorAutoDiscover.Collect(&c.perfDataObjectAutoDiscover)
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange Autodiscover metrics: %w", err)
	}

	for _, data := range c.perfDataObjectAutoDiscover {
		ch <- prometheus.MustNewConstMetric(
			c.autoDiscoverRequestsPerSec,
			prometheus.CounterValue,
			data.RequestsPerSec,
		)
	}

	return nil
}
