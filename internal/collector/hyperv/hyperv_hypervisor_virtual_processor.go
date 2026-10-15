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
	"fmt"
	"strings"

	"github.com/prometheus-community/windows_exporter/internal/labelcache"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorHypervisorVirtualProcessor Hyper-V Hypervisor Virtual Processor metrics
type collectorHypervisorVirtualProcessor struct {
	perfDataCollectorHypervisorVirtualProcessor *pdh.Collector[perfDataCounterValuesHypervisorVirtualProcessor]
	perfDataObjectHypervisorVirtualProcessor    []perfDataCounterValuesHypervisorVirtualProcessor

	// \Hyper-V Hypervisor Virtual Processor(*)\% Guest Run Time
	// \Hyper-V Hypervisor Virtual Processor(*)\% Hypervisor Run Time
	// \Hyper-V Hypervisor Virtual Processor(*)\% Remote Run Time
	hypervisorVirtualProcessorModeTimeTotal  *prometheus.Desc
	hypervisorVirtualProcessorRunTimeTotal   *prometheus.Desc
	hypervisorVirtualProcessorCPUWaitSeconds *prometheus.Desc // \Hyper-V Hypervisor Virtual Processor(*)\CPU Wait Time Per Dispatch
	hypervisorVirtualProcessorDispatches     *prometheus.Desc // \Hyper-V Hypervisor Virtual Processor(*)\Logical Processor Dispatches/sec
}

type perfDataCounterValuesHypervisorVirtualProcessor struct {
	Name string

	// The run time counters and CPU Wait Time Per Dispatch are PERF_PRECISION_100NS_TIMER.
	// Their raw values are cumulative times, which pdh scales to seconds. Despite its name
	// and help text, CPU Wait Time Per Dispatch is the total wait time, not an average.
	HypervisorVirtualProcessorGuestRunTimePercent      float64 `perfdata:"% Guest Run Time"`
	HypervisorVirtualProcessorHypervisorRunTimePercent float64 `perfdata:"% Hypervisor Run Time"`
	HypervisorVirtualProcessorTotalRunTimePercent      float64 `perfdata:"% Total Run Time"`
	HypervisorVirtualProcessorRemoteRunTimePercent     float64 `perfdata:"% Remote Run Time"`
	HypervisorVirtualProcessorCPUWaitTimePerDispatch   float64 `perfdata:"CPU Wait Time Per Dispatch"`
	HypervisorVirtualProcessorDispatches               float64 `perfdata:"Logical Processor Dispatches/sec"`
}

func (c *Collector) buildHypervisorVirtualProcessor() error {
	var err error

	c.perfDataCollectorHypervisorVirtualProcessor, err = pdh.NewCollector[perfDataCounterValuesHypervisorVirtualProcessor](c.logger, pdh.CounterTypeRaw, "Hyper-V Hypervisor Virtual Processor", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Hypervisor Virtual Processor collector: %w", err)
	}

	c.buildHypervisorVirtualProcessorDescriptors()

	return nil
}

func (c *Collector) buildHypervisorVirtualProcessorDescriptors() {
	c.hypervisorVirtualProcessorModeTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_virtual_processor_mode_time_seconds_total"),
		"Time in seconds that processor spent in different modes (hypervisor, guest, remote)",
		[]string{"vm", "core", "state"},
		nil,
	)
	c.hypervisorVirtualProcessorRunTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_virtual_processor_run_time_seconds_total"),
		"Time in seconds that processor spent running",
		[]string{"vm", "core"},
		nil,
	)
	c.hypervisorVirtualProcessorCPUWaitSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_virtual_processor_cpu_wait_seconds_total"),
		"The time spent waiting for a virtual processor to be dispatched onto a logical processor. "+
			"Divide its rate by the rate of windows_hyperv_hypervisor_virtual_processor_dispatches_total for the average wait per dispatch.",
		[]string{"vm", "core"},
		nil,
	)
	c.hypervisorVirtualProcessorDispatches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_virtual_processor_dispatches_total"),
		"The number of times the virtual processor was dispatched onto a logical processor.",
		[]string{"vm", "core"},
		nil,
	)
}

func (c *Collector) collectHypervisorVirtualProcessor(ch chan<- prometheus.Metric, scrape *labelcache.Scrape) error {
	err := c.perfDataCollectorHypervisorVirtualProcessor.Collect(&c.perfDataObjectHypervisorVirtualProcessor)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Hypervisor Virtual Processor metrics: %w", err)
	}

	return c.publishHypervisorVirtualProcessor(ch, scrape)
}

func (c *Collector) publishHypervisorVirtualProcessor(ch chan<- prometheus.Metric, scrape *labelcache.Scrape) error {
	for _, data := range c.perfDataObjectHypervisorVirtualProcessor {
		// The name format is <VM Name>:Hv VP <vcore id>
		parts := strings.Split(data.Name, ":")
		if len(parts) != 2 {
			return fmt.Errorf("unexpected format of Name in Hyper-V Hypervisor Virtual Processor: %q, expected %q", data.Name, "<VM Name>:Hv VP <vcore id>")
		}

		coreParts := strings.Split(parts[1], " ")
		if len(coreParts) != 3 {
			return fmt.Errorf("unexpected format of core identifier in Hyper-V Hypervisor Virtual Processor: %q, expected %q", parts[1], "Hv VP <vcore id>")
		}

		vmName := parts[0]
		coreID := coreParts[2]

		// Same metric with new name for better clarity, old one is kept for backward compatibility
		ch <- scrape.MustNewConstMetric(
			c.hypervisorVirtualProcessorModeTimeTotal,
			prometheus.CounterValue,
			data.HypervisorVirtualProcessorHypervisorRunTimePercent,
			vmName, coreID, "hypervisor",
		)

		ch <- scrape.MustNewConstMetric(
			c.hypervisorVirtualProcessorModeTimeTotal,
			prometheus.CounterValue,
			data.HypervisorVirtualProcessorGuestRunTimePercent,
			vmName, coreID, "guest",
		)

		ch <- scrape.MustNewConstMetric(
			c.hypervisorVirtualProcessorModeTimeTotal,
			prometheus.CounterValue,
			data.HypervisorVirtualProcessorRemoteRunTimePercent,
			vmName, coreID, "remote",
		)
		// end same metric
		// Same metric with new name for better clarity, old one is kept for backward compatibility
		ch <- scrape.MustNewConstMetric(
			c.hypervisorVirtualProcessorRunTimeTotal,
			prometheus.CounterValue,
			data.HypervisorVirtualProcessorTotalRunTimePercent,
			vmName, coreID,
		)
		// end same metric
		ch <- scrape.MustNewConstMetric(
			c.hypervisorVirtualProcessorCPUWaitSeconds,
			prometheus.CounterValue,
			data.HypervisorVirtualProcessorCPUWaitTimePerDispatch,
			vmName, coreID,
		)

		ch <- scrape.MustNewConstMetric(
			c.hypervisorVirtualProcessorDispatches,
			prometheus.CounterValue,
			data.HypervisorVirtualProcessorDispatches,
			vmName, coreID,
		)
	}

	return nil
}
