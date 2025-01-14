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
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorHTTPProxy struct {
	perfDataCollectorHTTPProxy *pdh.Collector
	perfDataObjectHTTPProxy    []perfDataCounterValuesHTTPProxy

	mailboxServerLocatorAverageLatency *prometheus.Desc
	averageAuthenticationLatency       *prometheus.Desc
	outstandingProxyRequests           *prometheus.Desc
	proxyRequestsPerSec                *prometheus.Desc
	averageCASProcessingLatency        *prometheus.Desc
	mailboxServerProxyFailureRate      *prometheus.Desc
}

type perfDataCounterValuesHTTPProxy struct {
	Name string

	MailboxServerLocatorAverageLatency float64 `perfdata:"MailboxServerLocator Average Latency (Moving Average)"`
	AverageAuthenticationLatency       float64 `perfdata:"Average Authentication Latency"`
	AverageCASProcessingLatency        float64 `perfdata:"Average ClientAccess Server Processing Latency"`
	MailboxServerProxyFailureRate      float64 `perfdata:"Mailbox Server Proxy Failure Rate"`
	OutstandingProxyRequests           float64 `perfdata:"Outstanding Proxy Requests"`
	ProxyRequestsPerSec                float64 `perfdata:"Proxy Requests/Sec"`
}

func (c *Collector) buildHTTPProxy() error {
	var err error

	c.perfDataCollectorHTTPProxy, err = pdh.NewCollector[perfDataCounterValuesHTTPProxy](pdh.CounterTypeRaw, "MSExchange HttpProxy", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange HttpProxy collector: %w", err)
	}

	c.mailboxServerLocatorAverageLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "http_proxy_mailbox_server_locator_avg_latency_sec"),
		"Average latency (sec) of MailboxServerLocator web service calls",
		[]string{"name"},
		nil,
	)
	c.averageAuthenticationLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "http_proxy_avg_auth_latency"),
		"Average time spent authenticating CAS requests over the last 200 samples",
		[]string{"name"},
		nil,
	)
	c.outstandingProxyRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "http_proxy_outstanding_proxy_requests"),
		"Number of concurrent outstanding proxy requests",
		[]string{"name"},
		nil,
	)
	c.proxyRequestsPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "http_proxy_requests_total"),
		"Number of proxy requests processed each second",
		[]string{"name"},
		nil,
	)
	c.averageCASProcessingLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "http_proxy_avg_cas_processing_latency_sec"),
		"Average latency (sec) of CAS processing time over the last 200 reqs",
		[]string{"name"},
		nil,
	)
	c.mailboxServerProxyFailureRate = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "http_proxy_mailbox_proxy_failure_rate"),
		"% of failures between this CAS and MBX servers over the last 200 samples",
		[]string{"name"},
		nil,
	)

	return nil
}

func (c *Collector) collectHTTPProxy(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorHTTPProxy.Collect(&c.perfDataObjectHTTPProxy)
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange HttpProxy Service metrics: %w", err)
	}

	for _, data := range c.perfDataObjectHTTPProxy {
		labelName := c.toLabelName(data.Name)
		ch <- prometheus.MustNewConstMetric(
			c.mailboxServerLocatorAverageLatency,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data.MailboxServerLocatorAverageLatency),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.averageAuthenticationLatency,
			prometheus.GaugeValue,
			data.AverageAuthenticationLatency,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.averageCASProcessingLatency,
			prometheus.GaugeValue,
			utils.MilliSecToSec(data.AverageCASProcessingLatency),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.mailboxServerProxyFailureRate,
			prometheus.GaugeValue,
			data.MailboxServerProxyFailureRate,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outstandingProxyRequests,
			prometheus.GaugeValue,
			data.OutstandingProxyRequests,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.proxyRequestsPerSec,
			prometheus.CounterValue,
			data.ProxyRequestsPerSec,
			labelName,
		)
	}

	return nil
}
