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
	externalActiveRemoteDeliveryQueueLength = "External Active Remote Delivery Queue Length"
	internalActiveRemoteDeliveryQueueLength = "Internal Active Remote Delivery Queue Length"
	activeMailboxDeliveryQueueLength        = "Active Mailbox Delivery Queue Length"
	retryMailboxDeliveryQueueLength         = "Retry Mailbox Delivery Queue Length"
	unreachableQueueLength                  = "Unreachable Queue Length"
	externalLargestDeliveryQueueLength      = "External Largest Delivery Queue Length"
	internalLargestDeliveryQueueLength      = "Internal Largest Delivery Queue Length"
	poisonQueueLength                       = "Poison Queue Length"
	messagesQueuedForDeliveryTotal          = "Messages Queued For Delivery Total"
	messagesSubmittedTotal                  = "Messages Submitted Total"
	messagesDelayedTotal                    = "Messages Delayed Total"
	messagesCompletedDeliveryTotal          = "Messages Completed Delivery Total"
	shadowQueueLength                       = "Shadow Queue Length"
	submissionQueueLength                   = "Submission Queue Length"
	delayQueueLength                        = "Delay Queue Length"
	itemsCompletedDeliveryTotal             = "Items Completed Delivery Total"
	itemsQueuedForDeliveryExpiredTotal      = "Items Queued For Delivery Expired Total"
	itemsQueuedForDeliveryTotal             = "Items Queued For Delivery Total"
	itemsResubmittedTotal                   = "Items Resubmitted Total"
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
		messagesQueuedForDeliveryTotal,
		messagesSubmittedTotal,
		messagesDelayedTotal,
		messagesCompletedDeliveryTotal,
		shadowQueueLength,
		submissionQueueLength,
		delayQueueLength,
		itemsCompletedDeliveryTotal,
		itemsQueuedForDeliveryExpiredTotal,
		itemsQueuedForDeliveryTotal,
		itemsResubmittedTotal,
	}

	var err error

	c.perfDataCollectorTransportQueues, err = perfdata.NewCollector("MSExchangeTransport Queues", perfdata.InstanceAll, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchangeTransport Queues collector: %w", err)
	}

	c.externalActiveRemoteDeliveryQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_external_active_remote_delivery"),
		"External Active Remote Delivery Queue length",
		[]string{"name"},
		nil,
	)
	c.internalActiveRemoteDeliveryQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_internal_active_remote_delivery"),
		"Internal Active Remote Delivery Queue length",
		[]string{"name"},
		nil,
	)
	c.activeMailboxDeliveryQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_active_mailbox_delivery"),
		"Active Mailbox Delivery Queue length",
		[]string{"name"},
		nil,
	)
	c.retryMailboxDeliveryQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_retry_mailbox_delivery"),
		"Retry Mailbox Delivery Queue length",
		[]string{"name"},
		nil,
	)
	c.unreachableQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_unreachable"),
		"Unreachable Queue length",
		[]string{"name"},
		nil,
	)
	c.externalLargestDeliveryQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_external_largest_delivery"),
		"External Largest Delivery Queue length",
		[]string{"name"},
		nil,
	)
	c.internalLargestDeliveryQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_internal_largest_delivery"),
		"Internal Largest Delivery Queue length",
		[]string{"name"},
		nil,
	)
	c.poisonQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_poison"),
		"Poison Queue length",
		[]string{"name"},
		nil,
	)
	c.messagesQueuedForDeliveryTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_messages_queued_for_delivery_total"),
		"Messages Queued For Delivery Total",
		[]string{"name"},
		nil,
	)
	c.messagesSubmittedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_messages_submitted_total"),
		"Messages Submitted Total",
		[]string{"name"},
		nil,
	)
	c.messagesDelayedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_messages_delayed_total"),
		"Messages Delayed Total",
		[]string{"name"},
		nil,
	)
	c.messagesCompletedDeliveryTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_messages_completed_delivery_total"),
		"Messages Completed Delivery Total",
		[]string{"name"},
		nil,
	)
	c.shadowQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_shadow_queue_length"),
		"Shadow Queue Length",
		[]string{"name"},
		nil,
	)
	c.submissionQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_submission_queue_length"),
		"Submission Queue Length",
		[]string{"name"},
		nil,
	)
	c.delayQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_delay_queue_length"),
		"Delay Queue Length",
		[]string{"name"},
		nil,
	)
	c.itemsCompletedDeliveryTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_items_completed_delivery_total"),
		"Items Completed Delivery Total",
		[]string{"name"},
		nil,
	)
	c.itemsQueuedForDeliveryExpiredTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_items_queued_for_delivery_expired_total"),
		"Items Queued For Delivery Expired Total",
		[]string{"name"},
		nil,
	)
	c.itemsQueuedForDeliveryTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_items_queued_for_delivery_total"),
		"Items Queued For Delivery Total",
		[]string{"name"},
		nil,
	)
	c.itemsResubmittedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_items_resubmitted_total"),
		"Items Resubmitted Total",
		[]string{"name"},
		nil,
	)

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
		ch <- prometheus.MustNewConstMetric(
			c.messagesQueuedForDeliveryTotal,
			prometheus.CounterValue,
			data[messagesQueuedForDeliveryTotal].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.messagesSubmittedTotal,
			prometheus.CounterValue,
			data[messagesSubmittedTotal].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.messagesDelayedTotal,
			prometheus.CounterValue,
			data[messagesDelayedTotal].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.messagesCompletedDeliveryTotal,
			prometheus.CounterValue,
			data[messagesCompletedDeliveryTotal].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.shadowQueueLength,
			prometheus.GaugeValue,
			data[shadowQueueLength].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.submissionQueueLength,
			prometheus.GaugeValue,
			data[submissionQueueLength].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.delayQueueLength,
			prometheus.GaugeValue,
			data[delayQueueLength].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.itemsCompletedDeliveryTotal,
			prometheus.CounterValue,
			data[itemsCompletedDeliveryTotal].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.itemsQueuedForDeliveryExpiredTotal,
			prometheus.CounterValue,
			data[itemsQueuedForDeliveryExpiredTotal].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.itemsQueuedForDeliveryTotal,
			prometheus.CounterValue,
			data[itemsQueuedForDeliveryTotal].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.itemsResubmittedTotal,
			prometheus.CounterValue,
			data[itemsResubmittedTotal].FirstValue,
			labelName,
		)
	}

	return nil
}
