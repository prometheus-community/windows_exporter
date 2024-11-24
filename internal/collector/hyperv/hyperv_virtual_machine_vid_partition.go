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

package hyperv

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorVirtualMachineVidPartition Hyper-V VM Vid Partition metrics
type collectorVirtualMachineVidPartition struct {
	perfDataCollectorVirtualMachineVidPartition *perfdata.Collector
	physicalPagesAllocated                      *prometheus.Desc // \Hyper-V VM Vid Partition(*)\Physical Pages Allocated
	preferredNUMANodeIndex                      *prometheus.Desc // \Hyper-V VM Vid Partition(*)\Preferred NUMA Node Index
	remotePhysicalPages                         *prometheus.Desc // \Hyper-V VM Vid Partition(*)\Remote Physical Pages
}

const (
	physicalPagesAllocated = "Physical Pages Allocated"
	preferredNUMANodeIndex = "Preferred NUMA Node Index"
	remotePhysicalPages    = "Remote Physical Pages"
)

func (c *Collector) buildVirtualMachineVidPartition() error {
	var err error

	c.perfDataCollectorVirtualMachineVidPartition, err = perfdata.NewCollector("Hyper-V VM Vid Partition", perfdata.InstancesAll, []string{
		physicalPagesAllocated,
		preferredNUMANodeIndex,
		remotePhysicalPages,
	})
	if err != nil && !errors.Is(err, perfdata.ErrNoData) {
		return fmt.Errorf("failed to create Hyper-V VM Vid Partition collector: %w", err)
	}

	c.physicalPagesAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vid_physical_pages_allocated"),
		"The number of physical pages allocated",
		[]string{"vm"},
		nil,
	)
	c.preferredNUMANodeIndex = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vid_preferred_numa_node_index"),
		"The preferred NUMA node index associated with this partition",
		[]string{"vm"},
		nil,
	)
	c.remotePhysicalPages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "vid_remote_physical_pages"),
		"The number of physical pages not allocated from the preferred NUMA node",
		[]string{"vm"},
		nil,
	)

	return nil
}

func (c *Collector) collectVirtualMachineVidPartition(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollectorVirtualMachineVidPartition.Collect()
	if err != nil && !errors.Is(err, perfdata.ErrNoData) {
		return fmt.Errorf("failed to collect Hyper-V VM Vid Partition metrics: %w", err)
	}

	for name, page := range data {
		ch <- prometheus.MustNewConstMetric(
			c.physicalPagesAllocated,
			prometheus.GaugeValue,
			page[physicalPagesAllocated].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.preferredNUMANodeIndex,
			prometheus.GaugeValue,
			page[preferredNUMANodeIndex].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.remotePhysicalPages,
			prometheus.GaugeValue,
			page[remotePhysicalPages].FirstValue,
			name,
		)
	}

	return nil
}
