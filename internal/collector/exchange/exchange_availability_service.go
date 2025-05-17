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

type collectorAvailabilityService struct {
	perfDataCollectorAvailabilityService *pdh.Collector
	perfDataObjectAvailabilityService    []perfDataCounterValuesAvailabilityService

	availabilityRequestsSec *prometheus.Desc
}

type perfDataCounterValuesAvailabilityService struct {
	AvailabilityRequestsPerSec float64 `perfdata:"Availability Requests (sec)"`
}

func (c *Collector) buildAvailabilityService() error {
	var err error

	c.perfDataCollectorAvailabilityService, err = pdh.NewCollector[perfDataCounterValuesAvailabilityService](pdh.CounterTypeRaw, "MSExchange Availability Service", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange Availability Service collector: %w", err)
	}

	c.availabilityRequestsSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availability_service_requests_per_sec"),
		"Number of requests serviced per second",
		nil,
		nil,
	)

	return nil
}

func (c *Collector) collectAvailabilityService(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorAvailabilityService.Collect(&c.perfDataObjectAvailabilityService)
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange Availability Service metrics: %w", err)
	}

	for _, data := range c.perfDataObjectAvailabilityService {
		ch <- prometheus.MustNewConstMetric(
			c.availabilityRequestsSec,
			prometheus.CounterValue,
			data.AvailabilityRequestsPerSec,
		)
	}

	return nil
}
