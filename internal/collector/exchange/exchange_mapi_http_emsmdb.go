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

const (
	activeUserCount = "Active User Count"
)

func (c *Collector) buildMapiHttpEmsmdb() error {
	counters := []string{
		activeUserCount,
	}

	var err error

	c.perfDataCollectorMapiHttpEmsmdb, err = pdh.NewCollector("MSExchange MapiHttp Emsmdb", pdh.InstancesAll, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange MapiHttp Emsmdb: %w", err)
	}

	c.activeUserCountMapiHttpEmsMDB = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "mapihttp_emsmdb_active_user_count"),
		"Number of unique outlook users that have shown some kind of activity in the last 2 minutes",
		nil,
		nil,
	)

	return nil
}

func (c *Collector) collectMapiHttpEmsmdb(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorMapiHttpEmsmdb.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange MapiHttp Emsmdb metrics: %w", err)
	}

	if len(perfData) == 0 {
		return fmt.Errorf("failed to collect MSExchange MapiHttp Emsmdb metrics: %w", types.ErrNoData)
	}

	for _, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.activeUserCountMapiHttpEmsMDB,
			prometheus.GaugeValue,
			data[activeUserCount].FirstValue,
		)
	}

	return nil
}
