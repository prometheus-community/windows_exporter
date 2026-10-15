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

// collectorHypervisorRootVirtualProcessor Hyper-V Hypervisor Root Virtual Processor metrics
type collectorHypervisorRootVirtualProcessor struct {
	perfDataCollectorHypervisorRootVirtualProcessor *pdh.Collector[perfDataCounterValuesHypervisorRootVirtualProcessor]
	perfDataObjectHypervisorRootVirtualProcessor    []perfDataCounterValuesHypervisorRootVirtualProcessor

	// \Hyper-V Hypervisor Root Virtual Processor(*)\% Guest Run Time
	// \Hyper-V Hypervisor Root Virtual Processor(*)\% Hypervisor Run Time
	// \Hyper-V Hypervisor Root Virtual Processor(*)\% Remote Run Time
	// \Hyper-V Hypervisor Root Virtual Processor(*)\% Total Run Time
	hypervisorRootVirtualProcessorTimeTotal         *prometheus.Desc
	hypervisorRootVirtualProcessorTotalRunTimeTotal *prometheus.Desc
	hypervisorRootVirtualProcessorCPUWaitSeconds    *prometheus.Desc // \Hyper-V Hypervisor Root Virtual Processor(*)\CPU Wait Time Per Dispatch
	hypervisorRootVirtualProcessorDispatches        *prometheus.Desc // \Hyper-V Hypervisor Root Virtual Processor(*)\Logical Processor Dispatches/sec
}

type perfDataCounterValuesHypervisorRootVirtualProcessor struct {
	Name string

	// The run time counters and CPU Wait Time Per Dispatch are PERF_PRECISION_100NS_TIMER.
	// Their raw values are cumulative times, which pdh scales to seconds. Despite its name
	// and help text, CPU Wait Time Per Dispatch is the total wait time, not an average.
	HypervisorRootVirtualProcessorGuestRunTimePercent      float64 `perfdata:"% Guest Run Time"`
	HypervisorRootVirtualProcessorHypervisorRunTimePercent float64 `perfdata:"% Hypervisor Run Time"`
	HypervisorRootVirtualProcessorTotalRunTimePercent      float64 `perfdata:"% Total Run Time"`
	HypervisorRootVirtualProcessorRemoteRunTimePercent     float64 `perfdata:"% Remote Run Time"`
	HypervisorRootVirtualProcessorCPUWaitTimePerDispatch   float64 `perfdata:"CPU Wait Time Per Dispatch"`
	HypervisorRootVirtualProcessorDispatches               float64 `perfdata:"Logical Processor Dispatches/sec"`
}

func (c *Collector) buildHypervisorRootVirtualProcessor() error {
	var err error

	c.perfDataCollectorHypervisorRootVirtualProcessor, err = pdh.NewCollector[perfDataCounterValuesHypervisorRootVirtualProcessor](c.logger, pdh.CounterTypeRaw, "Hyper-V Hypervisor Root Virtual Processor", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Hypervisor Root Virtual Processor collector: %w", err)
	}

	c.buildHypervisorRootVirtualProcessorDescriptors()

	return nil
}

func (c *Collector) buildHypervisorRootVirtualProcessorDescriptors() {
	c.hypervisorRootVirtualProcessorTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_root_virtual_processor_time_seconds_total"),
		"Time in seconds that processor spent in different modes (guest_run, hypervisor, remote)",
		[]string{"core", "state"},
		nil,
	)

	c.hypervisorRootVirtualProcessorTotalRunTimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_root_virtual_processor_total_run_time_seconds_total"),
		"Time in seconds that processor spent running",
		[]string{"core"},
		nil,
	)

	c.hypervisorRootVirtualProcessorCPUWaitSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_root_virtual_processor_cpu_wait_seconds_total"),
		"The time spent waiting for a virtual processor to be dispatched onto a logical processor. "+
			"Divide its rate by the rate of windows_hyperv_hypervisor_root_virtual_processor_dispatches_total for the average wait per dispatch.",
		[]string{"core"},
		nil,
	)

	c.hypervisorRootVirtualProcessorDispatches = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "hypervisor_root_virtual_processor_dispatches_total"),
		"The number of times the virtual processor was dispatched onto a logical processor.",
		[]string{"core"},
		nil,
	)
}

func (c *Collector) collectHypervisorRootVirtualProcessor(ch chan<- prometheus.Metric, scrape *labelcache.Scrape) error {
	err := c.perfDataCollectorHypervisorRootVirtualProcessor.Collect(&c.perfDataObjectHypervisorRootVirtualProcessor)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Hypervisor Root Virtual Processor metrics: %w", err)
	}

	return c.publishHypervisorRootVirtualProcessor(ch, scrape)
}

func (c *Collector) publishHypervisorRootVirtualProcessor(ch chan<- prometheus.Metric, scrape *labelcache.Scrape) error {
	for _, data := range c.perfDataObjectHypervisorRootVirtualProcessor {
		// The name format is Root VP <core id>
		parts := strings.Split(data.Name, " ")
		if len(parts) != 3 {
			return fmt.Errorf("unexpected Hyper-V Hypervisor Root Virtual Processor name format: %s", data.Name)
		}

		coreID := parts[2]

		ch <- scrape.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorTimeTotal,
			prometheus.CounterValue,
			data.HypervisorRootVirtualProcessorGuestRunTimePercent,
			coreID, "guest_run",
		)

		ch <- scrape.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorTimeTotal,
			prometheus.CounterValue,
			data.HypervisorRootVirtualProcessorHypervisorRunTimePercent,
			coreID, "hypervisor",
		)

		ch <- scrape.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorTimeTotal,
			prometheus.CounterValue,
			data.HypervisorRootVirtualProcessorRemoteRunTimePercent,
			coreID, "remote",
		)

		ch <- scrape.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorTotalRunTimeTotal,
			prometheus.CounterValue,
			data.HypervisorRootVirtualProcessorTotalRunTimePercent,
			coreID,
		)

		ch <- scrape.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorCPUWaitSeconds,
			prometheus.CounterValue,
			data.HypervisorRootVirtualProcessorCPUWaitTimePerDispatch,
			coreID,
		)

		ch <- scrape.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorDispatches,
			prometheus.CounterValue,
			data.HypervisorRootVirtualProcessorDispatches,
			coreID,
		)
	}

	return nil
}
