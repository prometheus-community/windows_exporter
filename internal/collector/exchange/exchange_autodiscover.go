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

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/pkg/public"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Collector) buildAutoDiscover() error {
	counters := []string{
		requestsPerSec,
	}

	var err error

	c.perfDataCollectorAutoDiscover, err = perfdata.NewCollector("MSExchange Autodiscover", perfdata.InstancesAll, counters)
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
	perfData, err := c.perfDataCollectorAutoDiscover.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange Autodiscover metrics: %w", err)
	}

	if len(perfData) == 0 {
		return fmt.Errorf("failed to collect MSExchange Autodiscover metrics: %w", public.ErrNoData)
	}

	for _, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.autoDiscoverRequestsPerSec,
			prometheus.CounterValue,
			data[requestsPerSec].FirstValue,
		)
	}

	return nil
}
