// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
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

package hyperv

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorReplica Hyper-V Replica VM metrics.
type collectorReplica struct {
	perfDataCollectorReplica *pdh.Collector
	perfDataObjectReplica    []perfDataCounterValuesReplica

	miSession     *mi.Session
	miQueryReplica mi.Query

	// \Hyper-V Replica VM(*)\Compression Efficiency
	replicaCompressionEfficiency *prometheus.Desc
	// \Hyper-V Replica VM(*)\Resynchronized Bytes
	replicaResynchronizedBytesTotal *prometheus.Desc
	// \Hyper-V Replica VM(*)\Network Bytes Recv
	replicaNetworkBytesRecvTotal *prometheus.Desc
	// \Hyper-V Replica VM(*)\Network Bytes Sent
	replicaNetworkBytesSentTotal *prometheus.Desc
	// \Hyper-V Replica VM(*)\Average Replication Latency
	replicaAverageReplicationLatencySeconds *prometheus.Desc
	// \Hyper-V Replica VM(*)\Replication Latency
	replicaReplicationLatencySeconds *prometheus.Desc
	// \Hyper-V Replica VM(*)\Replication Count
	replicaReplicationCountTotal *prometheus.Desc
	// \Hyper-V Replica VM(*)\Average Replication Size
	replicaAverageReplicationSizeBytes *prometheus.Desc
	// \Hyper-V Replica VM(*)\Last Replication Size
	replicaLastReplicationSizeBytes *prometheus.Desc

	// Msvm_ReplicationRelationship\ReplicationHealth
	replicaHealth *prometheus.Desc
	// Msvm_ReplicationRelationship\ReplicationState
	replicaState *prometheus.Desc
}

type perfDataCounterValuesReplica struct {
	Name string

	ReplicaCompressionEfficiency          float64 `perfdata:"Compression Efficiency"`
	ReplicaResynchronizedBytes            float64 `perfdata:"Resynchronized Bytes"`
	ReplicaNetworkBytesRecv               float64 `perfdata:"Network Bytes Recv"`
	ReplicaNetworkBytesSent               float64 `perfdata:"Network Bytes Sent"`
	ReplicaAverageReplicationLatency      float64 `perfdata:"Average Replication Latency"`
	ReplicaReplicationLatency             float64 `perfdata:"Replication Latency"`
	ReplicaReplicationCount               float64 `perfdata:"Replication Count"`
	ReplicaAverageReplicationSize         float64 `perfdata:"Average Replication Size"`
	ReplicaLastReplicationSize            float64 `perfdata:"Last Replication Size"`
}

type wmiMsvmReplicationRelationship struct {
	Name              string `mi:"Name"`
	ReplicationHealth uint16 `mi:"ReplicationHealth"`
	ReplicationState  uint16 `mi:"ReplicationState"`
}

func (c *Collector) buildReplica(miSession *mi.Session) error {
	var err error

	c.perfDataCollectorReplica, err = pdh.NewCollector[perfDataCounterValuesReplica](c.logger, pdh.CounterTypeRaw, "Hyper-V Replica VM", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Replica VM collector: %w", err)
	}

	c.replicaCompressionEfficiency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_compression_efficiency"),
		"Compression efficiency of the replication channel in percentage",
		[]string{"vm"},
		nil,
	)
	c.replicaResynchronizedBytesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_resynchronized_bytes_total"),
		"Total number of bytes resynchronized during replication",
		[]string{"vm"},
		nil,
	)
	c.replicaNetworkBytesRecvTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_network_bytes_recv_total"),
		"Total number of bytes received over the network for replication",
		[]string{"vm"},
		nil,
	)
	c.replicaNetworkBytesSentTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_network_bytes_sent_total"),
		"Total number of bytes sent over the network for replication",
		[]string{"vm"},
		nil,
	)
	c.replicaAverageReplicationLatencySeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_average_replication_latency_seconds"),
		"Average replication latency in seconds",
		[]string{"vm"},
		nil,
	)
	c.replicaReplicationLatencySeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_replication_latency_seconds"),
		"Current replication latency in seconds",
		[]string{"vm"},
		nil,
	)
	c.replicaReplicationCountTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_replication_count_total"),
		"Total number of replications performed",
		[]string{"vm"},
		nil,
	)
	c.replicaAverageReplicationSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_average_replication_size_bytes"),
		"Average size of a replication in bytes",
		[]string{"vm"},
		nil,
	)
	c.replicaLastReplicationSizeBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_last_replication_size_bytes"),
		"Size of the last replication in bytes",
		[]string{"vm"},
		nil,
	)

	c.replicaHealth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_health"),
		"Replication health of the virtual machine. 0 = NotApplicable, 1 = Ok, 2 = Warning, 3 = Critical",
		[]string{"vm"},
		nil,
	)
	c.replicaState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_state"),
		"Replication state of the virtual machine. 0 = Disabled, 1 = Ready, 2 = WaitingToCompleteInitialReplication, 3 = Replicating, 4 = SyncedReplicationComplete, 5 = Recovered, 6 = Committed, 7 = Suspended, 8 = Critical, 9 = WaitingForStartResync, 10 = ResyncInProgress, 11 = Failover, 12 = TestFailoverStarted",
		[]string{"vm"},
		nil,
	)

	if miSession == nil {
		return errors.New("miSession is nil")
	}

	miQuery, err := mi.NewQuery("SELECT Name, ReplicationHealth, ReplicationState FROM Msvm_ReplicationRelationship")
	if err != nil {
		return fmt.Errorf("failed to create WMI query for Msvm_ReplicationRelationship: %w", err)
	}

	c.miQueryReplica = miQuery
	c.miSession = miSession

	return nil
}

func (c *Collector) collectReplica(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0, 2)

	if err := c.collectReplicaPerfData(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect Hyper-V Replica VM performance metrics: %w", err))
	}

	if err := c.collectReplicaWMI(ch); err != nil {
		errs = append(errs, fmt.Errorf("failed to collect Hyper-V Replica WMI metrics: %w", err))
	}

	return errors.Join(errs...)
}

func (c *Collector) collectReplicaPerfData(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorReplica.Collect(&c.perfDataObjectReplica)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Replica VM metrics: %w", err)
	}

	for _, data := range c.perfDataObjectReplica {
		ch <- prometheus.MustNewConstMetric(
			c.replicaCompressionEfficiency,
			prometheus.GaugeValue,
			data.ReplicaCompressionEfficiency,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.replicaResynchronizedBytesTotal,
			prometheus.CounterValue,
			data.ReplicaResynchronizedBytes,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.replicaNetworkBytesRecvTotal,
			prometheus.CounterValue,
			data.ReplicaNetworkBytesRecv,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.replicaNetworkBytesSentTotal,
			prometheus.CounterValue,
			data.ReplicaNetworkBytesSent,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.replicaAverageReplicationLatencySeconds,
			prometheus.GaugeValue,
			data.ReplicaAverageReplicationLatency,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.replicaReplicationLatencySeconds,
			prometheus.GaugeValue,
			data.ReplicaReplicationLatency,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.replicaReplicationCountTotal,
			prometheus.CounterValue,
			data.ReplicaReplicationCount,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.replicaAverageReplicationSizeBytes,
			prometheus.GaugeValue,
			data.ReplicaAverageReplicationSize,
			data.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.replicaLastReplicationSizeBytes,
			prometheus.GaugeValue,
			data.ReplicaLastReplicationSize,
			data.Name,
		)
	}

	return nil
}

func (c *Collector) collectReplicaWMI(ch chan<- prometheus.Metric) error {
	var relationships []wmiMsvmReplicationRelationship

	if err := c.miSession.Query(&relationships, mi.NamespaceRootVirtualizationV2, c.miQueryReplica, 0); err != nil {
		return fmt.Errorf("WMI query for Msvm_ReplicationRelationship failed: %w", err)
	}

	for _, rel := range relationships {
		ch <- prometheus.MustNewConstMetric(
			c.replicaHealth,
			prometheus.GaugeValue,
			float64(rel.ReplicationHealth),
			rel.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			c.replicaState,
			prometheus.GaugeValue,
			float64(rel.ReplicationState),
			rel.Name,
		)
	}

	return nil
}
