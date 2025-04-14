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

type collectorTransportQueues struct {
	perfDataCollectorTransportQueues *pdh.Collector
	perfDataObjectTransportQueues    []perfDataCounterValuesTransportQueues

	activeMailboxDeliveryQueueLength        *prometheus.Desc
	externalActiveRemoteDeliveryQueueLength *prometheus.Desc
	externalLargestDeliveryQueueLength      *prometheus.Desc
	internalActiveRemoteDeliveryQueueLength *prometheus.Desc
	internalLargestDeliveryQueueLength      *prometheus.Desc
	poisonQueueLength                       *prometheus.Desc
	retryMailboxDeliveryQueueLength         *prometheus.Desc
	unreachableQueueLength                  *prometheus.Desc
	messagesQueuedForDeliveryTotal          *prometheus.Desc
	messagesSubmittedTotal                  *prometheus.Desc
	messagesDelayedTotal                    *prometheus.Desc
	messagesCompletedDeliveryTotal          *prometheus.Desc
	aggregateShadowQueueLength              *prometheus.Desc
	submissionQueueLength                   *prometheus.Desc
	delayQueueLength                        *prometheus.Desc
	itemsCompletedDeliveryTotal             *prometheus.Desc
	itemsQueuedForDeliveryExpiredTotal      *prometheus.Desc
	itemsQueuedForDeliveryTotal             *prometheus.Desc
	itemsResubmittedTotal                   *prometheus.Desc
}

type perfDataCounterValuesTransportQueues struct {
	Name string

	ExternalActiveRemoteDeliveryQueueLength float64 `perfdata:"External Active Remote Delivery Queue Length"`
	InternalActiveRemoteDeliveryQueueLength float64 `perfdata:"Internal Active Remote Delivery Queue Length"`
	ActiveMailboxDeliveryQueueLength        float64 `perfdata:"Active Mailbox Delivery Queue Length"`
	RetryMailboxDeliveryQueueLength         float64 `perfdata:"Retry Mailbox Delivery Queue Length"`
	UnreachableQueueLength                  float64 `perfdata:"Unreachable Queue Length"`
	ExternalLargestDeliveryQueueLength      float64 `perfdata:"External Largest Delivery Queue Length"`
	InternalLargestDeliveryQueueLength      float64 `perfdata:"Internal Largest Delivery Queue Length"`
	PoisonQueueLength                       float64 `perfdata:"Poison Queue Length"`
	MessagesQueuedForDeliveryTotal          float64 `perfdata:"Messages Queued For Delivery Total"`
	MessagesSubmittedTotal                  float64 `perfdata:"Messages Submitted Total"`
	MessagesDelayedTotal                    float64 `perfdata:"Messages Delayed Total"`
	MessagesCompletedDeliveryTotal          float64 `perfdata:"Messages Completed Delivery Total"`
	AggregateShadowQueueLength              float64 `perfdata:"Aggregate Shadow Queue Length"`
	SubmissionQueueLength                   float64 `perfdata:"Submission Queue Length"`
	DelayQueueLength                        float64 `perfdata:"Delay Queue Length"`
	ItemsCompletedDeliveryTotal             float64 `perfdata:"Items Completed Delivery Total"`
	ItemsQueuedForDeliveryExpiredTotal      float64 `perfdata:"Items Queued For Delivery Expired Total"`
	ItemsQueuedForDeliveryTotal             float64 `perfdata:"Items Queued For Delivery Total"`
	ItemsResubmittedTotal                   float64 `perfdata:"Items Resubmitted Total"`
}

func (c *Collector) buildTransportQueues() error {
	var err error

	c.perfDataCollectorTransportQueues, err = pdh.NewCollector[perfDataCounterValuesTransportQueues](pdh.CounterTypeRaw, "MSExchangeTransport Queues", pdh.InstancesAll)
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
	c.aggregateShadowQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "transport_queues_aggregate_shadow_queue_length"),
		"The current number of messages in shadow queues.",
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
	err := c.perfDataCollectorTransportQueues.Collect(&c.perfDataObjectTransportQueues)
	if err != nil {
		return fmt.Errorf("failed to collect MSExchangeTransport Queues: %w", err)
	}

	for _, data := range c.perfDataObjectTransportQueues {
		labelName := c.toLabelName(data.Name)

		ch <- prometheus.MustNewConstMetric(
			c.externalActiveRemoteDeliveryQueueLength,
			prometheus.GaugeValue,
			data.ExternalActiveRemoteDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.internalActiveRemoteDeliveryQueueLength,
			prometheus.GaugeValue,
			data.InternalActiveRemoteDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.activeMailboxDeliveryQueueLength,
			prometheus.GaugeValue,
			data.ActiveMailboxDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.retryMailboxDeliveryQueueLength,
			prometheus.GaugeValue,
			data.RetryMailboxDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.unreachableQueueLength,
			prometheus.GaugeValue,
			data.UnreachableQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.externalLargestDeliveryQueueLength,
			prometheus.GaugeValue,
			data.ExternalLargestDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.internalLargestDeliveryQueueLength,
			prometheus.GaugeValue,
			data.InternalLargestDeliveryQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.poisonQueueLength,
			prometheus.GaugeValue,
			data.PoisonQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.messagesQueuedForDeliveryTotal,
			prometheus.CounterValue,
			data.MessagesQueuedForDeliveryTotal,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.messagesSubmittedTotal,
			prometheus.CounterValue,
			data.MessagesSubmittedTotal,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.messagesDelayedTotal,
			prometheus.CounterValue,
			data.MessagesDelayedTotal,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.messagesCompletedDeliveryTotal,
			prometheus.CounterValue,
			data.MessagesCompletedDeliveryTotal,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.aggregateShadowQueueLength,
			prometheus.GaugeValue,
			data.AggregateShadowQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.submissionQueueLength,
			prometheus.GaugeValue,
			data.SubmissionQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.delayQueueLength,
			prometheus.GaugeValue,
			data.DelayQueueLength,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.itemsCompletedDeliveryTotal,
			prometheus.CounterValue,
			data.ItemsCompletedDeliveryTotal,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.itemsQueuedForDeliveryExpiredTotal,
			prometheus.CounterValue,
			data.ItemsQueuedForDeliveryExpiredTotal,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.itemsQueuedForDeliveryTotal,
			prometheus.CounterValue,
			data.ItemsQueuedForDeliveryTotal,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.itemsResubmittedTotal,
			prometheus.CounterValue,
			data.ItemsResubmittedTotal,
			labelName,
		)
	}

	return nil
}
