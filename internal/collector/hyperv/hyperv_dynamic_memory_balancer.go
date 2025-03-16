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

package hyperv

import (
	"fmt"

	"github.com/Microsoft/hcsshim/osversion"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorDynamicMemoryBalancer Hyper-V Dynamic Memory Balancer metrics
type collectorDynamicMemoryBalancer struct {
	perfDataCollectorDynamicMemoryBalancer *pdh.Collector
	perfDataObjectDynamicMemoryBalancer    []perfDataCounterValuesDynamicMemoryBalancer

	vmDynamicMemoryBalancerAvailableMemoryForBalancing *prometheus.Desc // \Hyper-V Dynamic Memory Balancer(*)\Available Memory For Balancing
	vmDynamicMemoryBalancerSystemCurrentPressure       *prometheus.Desc // \Hyper-V Dynamic Memory Balancer(*)\System Current Pressure
	vmDynamicMemoryBalancerAvailableMemory             *prometheus.Desc // \Hyper-V Dynamic Memory Balancer(*)\Available Memory
	vmDynamicMemoryBalancerAveragePressure             *prometheus.Desc // \Hyper-V Dynamic Memory Balancer(*)\Average Pressure
}

type perfDataCounterValuesDynamicMemoryBalancer struct {
	Name string

	// Hyper-V Dynamic Memory Balancer metrics
	VmDynamicMemoryBalancerAvailableMemory             float64 `perfdata:"Available Memory"`
	VmDynamicMemoryBalancerAvailableMemoryForBalancing float64 `perfdata:"Available Memory For Balancing" perfdata_min_build:"17763"`
	VmDynamicMemoryBalancerAveragePressure             float64 `perfdata:"Average Pressure"`
	VmDynamicMemoryBalancerSystemCurrentPressure       float64 `perfdata:"System Current Pressure"`
}

func (c *Collector) buildDynamicMemoryBalancer() error {
	var err error

	// https://learn.microsoft.com/en-us/archive/blogs/chrisavis/monitoring-dynamic-memory-in-windows-server-hyper-v-2012
	c.perfDataCollectorDynamicMemoryBalancer, err = pdh.NewCollector[perfDataCounterValuesDynamicMemoryBalancer](pdh.CounterTypeRaw, "Hyper-V Dynamic Memory Balancer", pdh.InstancesAll)
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
	err := c.perfDataCollectorDynamicMemoryBalancer.Collect(&c.perfDataObjectDynamicMemoryBalancer)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Dynamic Memory Balancer metrics: %w", err)
	}

	for _, data := range c.perfDataObjectDynamicMemoryBalancer {
		ch <- prometheus.MustNewConstMetric(
			c.vmDynamicMemoryBalancerAvailableMemory,
			prometheus.GaugeValue,
			utils.MBToBytes(data.VmDynamicMemoryBalancerAvailableMemory),
			data.Name,
		)

		if osversion.Build() >= osversion.LTSC2019 {
			ch <- prometheus.MustNewConstMetric(
				c.vmDynamicMemoryBalancerAvailableMemoryForBalancing,
				prometheus.GaugeValue,
				utils.MBToBytes(data.VmDynamicMemoryBalancerAvailableMemoryForBalancing),
				data.Name,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.vmDynamicMemoryBalancerAveragePressure,
			prometheus.GaugeValue,
			utils.PercentageToRatio(data.VmDynamicMemoryBalancerAveragePressure),
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.vmDynamicMemoryBalancerSystemCurrentPressure,
			prometheus.GaugeValue,
			utils.PercentageToRatio(data.VmDynamicMemoryBalancerSystemCurrentPressure),
			data.Name,
		)
	}

	return nil
}
