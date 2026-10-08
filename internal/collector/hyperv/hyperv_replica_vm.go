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
	"log/slog"
	"strings"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// replicaVMQueryTimeout bounds the replication relationship query, so a hung
// WMI service does not block the scrape.
const replicaVMQueryTimeout = 5 * time.Second

// collectorReplicaVM Hyper-V Replica VM metrics
type collectorReplicaVM struct {
	perfDataCollectorReplicaVM *pdh.Collector[perfDataCounterValuesReplicaVM]
	perfDataObjectReplicaVM    []perfDataCounterValuesReplicaVM

	miQueryReplicaVM mi.Query

	replicaVMAverageReplicationLatency *prometheus.Desc // \Hyper-V Replica VM(*)\Average Replication Latency
	replicaVMAverageReplicationSize    *prometheus.Desc // \Hyper-V Replica VM(*)\Average Replication Size
	replicaVMCompressionEfficiency     *prometheus.Desc // \Hyper-V Replica VM(*)\Compression Efficiency
	replicaVMLastReplicationSize       *prometheus.Desc // \Hyper-V Replica VM(*)\Last Replication Size
	replicaVMNetworkReceivedBytes      *prometheus.Desc // \Hyper-V Replica VM(*)\Network Bytes Recv
	replicaVMNetworkSentBytes          *prometheus.Desc // \Hyper-V Replica VM(*)\Network Bytes Sent
	replicaVMReplicationCount          *prometheus.Desc // \Hyper-V Replica VM(*)\Replication Count
	replicaVMReplicationLatency        *prometheus.Desc // \Hyper-V Replica VM(*)\Replication Latency
	replicaVMResynchronizedBytes       *prometheus.Desc // \Hyper-V Replica VM(*)\Resynchronized Bytes

	replicaVMHealth *prometheus.Desc // Msvm_ReplicationRelationship.ReplicationHealth
	replicaVMState  *prometheus.Desc // Msvm_ReplicationRelationship.ReplicationState
}

type perfDataCounterValuesReplicaVM struct {
	Name string

	AverageReplicationLatency float64 `perfdata:"Average Replication Latency"`
	AverageReplicationSize    float64 `perfdata:"Average Replication Size"`
	CompressionEfficiency     float64 `perfdata:"Compression Efficiency"`
	LastReplicationSize       float64 `perfdata:"Last Replication Size"`
	NetworkBytesRecv          float64 `perfdata:"Network Bytes Recv"`
	NetworkBytesSent          float64 `perfdata:"Network Bytes Sent"`
	ReplicationCount          float64 `perfdata:"Replication Count"`
	ReplicationLatency        float64 `perfdata:"Replication Latency"`
	ResynchronizedBytes       float64 `perfdata:"Resynchronized Bytes"`
}

// msvmReplicationRelationship is the Msvm_ReplicationRelationship WMI class.
// https://learn.microsoft.com/en-us/windows/win32/hyperv_v2/msvm-replicationrelationship
type msvmReplicationRelationship struct {
	ElementName       string `mi:"ElementName"`
	InstanceID        string `mi:"InstanceID"`
	ReplicationHealth uint16 `mi:"ReplicationHealth"`
	ReplicationState  uint16 `mi:"ReplicationState"`
}

func (c *Collector) buildReplicaVM() error {
	if c.miSession == nil {
		return mi.ErrNotInitialized
	}

	var err error

	// The Hyper-V Replica VM counter set is missing on hosts without Hyper-V Replica
	// support, e.g. Windows client editions. Depending on the system, this surfaces
	// here as a missing object or on collect as no data. The WMI metrics still work there.
	c.perfDataCollectorReplicaVM, err = pdh.NewCollector[perfDataCounterValuesReplicaVM](c.logger, pdh.CounterTypeRaw, "Hyper-V Replica VM", pdh.InstancesAll)
	if errors.Is(err, pdh.NewPdhError(pdh.CstatusNoObject)) {
		c.logger.Debug("Hyper-V Replica VM performance counters are not available", slog.Any("err", err))

		c.perfDataCollectorReplicaVM = nil
	} else if err != nil {
		return fmt.Errorf("failed to create Hyper-V Replica VM collector: %w", err)
	}

	c.miQueryReplicaVM, err = mi.NewQuery("SELECT ElementName, InstanceID, ReplicationHealth, ReplicationState FROM Msvm_ReplicationRelationship")
	if err != nil {
		return fmt.Errorf("failed to create WMI query: %w", err)
	}

	c.replicaVMAverageReplicationLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_vm_average_replication_latency_seconds"),
		"The average time taken to replicate the virtual machine",
		[]string{"vm"},
		nil,
	)
	c.replicaVMAverageReplicationSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_vm_average_replication_size_bytes"),
		"The average size of the replicated data",
		[]string{"vm"},
		nil,
	)
	c.replicaVMCompressionEfficiency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_vm_compression_efficiency"),
		"The compression efficiency of the replicated data",
		[]string{"vm"},
		nil,
	)
	c.replicaVMLastReplicationSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_vm_last_replication_size_bytes"),
		"The size of the last replicated data",
		[]string{"vm"},
		nil,
	)
	c.replicaVMNetworkReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_vm_network_received_bytes_total"),
		"The total number of bytes received over the network for replication",
		[]string{"vm"},
		nil,
	)
	c.replicaVMNetworkSentBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_vm_network_sent_bytes_total"),
		"The total number of bytes sent over the network for replication",
		[]string{"vm"},
		nil,
	)
	c.replicaVMReplicationCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_vm_replications_total"),
		"The total number of replication cycles",
		[]string{"vm"},
		nil,
	)
	c.replicaVMReplicationLatency = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_vm_replication_latency_seconds"),
		"The time taken by the last replication of the virtual machine",
		[]string{"vm"},
		nil,
	)
	c.replicaVMResynchronizedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_vm_resynchronized_bytes_total"),
		"The total number of bytes sent during resynchronization",
		[]string{"vm"},
		nil,
	)
	c.replicaVMHealth = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_vm_health"),
		"The replication health of the virtual machine. 0: Not applicable; 1: OK; 2: Warning; 3: Critical",
		[]string{"vm", "relationship"},
		nil,
	)
	c.replicaVMState = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "replica_vm_state"),
		"The replication state of the virtual machine. 0: Disabled; 1: Ready for replication; 2: Waiting to complete initial replication; "+
			"3: Replicating; 4: Synced replication complete; 5: Recovered; 6: Committed; 7: Suspended; 8: Critical; "+
			"9: Waiting to start resynchronization; 10: Resynchronizing; 11: Resynchronization suspended; 12: Failover in progress; "+
			"13: Failback in progress; 14: Failback complete; 15: Disk update in progress; 16: Disk update critical; 17: Unknown; "+
			"18: Repurpose replication in progress; 19: Prepared for sync replication; 20: Prepared for group reverse replication; "+
			"21: Fire drill in progress",
		[]string{"vm", "relationship"},
		nil,
	)

	return nil
}

func (c *Collector) collectReplicaVM(ch chan<- prometheus.Metric) error {
	return errors.Join(
		c.collectReplicaVMPerfData(ch),
		c.collectReplicaVMRelationships(ch),
	)
}

func (c *Collector) collectReplicaVMPerfData(ch chan<- prometheus.Metric) error {
	if c.perfDataCollectorReplicaVM == nil {
		return nil
	}

	err := c.perfDataCollectorReplicaVM.Collect(&c.perfDataObjectReplicaVM)
	if errors.Is(err, pdh.ErrNoData) {
		// No virtual machine is replicated, or Hyper-V Replica is not supported.
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Replica VM metrics: %w", err)
	}

	for _, data := range c.perfDataObjectReplicaVM {
		ch <- prometheus.MustNewConstMetric(
			c.replicaVMAverageReplicationLatency,
			prometheus.GaugeValue,
			data.AverageReplicationLatency,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.replicaVMAverageReplicationSize,
			prometheus.GaugeValue,
			data.AverageReplicationSize,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.replicaVMCompressionEfficiency,
			prometheus.GaugeValue,
			data.CompressionEfficiency,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.replicaVMLastReplicationSize,
			prometheus.GaugeValue,
			data.LastReplicationSize,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.replicaVMNetworkReceivedBytes,
			prometheus.CounterValue,
			data.NetworkBytesRecv,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.replicaVMNetworkSentBytes,
			prometheus.CounterValue,
			data.NetworkBytesSent,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.replicaVMReplicationCount,
			prometheus.CounterValue,
			data.ReplicationCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.replicaVMReplicationLatency,
			prometheus.GaugeValue,
			data.ReplicationLatency,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.replicaVMResynchronizedBytes,
			prometheus.CounterValue,
			data.ResynchronizedBytes,
			data.Name,
		)
	}

	return nil
}

func (c *Collector) collectReplicaVMRelationships(ch chan<- prometheus.Metric) error {
	var dst []msvmReplicationRelationship

	if err := c.miSession.Query(&dst, mi.NamespaceRootVirtualizationV2, c.miQueryReplicaVM, replicaVMQueryTimeout); err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	for _, relationship := range dst {
		relationshipType := replicationRelationshipType(relationship.InstanceID)

		ch <- prometheus.MustNewConstMetric(
			c.replicaVMHealth,
			prometheus.GaugeValue,
			float64(relationship.ReplicationHealth),
			relationship.ElementName,
			relationshipType,
		)

		ch <- prometheus.MustNewConstMetric(
			c.replicaVMState,
			prometheus.GaugeValue,
			float64(relationship.ReplicationState),
			relationship.ElementName,
			relationshipType,
		)
	}

	return nil
}

// replicationRelationshipType returns the relationship type encoded in the
// InstanceID ("Microsoft:<vm id>\HVR\<0|1>"). 0 is the primary relationship,
// 1 the extended replication.
func replicationRelationshipType(instanceID string) string {
	switch instanceID[strings.LastIndex(instanceID, `\`)+1:] {
	case "0":
		return "primary"
	case "1":
		return "extended"
	default:
		return "unknown"
	}
}
