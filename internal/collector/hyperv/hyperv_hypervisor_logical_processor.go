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
	"strings"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorHypervisorLogicalProcessor Hyper-V Hypervisor Logical Processor metrics
type collectorHypervisorLogicalProcessor struct {
	perfDataCollectorHypervisorLogicalProcessor *pdh.Collector
	perfDataObjectHypervisorLogicalProcessor    []perfDataCounterValuesHypervisorLogicalProcessor

	// \Hyper-V Hypervisor Logical Processor(*)\% Guest Run Time
	// \Hyper-V Hypervisor Logical Processor(*)\% Hypervisor Run Time
	// \Hyper-V Hypervisor Logical Processor(*)\% Idle Time
	hypervisorLogicalProcessorTimeTotal         *prometheus.Desc
	hypervisorLogicalProcessorTotalRunTimeTotal *prometheus.Desc // \Hyper-V Hypervisor Logical Processor(*)\% Total Run Time
	hypervisorLogicalProcessorContextSwitches   *prometheus.Desc // \Hyper-V Hypervisor Logical Processor(*)\Context Switches/sec
}

type perfDataCounterValuesHypervisorLogicalProcessor struct {
	Name string

	HypervisorLogicalProcessorGuestRunTimePercent      float64 `perfdata:"% Guest Run Time"`
	HypervisorLogicalProcessorHypervisorRunTimePercent float64 `perfdata:"% Hypervisor Run Time"`
	HypervisorLogicalProcessorTotalRunTimePercent      float64 `perfdata:"% Total Run Time"`
	HypervisorLogicalProcessorIdleRunTimePercent       float64 `perfdata:"% Idle Time"`
	HypervisorLogicalProcessorContextSwitches          float64 `perfdata:"Context Switches/sec"`
}

func (c *Collector) buildHypervisorLogicalProcessor() error {
	var err error

	c.perfDataCollectorHypervisorLogicalProcessor, err = pdh.NewCollector[perfDataCounterValuesHypervisorLogicalProcessor](pdh.CounterTypeRaw, "Hyper-V Hypervisor Logical Processor", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Hypervisor Logical Processor collector: %w", err)
	}

	c.hypervisorLogicalProcessorTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_logical_processor_time_total"),
		"Time that processor spent in different modes (hypervisor, guest, idle)",
		[]string{"core", "state"},
		nil,
	)
	c.hypervisorLogicalProcessorTotalRunTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_logical_processor_total_run_time_total"),
		"Time that processor spent",
		[]string{"core"},
		nil,
	)

	c.hypervisorLogicalProcessorContextSwitches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_logical_processor_context_switches_total"),
		"The rate of virtual processor context switches on the processor.",
		[]string{"core"},
		nil,
	)

	return nil
}

func (c *Collector) collectHypervisorLogicalProcessor(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorHypervisorLogicalProcessor.Collect(&c.perfDataObjectHypervisorLogicalProcessor)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Hypervisor Logical Processor metrics: %w", err)
	}

	for _, data := range c.perfDataObjectHypervisorLogicalProcessor {
		// The name format is Hv LP <core id>
		parts := strings.Split(data.Name, " ")
		if len(parts) != 3 {
			return fmt.Errorf("unexpected Hyper-V Hypervisor Logical Processor name format: %s", data.Name)
		}

		coreID := parts[2]

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorLogicalProcessorTimeTotal,
			prometheus.CounterValue,
			data.HypervisorLogicalProcessorGuestRunTimePercent,
			coreID, "guest",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorLogicalProcessorTimeTotal,
			prometheus.CounterValue,
			data.HypervisorLogicalProcessorHypervisorRunTimePercent,
			coreID, "hypervisor",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorLogicalProcessorTimeTotal,
			prometheus.CounterValue,
			data.HypervisorLogicalProcessorIdleRunTimePercent,
			coreID, "idle",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorLogicalProcessorTotalRunTimeTotal,
			prometheus.CounterValue,
			data.HypervisorLogicalProcessorTotalRunTimePercent,
			coreID,
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorLogicalProcessorContextSwitches,
			prometheus.CounterValue,
			data.HypervisorLogicalProcessorContextSwitches,
			coreID,
		)
	}

	return nil
}
