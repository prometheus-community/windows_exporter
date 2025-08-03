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
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorHttpServiceRequestQueues struct {
	perfDataCollectorHttpServiceRequestQueues *pdh.Collector
	perfDataObjectHttpServiceRequestQueues    []perfDataCounterValuesHttpServiceRequestQueues

	httpRequestQueuesCurrentQueueSize     *prometheus.Desc
	httpRequestQueuesTotalRejectedRequest *prometheus.Desc
	httpRequestQueuesMaxQueueItemAge      *prometheus.Desc
	httpRequestQueuesArrivalRate          *prometheus.Desc
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

	c.perfDataCollectorHttpServiceRequestQueues, err = pdh.NewCollector[perfDataCounterValuesHttpServiceRequestQueues](pdh.CounterTypeRaw, "HTTP Service Request Queues", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Http Service collector: %w", err)
	}

	c.httpRequestQueuesCurrentQueueSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "http_requests_current_queue_size"),
		"Http Request Current Queue Size",
		[]string{"site"},
		nil,
	)
	c.httpRequestQueuesTotalRejectedRequest = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "http_request_total_rejected_request"),
		"Http Request Total Rejected Request",
		[]string{"site"},
		nil,
	)
	c.httpRequestQueuesMaxQueueItemAge = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "http_requests_max_queue_item_age"),
		"Http Request Max Queue Item Age. The values might be bogus if the queue is empty.",
		[]string{"site"},
		nil,
	)
	c.httpRequestQueuesArrivalRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "http_requests_arrival_rate"),
		"Http Request Arrival Rate",
		[]string{"site"},
		nil,
	)

	return nil
}

func (c *Collector) collectHttpServiceRequestQueues(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorHttpServiceRequestQueues.Collect(&c.perfDataObjectHttpServiceRequestQueues)
	if err != nil {
		return fmt.Errorf("failed to collect Http Service Request Queues metrics: %w", err)
	}

	deduplicateIISNames(c.perfDataObjectHttpServiceRequestQueues)

	for _, data := range c.perfDataObjectHttpServiceRequestQueues {
		if strings.HasPrefix(data.Name, "---") {
			continue
		}

		if c.config.SiteExclude.MatchString(data.Name) || !c.config.SiteInclude.MatchString(data.Name) {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.httpRequestQueuesCurrentQueueSize,
			prometheus.GaugeValue,
			data.HttpRequestQueuesCurrentQueueSize,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.httpRequestQueuesTotalRejectedRequest,
			prometheus.GaugeValue,
			data.HttpRequestQueuesTotalRejectedRequests,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.httpRequestQueuesMaxQueueItemAge,
			prometheus.GaugeValue,
			data.HttpRequestQueuesMaxQueueItemAge,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.httpRequestQueuesArrivalRate,
			prometheus.GaugeValue,
			data.HttpRequestQueuesArrivalRate,
			data.Name,
		)
	}

	return nil
}
