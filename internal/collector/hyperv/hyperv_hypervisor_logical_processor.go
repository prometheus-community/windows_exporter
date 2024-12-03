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

	// \Hyper-V Hypervisor Logical Processor(*)\% Guest Run Time
	// \Hyper-V Hypervisor Logical Processor(*)\% Hypervisor Run Time
	// \Hyper-V Hypervisor Logical Processor(*)\% Idle Time
	hypervisorLogicalProcessorTimeTotal         *prometheus.Desc
	hypervisorLogicalProcessorTotalRunTimeTotal *prometheus.Desc // \Hyper-V Hypervisor Logical Processor(*)\% Total Run Time
	hypervisorLogicalProcessorContextSwitches   *prometheus.Desc // \Hyper-V Hypervisor Logical Processor(*)\Context Switches/sec
}

const (
	hypervisorLogicalProcessorGuestRunTimePercent      = "% Guest Run Time"
	hypervisorLogicalProcessorHypervisorRunTimePercent = "% Hypervisor Run Time"
	hypervisorLogicalProcessorTotalRunTimePercent      = "% Total Run Time"
	hypervisorLogicalProcessorIdleRunTimePercent       = "% Idle Time"
	hypervisorLogicalProcessorContextSwitches          = "Context Switches/sec"
)

func (c *Collector) buildHypervisorLogicalProcessor() error {
	var err error

	c.perfDataCollectorHypervisorLogicalProcessor, err = pdh.NewCollector("Hyper-V Hypervisor Logical Processor", pdh.InstancesAll, []string{
		hypervisorLogicalProcessorGuestRunTimePercent,
		hypervisorLogicalProcessorHypervisorRunTimePercent,
		hypervisorLogicalProcessorTotalRunTimePercent,
		hypervisorLogicalProcessorIdleRunTimePercent,
		hypervisorLogicalProcessorContextSwitches,
	})
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
	data, err := c.perfDataCollectorHypervisorLogicalProcessor.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Hypervisor Logical Processor metrics: %w", err)
	}

	for coreName, coreData := range data {
		// The name format is Hv LP <core id>
		parts := strings.Split(coreName, " ")
		if len(parts) != 3 {
			return fmt.Errorf("unexpected Hyper-V Hypervisor Logical Processor name format: %s", coreName)
		}

		coreId := parts[2]

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorLogicalProcessorTimeTotal,
			prometheus.CounterValue,
			coreData[hypervisorLogicalProcessorGuestRunTimePercent].FirstValue,
			coreId, "guest",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorLogicalProcessorTimeTotal,
			prometheus.CounterValue,
			coreData[hypervisorLogicalProcessorHypervisorRunTimePercent].FirstValue,
			coreId, "hypervisor",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorLogicalProcessorTimeTotal,
			prometheus.CounterValue,
			coreData[hypervisorLogicalProcessorIdleRunTimePercent].FirstValue,
			coreId, "idle",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorLogicalProcessorTotalRunTimeTotal,
			prometheus.CounterValue,
			coreData[hypervisorLogicalProcessorTotalRunTimePercent].FirstValue,
			coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorLogicalProcessorContextSwitches,
			prometheus.CounterValue,
			coreData[hypervisorLogicalProcessorContextSwitches].FirstValue,
			coreId,
		)
	}

	return nil
}
