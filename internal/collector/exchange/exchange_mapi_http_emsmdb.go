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

type collectorMapiHTTPEmsMDB struct {
	perfDataCollectorMapiHTTPEmsMDB *pdh.Collector[perfDataCounterValuesMapiHTTPEmsMDB]
	perfDataObjectMapiHTTPEmsMDB    []perfDataCounterValuesMapiHTTPEmsMDB

	activeUserCountMapiHTTPEmsMDB *prometheus.Desc
}

type perfDataCounterValuesMapiHTTPEmsMDB struct {
	ActiveUserCount float64 `perfdata:"Active User Count"`
}

func (c *Collector) buildMapiHTTPEmsMDB() error {
	var err error

	c.perfDataCollectorMapiHTTPEmsMDB, err = pdh.NewCollector[perfDataCounterValuesMapiHTTPEmsMDB](c.logger, pdh.CounterTypeRaw, "MSExchange MapiHttp Emsmdb", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange MapiHttp Emsmdb: %w", err)
	}

	c.activeUserCountMapiHTTPEmsMDB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mapihttp_emsmdb_active_users"),
		"Number of unique outlook users that have shown some kind of activity in the last 2 minutes",
		nil,
		nil,
	)

	return nil
}

func (c *Collector) collectMapiHTTPEmsMDB(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorMapiHTTPEmsMDB.Collect(&c.perfDataObjectMapiHTTPEmsMDB)
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange MapiHttp Emsmdb metrics: %w", err)
	}

	for _, data := range c.perfDataObjectMapiHTTPEmsMDB {
		ch <- prometheus.MustNewConstMetric(
			c.activeUserCountMapiHTTPEmsMDB,
			prometheus.GaugeValue,
			data.ActiveUserCount,
		)
	}

	return nil
}
