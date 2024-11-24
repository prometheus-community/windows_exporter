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

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorHypervisorRootVirtualProcessor Hyper-V Hypervisor Root Virtual Processor metrics
type collectorHypervisorRootVirtualProcessor struct {
	perfDataCollectorHypervisorRootVirtualProcessor *perfdata.Collector

	// \Hyper-V Hypervisor Root Virtual Processor(*)\% Guest Idle Time
	// \Hyper-V Hypervisor Root Virtual Processor(*)\% Guest Run Time
	// \Hyper-V Hypervisor Root Virtual Processor(*)\% Hypervisor Run Time
	// \Hyper-V Hypervisor Root Virtual Processor(*)\% Remote Run Time
	// \Hyper-V Hypervisor Root Virtual Processor(*)\% Total Run Time
	hypervisorRootVirtualProcessorTimeTotal              *prometheus.Desc
	hypervisorRootVirtualProcessorTotalRunTimeTotal      *prometheus.Desc
	hypervisorRootVirtualProcessorCPUWaitTimePerDispatch *prometheus.Desc // \Hyper-V Hypervisor Root Virtual Processor(*)\CPU Wait Time Per Dispatch
}

const (
	hypervisorRootVirtualProcessorGuestIdleTimePercent     = "% Guest Idle Time"
	hypervisorRootVirtualProcessorGuestRunTimePercent      = "% Guest Run Time"
	hypervisorRootVirtualProcessorHypervisorRunTimePercent = "% Hypervisor Run Time"
	hypervisorRootVirtualProcessorTotalRunTimePercent      = "% Total Run Time"
	hypervisorRootVirtualProcessorRemoteRunTimePercent     = "% Remote Run Time"
	hypervisorRootVirtualProcessorCPUWaitTimePerDispatch   = "CPU Wait Time Per Dispatch"
)

func (c *Collector) buildHypervisorRootVirtualProcessor() error {
	var err error

	c.perfDataCollectorHypervisorRootVirtualProcessor, err = perfdata.NewCollector("Hyper-V Hypervisor Root Virtual Processor", perfdata.InstancesAll, []string{
		hypervisorRootVirtualProcessorGuestIdleTimePercent,
		hypervisorRootVirtualProcessorGuestRunTimePercent,
		hypervisorRootVirtualProcessorHypervisorRunTimePercent,
		hypervisorRootVirtualProcessorTotalRunTimePercent,
		hypervisorRootVirtualProcessorRemoteRunTimePercent,
		hypervisorRootVirtualProcessorCPUWaitTimePerDispatch,
	})
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
	data, err := c.perfDataCollectorHypervisorRootVirtualProcessor.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Hypervisor Root Virtual Processor metrics: %w", err)
	}

	for coreName, coreData := range data {
		// The name format is Hv LP <core id>
		parts := strings.Split(coreName, " ")
		if len(parts) != 3 {
			return fmt.Errorf("unexpected Hyper-V Hypervisor Root Virtual Processor name format: %s", coreName)
		}

		coreId := parts[2]

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorTimeTotal,
			prometheus.CounterValue,
			coreData[hypervisorRootVirtualProcessorGuestRunTimePercent].FirstValue,
			coreId, "guest_run",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorTimeTotal,
			prometheus.CounterValue,
			coreData[hypervisorRootVirtualProcessorHypervisorRunTimePercent].FirstValue,
			coreId, "hypervisor",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorTimeTotal,
			prometheus.CounterValue,
			coreData[hypervisorRootVirtualProcessorGuestIdleTimePercent].FirstValue,
			coreId, "guest_idle",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorTimeTotal,
			prometheus.CounterValue,
			coreData[hypervisorRootVirtualProcessorRemoteRunTimePercent].FirstValue,
			coreId, "remote",
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorTotalRunTimeTotal,
			prometheus.CounterValue,
			coreData[hypervisorRootVirtualProcessorTotalRunTimePercent].FirstValue,
			coreId,
		)

		ch <- prometheus.MustNewConstMetric(
			c.hypervisorRootVirtualProcessorCPUWaitTimePerDispatch,
			prometheus.CounterValue,
			coreData[hypervisorRootVirtualProcessorCPUWaitTimePerDispatch].FirstValue,
			coreId,
		)
	}

	return nil
}
