package hyperv

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorDynamicMemoryVM Hyper-V Dynamic Memory VM metrics
type collectorDynamicMemoryVM struct {
	perfDataCollectorDynamicMemoryVM   *perfdata.Collector
	vmMemoryAddedMemory                *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Added Memory
	vmMemoryCurrentPressure            *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Current Pressure
	vmMemoryGuestVisiblePhysicalMemory *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Guest Visible Physical Memory
	vmMemoryMaximumPressure            *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Maximum Pressure
	vmMemoryMemoryAddOperations        *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Memory Add Operations
	vmMemoryMemoryRemoveOperations     *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Memory Remove Operations
	vmMemoryMinimumPressure            *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Minimum Pressure
	vmMemoryPhysicalMemory             *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Physical Memory
	vmMemoryRemovedMemory              *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Removed Memory
	vmMemoryGuestAvailableMemory       *prometheus.Desc // \Hyper-V Dynamic Memory VM(*)\Guest Available Memory
}

const (
	// Hyper-V Dynamic Memory VM metrics
	vmMemoryAddedMemory                = "Added Memory"
	vmMemoryCurrentPressure            = "Current Pressure"
	vmMemoryGuestAvailableMemory       = "Guest Available Memory"
	vmMemoryGuestVisiblePhysicalMemory = "Guest Visible Physical Memory"
	vmMemoryMaximumPressure            = "Maximum Pressure"
	vmMemoryMemoryAddOperations        = "Memory Add Operations"
	vmMemoryMemoryRemoveOperations     = "Memory Remove Operations"
	vmMemoryMinimumPressure            = "Minimum Pressure"
	vmMemoryPhysicalMemory             = "Physical Memory"
	vmMemoryRemovedMemory              = "Removed Memory"
)

func (c *Collector) buildDynamicMemoryVM() error {
	var err error

	c.perfDataCollectorDynamicMemoryVM, err = perfdata.NewCollector("Hyper-V Dynamic Memory VM", perfdata.InstanceAll, []string{
		vmMemoryAddedMemory,
		vmMemoryCurrentPressure,
		vmMemoryGuestVisiblePhysicalMemory,
		vmMemoryMaximumPressure,
		vmMemoryMemoryAddOperations,
		vmMemoryMemoryRemoveOperations,
		vmMemoryMinimumPressure,
		vmMemoryPhysicalMemory,
		vmMemoryRemovedMemory,
		vmMemoryGuestAvailableMemory,
	})
	if err != nil && !errors.Is(err, perfdata.ErrNoData) {
		return fmt.Errorf("failed to create Hyper-V Dynamic Memory VM collector: %w", err)
	}

	c.vmMemoryAddedMemory = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_added_total"),
		"Represents the cumulative amount of memory added to the VM.",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryCurrentPressure = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_pressure_current_ratio"),
		"Represents the current pressure in the VM.",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryGuestAvailableMemory = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_guest_available_bytes"),
		"Represents the current amount of available memory in the VM (reported by the VM).",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryGuestVisiblePhysicalMemory = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_guest_visible_physical_memory_bytes"),
		"Represents the amount of memory visible in the VM.'",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryMaximumPressure = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_pressure_maximum_ratio"),
		"Represents the maximum pressure band in the VM.",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryMemoryAddOperations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_add_operations_total"),
		"Represents the total number of add operations for the VM.",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryMemoryRemoveOperations = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_remove_operations_total"),
		"Represents the total number of remove operations for the VM.",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryMinimumPressure = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_pressure_minimum_ratio"),
		"Represents the minimum pressure band in the VM.",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryPhysicalMemory = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_physical_bytes"),
		"Represents the current amount of memory in the VM.",
		[]string{"vm"},
		nil,
	)
	c.vmMemoryRemovedMemory = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_vm_removed_bytes_total"),
		"Represents the cumulative amount of memory removed from the VM.",
		[]string{"vm"},
		nil,
	)

	return nil
}

func (c *Collector) collectDynamicMemoryVM(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollectorDynamicMemoryVM.Collect()
	if err != nil && !errors.Is(err, perfdata.ErrNoData) {
		return fmt.Errorf("failed to collect Hyper-V Dynamic Memory VM metrics: %w", err)
	}

	for vmName, vmData := range data {
		ch <- prometheus.MustNewConstMetric(
			c.vmMemoryAddedMemory,
			prometheus.CounterValue,
			utils.MBToBytes(vmData[vmMemoryAddedMemory].FirstValue),
			vmName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmMemoryCurrentPressure,
			prometheus.GaugeValue,
			utils.PercentageToRatio(vmData[vmMemoryCurrentPressure].FirstValue),
			vmName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmMemoryGuestAvailableMemory,
			prometheus.GaugeValue,
			utils.MBToBytes(vmData[vmMemoryGuestAvailableMemory].FirstValue),
			vmName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmMemoryGuestVisiblePhysicalMemory,
			prometheus.GaugeValue,
			utils.MBToBytes(vmData[vmMemoryGuestVisiblePhysicalMemory].FirstValue),
			vmName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmMemoryMaximumPressure,
			prometheus.GaugeValue,
			utils.PercentageToRatio(vmData[vmMemoryMaximumPressure].FirstValue),
			vmName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmMemoryMemoryAddOperations,
			prometheus.CounterValue,
			vmData[vmMemoryMemoryAddOperations].FirstValue,
			vmName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmMemoryMemoryRemoveOperations,
			prometheus.CounterValue,
			vmData[vmMemoryMemoryRemoveOperations].FirstValue,
			vmName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmMemoryMinimumPressure,
			prometheus.GaugeValue,
			utils.PercentageToRatio(vmData[vmMemoryMinimumPressure].FirstValue),
			vmName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmMemoryPhysicalMemory,
			prometheus.GaugeValue,
			utils.MBToBytes(vmData[vmMemoryPhysicalMemory].FirstValue),
			vmName,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmMemoryRemovedMemory,
			prometheus.CounterValue,
			utils.MBToBytes(vmData[vmMemoryRemovedMemory].FirstValue),
			vmName,
		)
	}

	return nil
}
