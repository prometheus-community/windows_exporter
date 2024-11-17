//go:build windows

package exchange

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	externalActiveRemoteDeliveryQueueLength = "External Active Remote Delivery Queue Length"
	internalActiveRemoteDeliveryQueueLength = "Internal Active Remote Delivery Queue Length"
	activeMailboxDeliveryQueueLength        = "Active Mailbox Delivery Queue Length"
	retryMailboxDeliveryQueueLength         = "Retry Mailbox Delivery Queue Length"
	unreachableQueueLength                  = "Unreachable Queue Length"
	externalLargestDeliveryQueueLength      = "External Largest Delivery Queue Length"
	internalLargestDeliveryQueueLength      = "Internal Largest Delivery Queue Length"
	poisonQueueLength                       = "Poison Queue Length"
)

func (c *Collector) buildTransportQueues() error {
	counters := []string{
		externalActiveRemoteDeliveryQueueLength,
		internalActiveRemoteDeliveryQueueLength,
		activeMailboxDeliveryQueueLength,
		retryMailboxDeliveryQueueLength,
		unreachableQueueLength,
		externalLargestDeliveryQueueLength,
		internalLargestDeliveryQueueLength,
		poisonQueueLength,
	}

	var err error

	c.perfDataCollectorTransportQueues, err = perfdata.NewCollector("MSExchangeTransport Queues", perfdata.InstanceAll, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchangeTransport Queues collector: %w", err)
	}

	return nil
}

func (c *Collector) collectTransportQueues(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorTransportQueues.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect MSExchangeTransport Queues: %w", err)
	}

	if len(perfData) == 0 {
		return errors.New("perflib query for MSExchangeTransport Queues returned empty result set")
	}

	for name, data := range perfData {
		labelName := c.toLabelName(name)

		ch <- prometheus.MustNewConstMetric(
			c.externalActiveRemoteDeliveryQueueLength,
			prometheus.GaugeValue,
			data[externalActiveRemoteDeliveryQueueLength].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.internalActiveRemoteDeliveryQueueLength,
			prometheus.GaugeValue,
			data[internalActiveRemoteDeliveryQueueLength].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.activeMailboxDeliveryQueueLength,
			prometheus.GaugeValue,
			data[activeMailboxDeliveryQueueLength].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.retryMailboxDeliveryQueueLength,
			prometheus.GaugeValue,
			data[retryMailboxDeliveryQueueLength].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.unreachableQueueLength,
			prometheus.GaugeValue,
			data[unreachableQueueLength].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.externalLargestDeliveryQueueLength,
			prometheus.GaugeValue,
			data[externalLargestDeliveryQueueLength].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.internalLargestDeliveryQueueLength,
			prometheus.GaugeValue,
			data[internalLargestDeliveryQueueLength].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.poisonQueueLength,
			prometheus.GaugeValue,
			data[poisonQueueLength].FirstValue,
			labelName,
		)
	}

	return nil
}
