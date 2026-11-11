// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package exchange

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorAvailabilityService struct {
	perfDataCollectorAvailabilityService *pdh.Collector[perfDataCounterValuesAvailabilityService]
	perfDataObjectAvailabilityService    []perfDataCounterValuesAvailabilityService

	availabilityRequestsSec *prometheus.Desc
}

type perfDataCounterValuesAvailabilityService struct {
	AvailabilityRequestsPerSec float64 `perfdata:"Availability Requests (sec)"`
}

func (c *Collector) buildAvailabilityService() error {
	var err error

	c.perfDataCollectorAvailabilityService, err = pdh.NewCollector[perfDataCounterValuesAvailabilityService](c.logger, pdh.CounterTypeRaw, "MSExchange Availability Service", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange Availability Service collector: %w", err)
	}

	c.availabilityRequestsSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availability_service_requests_total"),
		"Total number of requests serviced",
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
