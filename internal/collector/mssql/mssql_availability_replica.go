//go:build windows

package mssql

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorAvailabilityReplica struct {
	availabilityReplicaPerfDataCollectors map[string]*perfdata.Collector

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

const (
	availReplicaBytesReceivedFromReplicaPerSec = "Bytes Received from Replica/sec"
	availReplicaBytesSentToReplicaPerSec       = "Bytes Sent to Replica/sec"
	availReplicaBytesSentToTransportPerSec     = "Bytes Sent to Transport/sec"
	availReplicaFlowControlPerSec              = "Flow Control/sec"
	availReplicaFlowControlTimeMSPerSec        = "Flow Control Time (ms/sec)"
	availReplicaReceivesFromReplicaPerSec      = "Receives from Replica/sec"
	availReplicaResentMessagesPerSec           = "Resent Messages/sec"
	availReplicaSendsToReplicaPerSec           = "Sends to Replica/sec"
	availReplicaSendsToTransportPerSec         = "Sends to Transport/sec"
)

func (c *Collector) buildAvailabilityReplica() error {
	var err error

	c.availabilityReplicaPerfDataCollectors = make(map[string]*perfdata.Collector, len(c.mssqlInstances))
	counters := []string{
		availReplicaBytesReceivedFromReplicaPerSec,
		availReplicaBytesSentToReplicaPerSec,
		availReplicaBytesSentToTransportPerSec,
		availReplicaFlowControlPerSec,
		availReplicaFlowControlTimeMSPerSec,
		availReplicaReceivesFromReplicaPerSec,
		availReplicaResentMessagesPerSec,
		availReplicaSendsToReplicaPerSec,
		availReplicaSendsToTransportPerSec,
	}

	for sqlInstance := range c.mssqlInstances {
		c.availabilityReplicaPerfDataCollectors[sqlInstance], err = perfdata.NewCollector(c.mssqlGetPerfObjectName(sqlInstance, "Availability Replica"), perfdata.InstanceAll, counters)
		if err != nil {
			return fmt.Errorf("failed to create Availability Replica collector for instance %s: %w", sqlInstance, err)
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

	return nil
}

func (c *Collector) collectAvailabilityReplica(ch chan<- prometheus.Metric) error {
	return c.collect(ch, subCollectorAvailabilityReplica, c.availabilityReplicaPerfDataCollectors, c.collectAvailabilityReplicaInstance)
}

func (c *Collector) collectAvailabilityReplicaInstance(ch chan<- prometheus.Metric, sqlInstance string, perfDataCollector *perfdata.Collector) error {
	perfData, err := perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect %s metrics: %w", c.mssqlGetPerfObjectName(sqlInstance, "Availability Replica"), err)
	}

	for replicaName, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.availReplicaBytesReceivedFromReplica,
			prometheus.CounterValue,
			data[availReplicaBytesReceivedFromReplicaPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaBytesSentToReplica,
			prometheus.CounterValue,
			data[availReplicaBytesSentToReplicaPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaBytesSentToTransport,
			prometheus.CounterValue,
			data[availReplicaBytesSentToTransportPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaFlowControl,
			prometheus.CounterValue,
			data[availReplicaFlowControlPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaFlowControlTimeMS,
			prometheus.CounterValue,
			utils.MilliSecToSec(data[availReplicaFlowControlTimeMSPerSec].FirstValue),
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaReceivesFromReplica,
			prometheus.CounterValue,
			data[availReplicaReceivesFromReplicaPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaResentMessages,
			prometheus.CounterValue,
			data[availReplicaResentMessagesPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaSendsToReplica,
			prometheus.CounterValue,
			data[availReplicaSendsToReplicaPerSec].FirstValue,
			sqlInstance, replicaName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.availReplicaSendsToTransport,
			prometheus.CounterValue,
			data[availReplicaSendsToTransportPerSec].FirstValue,
			sqlInstance, replicaName,
		)
	}

	return nil
}

func (c *Collector) closeAvailabilityReplica() {
	for _, perfDataCollector := range c.availabilityReplicaPerfDataCollectors {
		perfDataCollector.Close()
	}
}
