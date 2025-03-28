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

package mssql

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorAvailabilityReplica struct {
	availabilityReplicaPerfDataCollectors map[mssqlInstance]*pdh.Collector
	availabilityReplicaPerfDataObject     []perfDataCounterValuesAvailabilityReplica

	availReplicaBytesReceivedFromReplica *prometheus.Desc
	availReplicaBytesSentToReplica       *prometheus.Desc
	availReplicaBytesSentToTransport     *prometheus.Desc
	availReplicaFlowControl              *prometheus.Desc
	availReplicaFlowControlTimeMS        *prometheus.Desc
	availReplicaReceivesFromReplica      *prometheus.Desc
	availReplicaResentMessages           *prometheus.Desc
	availReplicaSendsToReplica           *prometheus.Desc
	availReplicaSendsToTransport         *prometheus.Desc
}

type perfDataCounterValuesAvailabilityReplica struct {
	Name string

	AvailReplicaBytesReceivedFromReplicaPerSec float64 `perfdata:"Bytes Received from Replica/sec"`
	AvailReplicaBytesSentToReplicaPerSec       float64 `perfdata:"Bytes Sent to Replica/sec"`
	AvailReplicaBytesSentToTransportPerSec     float64 `perfdata:"Bytes Sent to Transport/sec"`
	AvailReplicaFlowControlPerSec              float64 `perfdata:"Flow Control/sec"`
	AvailReplicaFlowControlTimeMSPerSec        float64 `perfdata:"Flow Control Time (ms/sec)"`
	AvailReplicaReceivesFromReplicaPerSec      float64 `perfdata:"Receives from Replica/sec"`
	AvailReplicaResentMessagesPerSec           float64 `perfdata:"Resent Messages/sec"`
	AvailReplicaSendsToReplicaPerSec           float64 `perfdata:"Sends to Replica/sec"`
	AvailReplicaSendsToTransportPerSec         float64 `perfdata:"Sends to Transport/sec"`
}

func (c *Collector) buildAvailabilityReplica() error {
	var err error

	c.availabilityReplicaPerfDataCollectors = make(map[mssqlInstance]*pdh.Collector, len(c.mssqlInstances))
	errs := make([]error, 0, len(c.mssqlInstances))

	for _, sqlInstance := range c.mssqlInstances {
		c.availabilityReplicaPerfDataCollectors[sqlInstance], err = pdh.NewCollector[perfDataCounterValuesAvailabilityReplica](pdh.CounterTypeRaw, c.mssqlGetPerfObjectName(sqlInstance, "Availability Replica"), pdh.InstancesAll)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create Availability Replica collector for instance %s: %w", sqlInstance.name, err))
		}
	}

	// Win32_PerfRawData_{instance}_SQLServerAvailabilityReplica
	c.availReplicaBytesReceivedFromReplica = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_received_from_replica_bytes"),
		"(AvailabilityReplica.BytesReceivedfromReplica)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.availReplicaBytesSentToReplica = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_sent_to_replica_bytes"),
		"(AvailabilityReplica.BytesSenttoReplica)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.availReplicaBytesSentToTransport = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_sent_to_transport_bytes"),
		"(AvailabilityReplica.BytesSenttoTransport)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.availReplicaFlowControl = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_initiated_flow_controls"),
		"(AvailabilityReplica.FlowControl)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.availReplicaFlowControlTimeMS = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_flow_control_wait_seconds"),
		"(AvailabilityReplica.FlowControlTimems)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.availReplicaReceivesFromReplica = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_receives_from_replica"),
		"(AvailabilityReplica.ReceivesfromReplica)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.availReplicaResentMessages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_resent_messages"),
		"(AvailabilityReplica.ResentMessages)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.availReplicaSendsToReplica = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_sends_to_replica"),
		"(AvailabilityReplica.SendstoReplica)",
		[]string{"mssql_instance", "replica"},
		nil,
	)
	c.availReplicaSendsToTransport = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "availreplica_sends_to_transport"),
		"(AvailabilityReplica.SendstoTransport)",
		[]string{"mssql_instance", "replica"},
		nil,
	)

	return errors.Join(errs...)
}

func (c *Collector) collectAvailabilityReplica(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorAvailabilityReplica, c.availabilityReplicaPerfDataCollectors, c.collectAvailabilityReplicaInstance)
}

func (c *Collector) collectAvailabilityReplicaInstance(ch chan<- prometheus.Metric, sqlInstance mssqlInstance, perfDataCollector *pdh.Collector) error {
	err := perfDataCollector.Collect(&c.availabilityReplicaPerfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Availability Replica"), err)
	}

	for _, data := range c.availabilityReplicaPerfDataObject {
		ch <- prometheus.MustNewConstMetric(
			c.availReplicaBytesReceivedFromReplica,
			prometheus.CounterValue,
			data.AvailReplicaBytesReceivedFromReplicaPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaBytesSentToReplica,
			prometheus.CounterValue,
			data.AvailReplicaBytesSentToReplicaPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaBytesSentToTransport,
			prometheus.CounterValue,
			data.AvailReplicaBytesSentToTransportPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaFlowControl,
			prometheus.CounterValue,
			data.AvailReplicaFlowControlPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaFlowControlTimeMS,
			prometheus.CounterValue,
			utils.MilliSecToSec(data.AvailReplicaFlowControlTimeMSPerSec),
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaReceivesFromReplica,
			prometheus.CounterValue,
			data.AvailReplicaReceivesFromReplicaPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaResentMessages,
			prometheus.CounterValue,
			data.AvailReplicaResentMessagesPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaSendsToReplica,
			prometheus.CounterValue,
			data.AvailReplicaSendsToReplicaPerSec,
			sqlInstance.name, data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaSendsToTransport,
			prometheus.CounterValue,
			data.AvailReplicaSendsToTransportPerSec,
			sqlInstance.name, data.Name,
		)
	}

	return nil
}

func (c *Collector) closeAvailabilityReplica() {
	for _, perfDataCollector := range c.availabilityReplicaPerfDataCollectors {
		perfDataCollector.Close()
	}
}
