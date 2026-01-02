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
	"os"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorHost Hyper-V Host metrics
type collectorHost struct {
	perfDataCollectorLogicalProcessor *pdh.Collector
	perfDataObjectLogicalProcessor    []perfDataCounterValuesHost
	perfDataCollectorVirtualProcessor *pdh.Collector
	perfDataObjectVirtualProcessor    []perfDataCounterValuesHost

	hostCPURatio *prometheus.Desc
}

type perfDataCounterValuesHost struct {
	Name string
}

func (c *Collector) buildHost() error {
	var err error

	c.perfDataCollectorLogicalProcessor, err = pdh.NewCollector[perfDataCounterValuesHost](c.logger, pdh.CounterTypeRaw, "Hyper-V Hypervisor Logical Processor", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Hypervisor Logical Processor collector: %w", err)
	}

	c.perfDataCollectorVirtualProcessor, err = pdh.NewCollector[perfDataCounterValuesHost](c.logger, pdh.CounterTypeRaw, "Hyper-V Hypervisor Virtual Processor", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Hypervisor Virtual Processor collector: %w", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	c.hostCPURatio = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "host_cpu_ratio"),
		"Ratio of physical logical CPU cores to virtual cores assigned to all VMs",
		[]string{"host"},
		prometheus.Labels{"host": hostname},
	)

	return nil
}

func (c *Collector) collectHost(ch chan<- prometheus.Metric) error {
	// Collect logical processor count
	err := c.perfDataCollectorLogicalProcessor.Collect(&c.perfDataObjectLogicalProcessor)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Hypervisor Logical Processor metrics: %w", err)
	}

	// Collect virtual processor count
	err = c.perfDataCollectorVirtualProcessor.Collect(&c.perfDataObjectVirtualProcessor)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Hypervisor Virtual Processor metrics: %w", err)
	}

	logicalCoreCount := float64(len(c.perfDataObjectLogicalProcessor))
	virtualCoreCount := float64(len(c.perfDataObjectVirtualProcessor))

	var ratio float64
	if virtualCoreCount > 0 {
		ratio = logicalCoreCount / virtualCoreCount
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	ch <- prometheus.MustNewConstMetric(
		c.hostCPURatio,
		prometheus.GaugeValue,
		ratio,
		hostname,
	)

	return nil
}
