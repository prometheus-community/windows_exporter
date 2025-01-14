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

type collectorActiveSync struct {
	perfDataCollectorActiveSync *pdh.Collector
	perfDataObjectActiveSync    []perfDataCounterValuesActiveSync

	activeSyncRequestsPerSec *prometheus.Desc
	pingCommandsPending      *prometheus.Desc
	syncCommandsPerSec       *prometheus.Desc
}

type perfDataCounterValuesActiveSync struct {
	RequestsPerSec      float64 `perfdata:"Requests/sec"`
	PingCommandsPending float64 `perfdata:"Ping Commands Pending"`
	SyncCommandsPerSec  float64 `perfdata:"Sync Commands/sec"`
}

func (c *Collector) buildActiveSync() error {
	var err error

	c.perfDataCollectorActiveSync, err = pdh.NewCollector[perfDataCounterValuesActiveSync](pdh.CounterTypeRaw, "MSExchange ActiveSync", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange ActiveSync collector: %w", err)
	}

	c.pingCommandsPending = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "activesync_ping_cmds_pending"),
		"Number of ping commands currently pending in the queue",
		nil,
		nil,
	)
	c.syncCommandsPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "activesync_sync_cmds_total"),
		"Number of sync commands processed per second. Clients use this command to synchronize items within a folder",
		nil,
		nil,
	)
	c.activeSyncRequestsPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "activesync_requests_total"),
		"Num HTTP requests received from the client via ASP.NET per sec. Shows Current user load",
		nil,
		nil,
	)

	return nil
}

func (c *Collector) collectActiveSync(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorActiveSync.Collect(&c.perfDataObjectActiveSync)
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange ActiveSync metrics: %w", err)
	}

	for _, data := range c.perfDataObjectActiveSync {
		ch <- prometheus.MustNewConstMetric(
			c.activeSyncRequestsPerSec,
			prometheus.CounterValue,
			data.RequestsPerSec,
		)
		ch <- prometheus.MustNewConstMetric(
			c.pingCommandsPending,
			prometheus.GaugeValue,
			data.PingCommandsPending,
		)
		ch <- prometheus.MustNewConstMetric(
			c.syncCommandsPerSec,
			prometheus.CounterValue,
			data.SyncCommandsPerSec,
		)
	}

	return nil
}
