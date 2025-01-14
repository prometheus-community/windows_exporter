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

// collectorHypervisorRootVirtualProcessor Hyper-V Hypervisor Root Virtual Processor metrics
type collectorHypervisorRootVirtualProcessor struct {
	perfDataCollectorHypervisorRootVirtualProcessor *pdh.Collector
	perfDataObjectHypervisorRootVirtualProcessor    []perfDataCounterValuesHypervisorRootVirtualProcessor

	// \Hyper-V Hypervisor Root Virtual Processor(*)\% Guest Run Time
	// \Hyper-V Hypervisor Root Virtual Processor(*)\% Hypervisor Run Time
	// \Hyper-V Hypervisor Root Virtual Processor(*)\% Remote Run Time
	// \Hyper-V Hypervisor Root Virtual Processor(*)\% Total Run Time
	hypervisorRootVirtualProcessorTimeTotal              *prometheus.Desc
	hypervisorRootVirtualProcessorTotalRunTimeTotal      *prometheus.Desc
	hypervisorRootVirtualProcessorCPUWaitTimePerDispatch *prometheus.Desc // \Hyper-V Hypervisor Root Virtual Processor(*)\CPU Wait Time Per Dispatch
}

type perfDataCounterValuesHypervisorRootVirtualProcessor struct {
	Name string

	HypervisorRootVirtualProcessorGuestRunTimePercent      float64 `perfdata:"% Guest Run Time"`
	HypervisorRootVirtualProcessorHypervisorRunTimePercent float64 `perfdata:"% Hypervisor Run Time"`
	HypervisorRootVirtualProcessorTotalRunTimePercent      float64 `perfdata:"% Total Run Time"`
	HypervisorRootVirtualProcessorRemoteRunTimePercent     float64 `perfdata:"% Remote Run Time"`
	HypervisorRootVirtualProcessorCPUWaitTimePerDispatch   float64 `perfdata:"CPU Wait Time Per Dispatch"`
}

func (c *Collector) buildHypervisorRootVirtualProcessor() error {
	var err error

	c.perfDataCollectorHypervisorRootVirtualProcessor, err = pdh.NewCollector[perfDataCounterValuesHypervisorRootVirtualProcessor](pdh.CounterTypeRaw, "Hyper-V Hypervisor Root Virtual Processor", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Hypervisor Root Virtual Processor collector: %w", err)
	}

	c.hypervisorRootVirtualProcessorTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_root_virtual_processor_time_total"),
		"Time that processor spent in different modes (hypervisor, guest_run, guest_idle, remote)",
		[]string{"core", "state"},
		nil,
	)

	c.hypervisorRootVirtualProcessorTotalRunTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_root_virtual_processor_total_run_time_total"),
		"Time that processor spent",
		[]string{"core"},
		nil,
	)

	c.hypervisorRootVirtualProcessorCPUWaitTimePerDispatch = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_root_virtual_cpu_wait_time_per_dispatch_total"),
		"The average time (in nanoseconds) spent waiting for a virtual processor to be dispatched onto a logical processor.",
		[]string{"core"},
		nil,
	)

	return nil
}

func (c *Collector) collectHypervisorRootVirtualProcessor(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorHypervisorRootVirtualProcessor.Collect(&c.perfDataObjectHypervisorRootVirtualProcessor)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Hypervisor Root Virtual Processor metrics: %w", err)
	}

	for _, data := range c.perfDataObjectHypervisorRootVirtualProcessor {
		// The name format is Hv LP <core id>
		parts := strings.Split(data.Name, " ")
		if len(parts) != 3 {
			return fmt.Errorf("unexpected Hyper-V Hypervisor Root Virtual Processor name format: %s", data.Name)
		}

		coreID := parts[2]

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorTimeTotal,
			prometheus.CounterValue,
			data.HypervisorRootVirtualProcessorGuestRunTimePercent,
			coreID, "guest_run",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorTimeTotal,
			prometheus.CounterValue,
			data.HypervisorRootVirtualProcessorHypervisorRunTimePercent,
			coreID, "hypervisor",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorTimeTotal,
			prometheus.CounterValue,
			data.HypervisorRootVirtualProcessorRemoteRunTimePercent,
			coreID, "remote",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorTotalRunTimeTotal,
			prometheus.CounterValue,
			data.HypervisorRootVirtualProcessorTotalRunTimePercent,
			coreID,
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorCPUWaitTimePerDispatch,
			prometheus.CounterValue,
			data.HypervisorRootVirtualProcessorCPUWaitTimePerDispatch,
			coreID,
		)
	}

	return nil
}
