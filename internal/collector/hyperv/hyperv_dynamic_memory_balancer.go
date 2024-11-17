package hyperv

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorDynamicMemoryBalancer Hyper-V Dynamic Memory Balancer metrics
type collectorDynamicMemoryBalancer struct {
	perfDataCollectorDynamicMemoryBalancer             *perfdata.Collector
	vmDynamicMemoryBalancerAvailableMemoryForBalancing *prometheus.Desc // \Hyper-V Dynamic Memory Balancer(*)\Available Memory For Balancing
	vmDynamicMemoryBalancerSystemCurrentPressure       *prometheus.Desc // \Hyper-V Dynamic Memory Balancer(*)\System Current Pressure
	vmDynamicMemoryBalancerAvailableMemory             *prometheus.Desc // \Hyper-V Dynamic Memory Balancer(*)\Available Memory
	vmDynamicMemoryBalancerAveragePressure             *prometheus.Desc // \Hyper-V Dynamic Memory Balancer(*)\Average Pressure
}

const (
	// Hyper-V Dynamic Memory Balancer metrics
	vmDynamicMemoryBalancerAvailableMemory             = "Available Memory"
	vmDynamicMemoryBalancerAvailableMemoryForBalancing = "Available Memory For Balancing"
	vmDynamicMemoryBalancerAveragePressure             = "Average Pressure"
	vmDynamicMemoryBalancerSystemCurrentPressure       = "System Current Pressure"
)

func (c *Collector) buildDynamicMemoryBalancer() error {
	var err error

	// https://learn.microsoft.com/en-us/archive/blogs/chrisavis/monitoring-dynamic-memory-in-windows-server-hyper-v-2012
	c.perfDataCollectorDynamicMemoryBalancer, err = perfdata.NewCollector("Hyper-V Dynamic Memory Balancer", perfdata.InstanceAll, []string{
		vmDynamicMemoryBalancerAvailableMemory,
		vmDynamicMemoryBalancerAvailableMemoryForBalancing,
		vmDynamicMemoryBalancerAveragePressure,
		vmDynamicMemoryBalancerSystemCurrentPressure,
	})
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Virtual Machine Health Summary collector: %w", err)
	}

	c.vmDynamicMemoryBalancerAvailableMemory = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_balancer_available_memory_bytes"),
		"Represents the amount of memory left on the node.",
		[]string{"balancer"},
		nil,
	)
	c.vmDynamicMemoryBalancerAvailableMemoryForBalancing = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_balancer_available_memory_for_balancing_bytes"),
		"Represents the available memory for balancing purposes.",
		[]string{"balancer"},
		nil,
	)
	c.vmDynamicMemoryBalancerAveragePressure = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_balancer_average_pressure_ratio"),
		"Represents the average system pressure on the balancer node among all balanced objects.",
		[]string{"balancer"},
		nil,
	)
	c.vmDynamicMemoryBalancerSystemCurrentPressure = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dynamic_memory_balancer_system_current_pressure_ratio"),
		"Represents the current pressure in the system.",
		[]string{"balancer"},
		nil,
	)

	return nil
}

func (c *Collector) collectDynamicMemoryBalancer(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollectorDynamicMemoryBalancer.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Dynamic Memory Balancer metrics: %w", err)
	}

	for name, page := range data {
		ch <- prometheus.MustNewConstMetric(
			c.vmDynamicMemoryBalancerAvailableMemory,
			prometheus.GaugeValue,
			utils.MBToBytes(page[vmDynamicMemoryBalancerAvailableMemory].FirstValue),
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmDynamicMemoryBalancerAvailableMemoryForBalancing,
			prometheus.GaugeValue,
			utils.MBToBytes(page[vmDynamicMemoryBalancerAvailableMemoryForBalancing].FirstValue),
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmDynamicMemoryBalancerAveragePressure,
			prometheus.GaugeValue,
			utils.PercentageToRatio(page[vmDynamicMemoryBalancerAveragePressure].FirstValue),
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmDynamicMemoryBalancerSystemCurrentPressure,
			prometheus.GaugeValue,
			utils.PercentageToRatio(page[vmDynamicMemoryBalancerSystemCurrentPressure].FirstValue),
			name,
		)
	}

	return nil
}
