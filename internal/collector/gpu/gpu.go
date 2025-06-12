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

package gpu

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/setupapi"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "gpu"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

type Collector struct {
	config Config

	// GPU Engine
	gpuEnginePerfDataCollector *pdh.Collector
	gpuEnginePerfDataObject    []gpuEnginePerfDataCounterValues

	gpuInfo              *prometheus.Desc
	gpuEngineRunningTime *prometheus.Desc

	// GPU Adapter Memory
	gpuAdapterMemoryPerfDataCollector *pdh.Collector
	gpuAdapterMemoryPerfDataObject    []gpuAdapterMemoryPerfDataCounterValues

	gpuAdapterMemoryDedicatedUsage *prometheus.Desc
	gpuAdapterMemorySharedUsage    *prometheus.Desc
	gpuAdapterMemoryTotalCommitted *prometheus.Desc

	// GPU Local Adapter Memory
	gpuLocalAdapterMemoryPerfDataCollector *pdh.Collector
	gpuLocalAdapterMemoryPerfDataObject    []gpuLocalAdapterMemoryPerfDataCounterValues

	gpuLocalAdapterMemoryUsage *prometheus.Desc

	// GPU Non Local Adapter Memory
	gpuNonLocalAdapterMemoryPerfDataCollector *pdh.Collector
	gpuNonLocalAdapterMemoryPerfDataObject    []gpuNonLocalAdapterMemoryPerfDataCounterValues

	gpuNonLocalAdapterMemoryUsage *prometheus.Desc

	// GPU Process Memory
	gpuProcessMemoryPerfDataCollector *pdh.Collector
	gpuProcessMemoryPerfDataObject    []gpuProcessMemoryPerfDataCounterValues

	gpuProcessMemoryDedicatedUsage *prometheus.Desc
	gpuProcessMemoryLocalUsage     *prometheus.Desc
	gpuProcessMemoryNonLocalUsage  *prometheus.Desc
	gpuProcessMemorySharedUsage    *prometheus.Desc
	gpuProcessMemoryTotalCommitted *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	c.gpuEnginePerfDataCollector.Close()
	c.gpuAdapterMemoryPerfDataCollector.Close()
	c.gpuLocalAdapterMemoryPerfDataCollector.Close()
	c.gpuNonLocalAdapterMemoryPerfDataCollector.Close()
	c.gpuProcessMemoryPerfDataCollector.Close()

	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	var err error

	c.gpuInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"A metric with a constant '1' value labeled with gpu device information.",
		[]string{"phys", "physical_device_object_name", "hardware_id", "friendly_name", "description"},
		nil,
	)

	c.gpuEngineRunningTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "engine_time_seconds"),
		"Total running time of the GPU in seconds.",
		[]string{"process_id", "phys", "eng", "engtype"},
		nil,
	)

	c.gpuAdapterMemoryDedicatedUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "adapter_memory_dedicated_bytes"),
		"Dedicated GPU memory usage in bytes.",
		[]string{"phys"},
		nil,
	)
	c.gpuAdapterMemorySharedUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "adapter_memory_shared_bytes"),
		"Shared GPU memory usage in bytes.",
		[]string{"phys"},
		nil,
	)
	c.gpuAdapterMemoryTotalCommitted = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "adapter_memory_committed_bytes"),
		"Total committed GPU memory in bytes.",
		[]string{"phys"},
		nil,
	)

	c.gpuLocalAdapterMemoryUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "local_adapter_memory_bytes"),
		"Local adapter memory usage in bytes.",
		[]string{"phys"},
		nil,
	)

	c.gpuNonLocalAdapterMemoryUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "non_local_adapter_memory_bytes"),
		"Non-local adapter memory usage in bytes.",
		[]string{"phys"},
		nil,
	)

	c.gpuProcessMemoryDedicatedUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "process_memory_dedicated_bytes"),
		"Dedicated process memory usage in bytes.",
		[]string{"process_id", "phys"},
		nil,
	)
	c.gpuProcessMemoryLocalUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "process_memory_local_bytes"),
		"Local process memory usage in bytes.",
		[]string{"process_id", "phys"},
		nil,
	)
	c.gpuProcessMemoryNonLocalUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "process_memory_non_local_bytes"),
		"Non-local process memory usage in bytes.",
		[]string{"process_id", "phys"},
		nil,
	)
	c.gpuProcessMemorySharedUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "process_memory_shared_bytes"),
		"Shared process memory usage in bytes.",
		[]string{"process_id", "phys"},
		nil,
	)
	c.gpuProcessMemoryTotalCommitted = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "process_memory_committed_bytes"),
		"Total committed process memory in bytes.",
		[]string{"process_id", "phys"},
		nil,
	)

	errs := make([]error, 0)

	c.gpuEnginePerfDataCollector, err = pdh.NewCollector[gpuEnginePerfDataCounterValues](pdh.CounterTypeRaw, "GPU Engine", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create GPU Engine perf data collector: %w", err))
	}

	c.gpuAdapterMemoryPerfDataCollector, err = pdh.NewCollector[gpuAdapterMemoryPerfDataCounterValues](pdh.CounterTypeRaw, "GPU Adapter Memory", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create GPU Adapter Memory perf data collector: %w", err))
	}

	c.gpuLocalAdapterMemoryPerfDataCollector, err = pdh.NewCollector[gpuLocalAdapterMemoryPerfDataCounterValues](pdh.CounterTypeRaw, "GPU Local Adapter Memory", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create GPU Local Adapter Memory perf data collector: %w", err))
	}

	c.gpuNonLocalAdapterMemoryPerfDataCollector, err = pdh.NewCollector[gpuNonLocalAdapterMemoryPerfDataCounterValues](pdh.CounterTypeRaw, "GPU Non Local Adapter Memory", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create GPU Non Local Adapter Memory perf data collector: %w", err))
	}

	c.gpuProcessMemoryPerfDataCollector, err = pdh.NewCollector[gpuProcessMemoryPerfDataCounterValues](pdh.CounterTypeRaw, "GPU Process Memory", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create GPU Process Memory perf data collector: %w", err))
	}

	return errors.Join(errs...)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0)

	if err := c.collectGpuInfo(ch); err != nil {
		errs = append(errs, err)
	}

	if err := c.collectGpuEngineMetrics(ch); err != nil {
		errs = append(errs, err)
	}

	if err := c.collectGpuAdapterMemoryMetrics(ch); err != nil {
		errs = append(errs, err)
	}

	if err := c.collectGpuLocalAdapterMemoryMetrics(ch); err != nil {
		errs = append(errs, err)
	}

	if err := c.collectGpuNonLocalAdapterMemoryMetrics(ch); err != nil {
		errs = append(errs, err)
	}

	if err := c.collectGpuProcessMemoryMetrics(ch); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (c *Collector) collectGpuInfo(ch chan<- prometheus.Metric) error {
	gpus, err := setupapi.GetGPUDevices()
	if err != nil {
		return fmt.Errorf("failed to get GPU devices: %w", err)
	}

	for i, gpu := range gpus {
		ch <- prometheus.MustNewConstMetric(
			c.gpuInfo,
			prometheus.GaugeValue,
			1.0,
			strconv.Itoa(i),
			gpu.PhysicalDeviceObjectName,
			gpu.HardwareID,
			gpu.FriendlyName,
			gpu.DeviceDesc,
		)
	}

	return nil
}

func (c *Collector) collectGpuEngineMetrics(ch chan<- prometheus.Metric) error {
	// Collect the GPU Engine perf data.
	if err := c.gpuEnginePerfDataCollector.Collect(&c.gpuEnginePerfDataObject); err != nil {
		return fmt.Errorf("failed to collect GPU Engine perf data: %w", err)
	}

	runningTimeMap := make(map[PidPhysEngEngType]float64)
	// Iterate over the GPU Engine perf data and aggregate the values.
	for _, data := range c.gpuEnginePerfDataObject {
		instance := parseGPUCounterInstanceString(data.Name)

		key := PidPhysEngEngType{
			Pid:     instance.Pid,
			Phys:    instance.Phys,
			Eng:     instance.Eng,
			Engtype: instance.Engtype,
		}
		runningTimeMap[key] += data.RunningTime / 10_000_000 // RunningTime is in 100ns units, convert to seconds.
	}

	for key, runningTime := range runningTimeMap {
		ch <- prometheus.MustNewConstMetric(
			c.gpuEngineRunningTime,
			prometheus.CounterValue,
			runningTime,
			key.Pid, key.Phys, key.Eng, key.Engtype,
		)
	}

	return nil
}

func (c *Collector) collectGpuAdapterMemoryMetrics(ch chan<- prometheus.Metric) error {
	// Collect the GPU Adapter Memory perf data.
	if err := c.gpuAdapterMemoryPerfDataCollector.Collect(&c.gpuAdapterMemoryPerfDataObject); err != nil {
		return fmt.Errorf("failed to collect GPU Adapter Memory perf data: %w", err)
	}

	dedicatedUsageMap := make(map[PidPhysEngEngType]float64)
	sharedUsageMap := make(map[PidPhysEngEngType]float64)
	totalCommittedMap := make(map[PidPhysEngEngType]float64)

	for _, data := range c.gpuAdapterMemoryPerfDataObject {
		instance := parseGPUCounterInstanceString(data.Name)

		key := PidPhysEngEngType{
			Pid:     instance.Pid,
			Phys:    instance.Phys,
			Eng:     instance.Eng,
			Engtype: instance.Engtype,
		}
		dedicatedUsageMap[key] += data.DedicatedUsage
		sharedUsageMap[key] += data.SharedUsage
		totalCommittedMap[key] += data.TotalCommitted
	}

	for key, dedicatedUsage := range dedicatedUsageMap {
		ch <- prometheus.MustNewConstMetric(
			c.gpuAdapterMemoryDedicatedUsage,
			prometheus.GaugeValue,
			dedicatedUsage,
			key.Phys,
		)
		ch <- prometheus.MustNewConstMetric(
			c.gpuAdapterMemorySharedUsage,
			prometheus.GaugeValue,
			sharedUsageMap[key],
			key.Phys,
		)
		ch <- prometheus.MustNewConstMetric(
			c.gpuAdapterMemoryTotalCommitted,
			prometheus.GaugeValue,
			totalCommittedMap[key],
			key.Phys,
		)
	}

	return nil
}

func (c *Collector) collectGpuLocalAdapterMemoryMetrics(ch chan<- prometheus.Metric) error {
	// Collect the GPU Local Adapter Memory perf data.
	if err := c.gpuLocalAdapterMemoryPerfDataCollector.Collect(&c.gpuLocalAdapterMemoryPerfDataObject); err != nil {
		return fmt.Errorf("failed to collect GPU Local Adapter Memory perf data: %w", err)
	}

	localAdapterMemoryMap := make(map[string]float64)

	for _, data := range c.gpuLocalAdapterMemoryPerfDataObject {
		instance := parseGPUCounterInstanceString(data.Name)

		localAdapterMemoryMap[instance.Phys] += data.LocalUsage
	}

	for phys, localUsage := range localAdapterMemoryMap {
		ch <- prometheus.MustNewConstMetric(
			c.gpuLocalAdapterMemoryUsage,
			prometheus.GaugeValue,
			localUsage,
			phys,
		)
	}

	return nil
}

func (c *Collector) collectGpuNonLocalAdapterMemoryMetrics(ch chan<- prometheus.Metric) error {
	// Collect the GPU Non Local Adapter Memory perf data.
	if err := c.gpuNonLocalAdapterMemoryPerfDataCollector.Collect(&c.gpuNonLocalAdapterMemoryPerfDataObject); err != nil {
		return fmt.Errorf("failed to collect GPU Non Local Adapter Memory perf data: %w", err)
	}

	nonLocalAdapterMemoryMap := make(map[string]float64)

	for _, data := range c.gpuNonLocalAdapterMemoryPerfDataObject {
		instance := parseGPUCounterInstanceString(data.Name)

		nonLocalAdapterMemoryMap[instance.Phys] += data.NonLocalUsage
	}

	for phys, nonLocalUsage := range nonLocalAdapterMemoryMap {
		ch <- prometheus.MustNewConstMetric(
			c.gpuNonLocalAdapterMemoryUsage,
			prometheus.GaugeValue,
			nonLocalUsage,
			phys,
		)
	}

	return nil
}

func (c *Collector) collectGpuProcessMemoryMetrics(ch chan<- prometheus.Metric) error {
	// Collect the GPU Process Memory perf data.
	if err := c.gpuProcessMemoryPerfDataCollector.Collect(&c.gpuProcessMemoryPerfDataObject); err != nil {
		return fmt.Errorf("failed to collect GPU Process Memory perf data: %w", err)
	}

	processDedicatedUsageMap := make(map[PidPhys]float64)
	processLocalUsageMap := make(map[PidPhys]float64)
	processNonLocalUsageMap := make(map[PidPhys]float64)
	processSharedUsageMap := make(map[PidPhys]float64)
	processTotalCommittedMap := make(map[PidPhys]float64)

	for _, data := range c.gpuProcessMemoryPerfDataObject {
		instance := parseGPUCounterInstanceString(data.Name)

		key := PidPhys{
			Pid:  instance.Pid,
			Phys: instance.Phys,
		}
		processDedicatedUsageMap[key] += data.DedicatedUsage
		processLocalUsageMap[key] += data.LocalUsage
		processNonLocalUsageMap[key] += data.NonLocalUsage
		processSharedUsageMap[key] += data.SharedUsage
		processTotalCommittedMap[key] += data.TotalCommitted
	}

	for key, dedicatedUsage := range processDedicatedUsageMap {
		ch <- prometheus.MustNewConstMetric(
			c.gpuProcessMemoryDedicatedUsage,
			prometheus.GaugeValue,
			dedicatedUsage,
			key.Pid, key.Phys,
		)
		ch <- prometheus.MustNewConstMetric(
			c.gpuProcessMemoryLocalUsage,
			prometheus.GaugeValue,
			processLocalUsageMap[key],
			key.Pid, key.Phys,
		)
		ch <- prometheus.MustNewConstMetric(
			c.gpuProcessMemoryNonLocalUsage,
			prometheus.GaugeValue,
			processNonLocalUsageMap[key],
			key.Pid, key.Phys,
		)
		ch <- prometheus.MustNewConstMetric(
			c.gpuProcessMemorySharedUsage,
			prometheus.GaugeValue,
			processSharedUsageMap[key],
			key.Pid, key.Phys,
		)
		ch <- prometheus.MustNewConstMetric(
			c.gpuProcessMemoryTotalCommitted,
			prometheus.GaugeValue,
			processTotalCommittedMap[key],
			key.Pid, key.Phys,
		)
	}

	return nil
}
