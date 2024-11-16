package exchange

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	v1 "github.com/prometheus-community/windows_exporter/internal/perfdata/v1"
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

// Perflib: [36934] MSExchange HttpProxy.
type perflibHTTPProxy struct {
	Name string

	MailboxServerLocatorAverageLatency float64 `perflib:"MailboxServerLocator Average Latency (Moving Average)"`
	AverageAuthenticationLatency       float64 `perflib:"Average Authentication Latency"`
	AverageCASProcessingLatency        float64 `perflib:"Average ClientAccess Server Processing Latency"`
	MailboxServerProxyFailureRate      float64 `perflib:"Mailbox Server Proxy Failure Rate"`
	OutstandingProxyRequests           float64 `perflib:"Outstanding Proxy Requests"`
	ProxyRequestsPerSec                float64 `perflib:"Proxy Requests/Sec"`
}

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

func (c *Collector) collectHTTPProxy(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	var data []perflibHTTPProxy

	if err := v1.UnmarshalObject(ctx.PerfObjects["MSExchange HttpProxy"], &data, logger); err != nil {
		return err
	}

	for _, instance := range data {
		labelName := c.toLabelName(instance.Name)
		ch <- prometheus.MustNewConstMetric(
			c.mailboxServerLocatorAverageLatency,
			prometheus.GaugeValue,
			c.msToSec(instance.MailboxServerLocatorAverageLatency),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.averageAuthenticationLatency,
			prometheus.GaugeValue,
			instance.AverageAuthenticationLatency,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.averageCASProcessingLatency,
			prometheus.GaugeValue,
			c.msToSec(instance.AverageCASProcessingLatency),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.mailboxServerProxyFailureRate,
			prometheus.GaugeValue,
			instance.MailboxServerProxyFailureRate,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.outstandingProxyRequests,
			prometheus.GaugeValue,
			instance.OutstandingProxyRequests,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.proxyRequestsPerSec,
			prometheus.CounterValue,
			instance.ProxyRequestsPerSec,
			labelName,
		)
	}

	return nil
}

func (c *Collector) collectPDHHTTPProxy(ch chan<- prometheus.Metric) error {
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
