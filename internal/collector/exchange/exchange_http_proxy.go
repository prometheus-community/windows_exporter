//go:build windows

package exchange

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
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
