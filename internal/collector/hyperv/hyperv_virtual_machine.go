package hyperv

import (
	"errors"
	"fmt"
	"strings"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/perfdata/perftypes"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Hyper-V Virtual Machine Health Summary
	healthCritical = "Health Critical"
	healthOk       = "Health Ok"

	// Hyper-V VM Vid Partition
	physicalPagesAllocated = "Physical Pages Allocated"
	preferredNUMANodeIndex = "Preferred NUMA Node Index"
	remotePhysicalPages    = "Remote Physical Pages"
)

func (c *Collector) buildVirtualMachine() error {
	var err error

	c.perfDataCollectorVirtualMachineHealthSummary, err = perfdata.NewCollector(perfdata.V2, "Hyper-V Virtual Machine Health Summary", perfdata.AllInstances, []string{
		healthCritical,
		healthOk,
	})

	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Virtual Machine Health Summary collector: %w", err)
	}

	c.healthCritical = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "health_critical"),
		"This counter represents the number of virtual machines with critical health",
		nil,
		nil,
	)
	c.healthOk = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "health_ok"),
		"This counter represents the number of virtual machines with ok health",
		nil,
		nil,
	)

	c.perfDataCollectorVMVidPartition, err = perfdata.NewCollector(perfdata.V2, "Hyper-V VM Vid Partition", perfdata.AllInstances, []string{
		physicalPagesAllocated,
		preferredNUMANodeIndex,
		remotePhysicalPages,
	})

	if err != nil {
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

func (c *Collector) collectVirtualMachine(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollectorVirtualMachineHealthSummary.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Virtual Machine Health Summary metrics: %w", err)
	} else if len(data) == 0 {
		return errors.New("perflib query for Hyper-V Virtual Machine Health Summary returned empty result set")
	}

	healthData, ok := data[perftypes.EmptyInstance]
	if !ok {
		return errors.New("no data returned for Hyper-V Virtual Machine Health Summary")
	}

	ch <- prometheus.MustNewConstMetric(
		c.healthCritical,
		prometheus.GaugeValue,
		healthData[healthCritical].FirstValue,
	)

	ch <- prometheus.MustNewConstMetric(
		c.healthOk,
		prometheus.GaugeValue,
		healthData[healthOk].FirstValue,
	)

	data, err = c.perfDataCollectorVMVidPartition.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Virtual Machine Health Summary metrics: %w", err)
	} else if len(data) == 0 {
		return errors.New("perflib query for Hyper-V Virtual Machine Health Summary returned empty result set")
	}

	for name, page := range data {
		if strings.Contains(name, "*") {
			continue
		}

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
