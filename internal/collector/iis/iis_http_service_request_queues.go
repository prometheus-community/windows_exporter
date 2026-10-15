// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
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

package iis

import (
	"fmt"
	"strings"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	pdhtypes "github.com/prometheus-community/windows_exporter/internal/pdh/types"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorHttpServiceRequestQueues struct {
	perfDataCollectorHttpServiceRequestQueues pdhtypes.Collector[perfDataCounterValuesHttpServiceRequestQueues]
	perfDataObjectHttpServiceRequestQueues    []perfDataCounterValuesHttpServiceRequestQueues

	httpRequestQueuesCurrentQueueSize *prometheus.Desc
	httpRequestQueuesRejectedRequests *prometheus.Desc
	httpRequestQueuesMaxQueueItemAge  *prometheus.Desc
	httpRequestQueuesArrivedRequests  *prometheus.Desc
}

type perfDataCounterValuesHttpServiceRequestQueues struct {
	Name string

	HttpRequestQueuesCurrentQueueSize      float64 `perfdata:"CurrentQueueSize"`
	HttpRequestQueuesTotalRejectedRequests float64 `perfdata:"RejectedRequests"`
	HttpRequestQueuesMaxQueueItemAge       float64 `perfdata:"MaxQueueItemAge"`
	HttpRequestQueuesArrivalRate           float64 `perfdata:"ArrivalRate"`
}

func (p perfDataCounterValuesHttpServiceRequestQueues) GetName() string {
	return p.Name
}

func (c *Collector) buildHttpServiceRequestQueues() error {
	var err error

	c.logger.Info("IIS/HttpServiceRequestQueues collector is in an experimental state! The configuration and metrics may change in future. Please report any issues.")

	c.perfDataCollectorHttpServiceRequestQueues, err = pdh.NewCollector[perfDataCounterValuesHttpServiceRequestQueues](c.logger, pdh.CounterTypeRaw, "HTTP Service Request Queues", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Http Service collector: %w", err)
	}

	c.buildHttpServiceRequestQueuesDescriptors()

	return nil
}

func (c *Collector) buildHttpServiceRequestQueuesDescriptors() {
	// The named instances of HTTP Service Request Queues are the HTTP.sys
	// request queues of the IIS application pools, so the label is "app".
	c.httpRequestQueuesCurrentQueueSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "http_requests_current_queue_size"),
		"Number of requests in the HTTP.sys request queue of the application pool (CurrentQueueSize)",
		[]string{"app"},
		nil,
	)
	c.httpRequestQueuesRejectedRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "http_requests_rejected_total"),
		"Total number of requests rejected by the HTTP.sys request queue of the application pool (RejectedRequests)",
		[]string{"app"},
		nil,
	)
	c.httpRequestQueuesMaxQueueItemAge = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "http_requests_max_queue_item_age_seconds"),
		"Age, in seconds, of the oldest request in the HTTP.sys request queue of the application pool. Not reported while the queue is empty (MaxQueueItemAge)",
		[]string{"app"},
		nil,
	)
	c.httpRequestQueuesArrivedRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "http_requests_arrived_total"),
		"Total number of requests that arrived in the HTTP.sys request queue of the application pool (ArrivalRate)",
		[]string{"app"},
		nil,
	)
}

func (c *Collector) collectHttpServiceRequestQueues(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorHttpServiceRequestQueues.Collect(&c.perfDataObjectHttpServiceRequestQueues)
	if err != nil {
		return fmt.Errorf("failed to collect Http Service Request Queues metrics: %w", err)
	}

	c.perfDataObjectHttpServiceRequestQueues = deduplicateIISNames(c.perfDataObjectHttpServiceRequestQueues)

	for _, data := range c.perfDataObjectHttpServiceRequestQueues {
		// Present recycled instances ("Site_B#2") under their base name ("Site_B").
		data.Name = iisCounterBaseName(data.Name)

		if strings.HasPrefix(data.Name, "---") {
			continue
		}

		if c.config.AppExclude.MatchString(data.Name) || !c.config.AppInclude.MatchString(data.Name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.httpRequestQueuesCurrentQueueSize,
			prometheus.GaugeValue,
			data.HttpRequestQueuesCurrentQueueSize,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.httpRequestQueuesRejectedRequests,
			prometheus.CounterValue,
			data.HttpRequestQueuesTotalRejectedRequests,
			data.Name,
		)

		// MaxQueueItemAge is PERF_ELAPSED_TIME. On an empty queue its raw
		// value is math.MaxInt64 instead of a start time, which gives a large
		// negative elapsed time. That is no sample, so leave the series out.
		if data.HttpRequestQueuesMaxQueueItemAge >= 0 {
			ch <- prometheus.MustNewConstMetric(
				c.httpRequestQueuesMaxQueueItemAge,
				prometheus.GaugeValue,
				data.HttpRequestQueuesMaxQueueItemAge,
				data.Name,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.httpRequestQueuesArrivedRequests,
			prometheus.CounterValue,
			data.HttpRequestQueuesArrivalRate,
			data.Name,
		)
	}

	return nil
}
