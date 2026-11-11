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

type collectorAutoDiscover struct {
	perfDataCollectorAutoDiscover *pdh.Collector[perfDataCounterValuesAutoDiscover]
	perfDataObjectAutoDiscover    []perfDataCounterValuesAutoDiscover

	autoDiscoverRequestsPerSec *prometheus.Desc
}

type perfDataCounterValuesAutoDiscover struct {
	RequestsPerSec float64 `perfdata:"Requests/sec"`
}

func (c *Collector) buildAutoDiscover() error {
	var err error

	c.perfDataCollectorAutoDiscover, err = pdh.NewCollector[perfDataCounterValuesAutoDiscover](c.logger, pdh.CounterTypeRaw, "MSExchangeAutodiscover", nil)
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
