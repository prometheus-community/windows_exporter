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

type collectorOWA struct {
	perfDataCollectorOWA *pdh.Collector[perfDataCounterValuesOWA]
	perfDataObjectOWA    []perfDataCounterValuesOWA

	currentUniqueUsers *prometheus.Desc
	owaRequestsPerSec  *prometheus.Desc
}

type perfDataCounterValuesOWA struct {
	CurrentUniqueUsers float64 `perfdata:"Current Unique Users"`
	RequestsPerSec     float64 `perfdata:"Requests/sec"`
}

func (c *Collector) buildOWA() error {
	var err error

	c.perfDataCollectorOWA, err = pdh.NewCollector[perfDataCounterValuesOWA](c.logger, pdh.CounterTypeRaw, "MSExchange OWA", pdh.InstancesAll)
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
	err := c.perfDataCollectorOWA.Collect(&c.perfDataObjectOWA)
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange OWA metrics: %w", err)
	}

	for _, data := range c.perfDataObjectOWA {
		ch <- prometheus.MustNewConstMetric(
			c.currentUniqueUsers,
			prometheus.GaugeValue,
			data.CurrentUniqueUsers,
		)

		ch <- prometheus.MustNewConstMetric(
			c.owaRequestsPerSec,
			prometheus.CounterValue,
			data.RequestsPerSec,
		)
	}

	return nil
}
