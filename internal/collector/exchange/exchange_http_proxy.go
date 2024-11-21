//go:build windows

package exchange

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	mailboxServerLocatorAverageLatency = "MailboxServerLocator Average Latency (Moving Average)"
	averageAuthenticationLatency       = "Average Authentication Latency"
	averageCASProcessingLatency        = "Average ClientAccess Server Processing Latency"
	mailboxServerProxyFailureRate      = "Mailbox Server Proxy Failure Rate"
	outstandingProxyRequests           = "Outstanding Proxy Requests"
	proxyRequestsPerSec                = "Proxy Requests/Sec"
)

func (c *Collector) buildHTTPProxy() error {
	counters := []string{
		mailboxServerLocatorAverageLatency,
		averageAuthenticationLatency,
		averageCASProcessingLatency,
		mailboxServerProxyFailureRate,
		outstandingProxyRequests,
		proxyRequestsPerSec,
	}

	var err error

	c.perfDataCollectorHttpProxy, err = perfdata.NewCollector("MSExchange HttpProxy", perfdata.InstanceAll, counters)
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
	perfData, err := c.perfDataCollectorHttpProxy.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange HttpProxy Service metrics: %w", err)
	}

	if len(perfData) == 0 {
		return errors.New("perflib query for MSExchange HttpProxy Service returned empty result set")
	}

	for name, data := range perfData {
		labelName := c.toLabelName(name)
		ch <- prometheus.MustNewConstMetric(
			c.mailboxServerLocatorAverageLatency,
			prometheus.GaugeValue,
			c.msToSec(data[mailboxServerLocatorAverageLatency].FirstValue),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.averageAuthenticationLatency,
			prometheus.GaugeValue,
			data[averageAuthenticationLatency].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.averageCASProcessingLatency,
			prometheus.GaugeValue,
			c.msToSec(data[averageCASProcessingLatency].FirstValue),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.mailboxServerProxyFailureRate,
			prometheus.GaugeValue,
			data[mailboxServerProxyFailureRate].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outstandingProxyRequests,
			prometheus.GaugeValue,
			data[outstandingProxyRequests].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.proxyRequestsPerSec,
			prometheus.CounterValue,
			data[proxyRequestsPerSec].FirstValue,
			labelName,
		)
	}

	return nil
}
