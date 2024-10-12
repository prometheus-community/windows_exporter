package exchange

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	v1 "github.com/prometheus-community/windows_exporter/internal/perfdata/v1"
	"github.com/prometheus-community/windows_exporter/internal/types"
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

// Perflib: [20524] MSExchangeTransport Queues.
type perflibTransportQueues struct {
	Name string

	ExternalActiveRemoteDeliveryQueueLength float64 `perflib:"External Active Remote Delivery Queue Length"`
	InternalActiveRemoteDeliveryQueueLength float64 `perflib:"Internal Active Remote Delivery Queue Length"`
	ActiveMailboxDeliveryQueueLength        float64 `perflib:"Active Mailbox Delivery Queue Length"`
	RetryMailboxDeliveryQueueLength         float64 `perflib:"Retry Mailbox Delivery Queue Length"`
	UnreachableQueueLength                  float64 `perflib:"Unreachable Queue Length"`
	ExternalLargestDeliveryQueueLength      float64 `perflib:"External Largest Delivery Queue Length"`
	InternalLargestDeliveryQueueLength      float64 `perflib:"Internal Largest Delivery Queue Length"`
	PoisonQueueLength                       float64 `perflib:"Poison Queue Length"`
}

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

	c.perfDataCollectorTransportQueues, err = perfdata.NewCollector(perfdata.V1, "MSExchangeTransport Queues", perfdata.AllInstances, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchangeTransport Queues collector: %w", err)
	}

	return nil
}

func (c *Collector) collectTransportQueues(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	var data []perflibTransportQueues

	if err := v1.UnmarshalObject(ctx.PerfObjects["MSExchangeTransport Queues"], &data, logger); err != nil {
		return err
	}

	for _, queue := range data {
		labelName := c.toLabelName(queue.Name)
		if strings.HasSuffix(labelName, "_total") {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			c.externalActiveRemoteDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.ExternalActiveRemoteDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.internalActiveRemoteDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.InternalActiveRemoteDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.activeMailboxDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.ActiveMailboxDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.retryMailboxDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.RetryMailboxDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.unreachableQueueLength,
			prometheus.GaugeValue,
			queue.UnreachableQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.externalLargestDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.ExternalLargestDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.internalLargestDeliveryQueueLength,
			prometheus.GaugeValue,
			queue.InternalLargestDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.poisonQueueLength,
			prometheus.GaugeValue,
			queue.PoisonQueueLength,
			labelName,
		)
	}

	return nil
}

func (c *Collector) collectPDHTransportQueues(ch chan<- prometheus.Metric) error {
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
