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

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorVirtualMachineHealthSummary Hyper-V Virtual Machine Health Summary metrics
type collectorVirtualMachineHealthSummary struct {
	perfDataCollectorVirtualMachineHealthSummary *pdh.Collector
	perfDataObjectVirtualMachineHealthSummary    []perfDataCounterValuesVirtualMachineHealthSummary

	// \Hyper-V Virtual Machine Health Summary\Health Critical
	// \Hyper-V Virtual Machine Health Summary\Health Ok
	health *prometheus.Desc
}

type perfDataCounterValuesVirtualMachineHealthSummary struct {
	// Hyper-V Virtual Machine Health Summary
	HealthCritical float64 `perfdata:"Health Critical"`
	HealthOk       float64 `perfdata:"Health Ok"`
}

func (c *Collector) buildVirtualMachineHealthSummary() error {
	var err error

	c.perfDataCollectorVirtualMachineHealthSummary, err = pdh.NewCollector[perfDataCounterValuesVirtualMachineHealthSummary](pdh.CounterTypeRaw, "Hyper-V Virtual Machine Health Summary", nil)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Virtual Machine Health Summary collector: %w", err)
	}

	c.health = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_machine_health_total_count"),
		"Represents the number of virtual machines with critical health",
		[]string{"state"},
		nil,
	)

	return nil
}

func (c *Collector) collectVirtualMachineHealthSummary(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorVirtualMachineHealthSummary.Collect(&c.perfDataObjectVirtualMachineHealthSummary)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Virtual Machine Health Summary metrics: %w", err)
	}

	ch <- prometheus.MustNewConstMetric(
		c.health,
		prometheus.GaugeValue,
		c.perfDataObjectVirtualMachineHealthSummary[0].HealthCritical,
		"critical",
	)

	ch <- prometheus.MustNewConstMetric(
		c.health,
		prometheus.GaugeValue,
		c.perfDataObjectVirtualMachineHealthSummary[0].HealthOk,
		"ok",
	)

	return nil
}
