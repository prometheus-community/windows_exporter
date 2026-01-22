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

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/headers/cfgmgr32"
	"github.com/prometheus-community/windows_exporter/internal/headers/gdi32"
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

	gpuDeviceCache map[string]gpuDevice

	// GPU Engine
	gpuEnginePerfDataCollector *pdh.Collector
	gpuEnginePerfDataObject    []gpuEnginePerfDataCounterValues

	gpuInfo              *prometheus.Desc
	gpuEngineRunningTime *prometheus.Desc

	gpuSharedSystemMemorySize    *prometheus.Desc
	gpuDedicatedSystemMemorySize *prometheus.Desc
	gpuDedicatedVideoMemorySize  *prometheus.Desc

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

type gpuDevice struct {
	gdi32    gdi32.GPUDevice
	cfgmgr32 cfgmgr32.Device
	ID       string
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

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	var err error

	c.gpuInfo = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "info"),
		"A metric with a constant '1' value labeled with gpu device information.",
		[]string{"luid", "device_id", "name", "bus_number", "phys", "function_number"},
		nil,
	)

	c.gpuSharedSystemMemorySize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "shared_system_memory_size_bytes"),
		"The size, in bytes, of memory from system memory that can be shared by many users.",
		[]string{"luid", "device_id"},
		nil,
	)
	c.gpuDedicatedSystemMemorySize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dedicated_system_memory_size_bytes"),
		"The size, in bytes, of memory that is dedicated from system memory.",
		[]string{"luid", "device_id"},
		nil,
	)
	c.gpuDedicatedVideoMemorySize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dedicated_video_memory_size_bytes"),
		"The size, in bytes, of memory that is dedicated from video memory.",
		[]string{"luid", "device_id"},
		nil,
	)

	c.gpuEngineRunningTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "engine_time_seconds"),
		"Total running time of the GPU in seconds.",
		[]string{"process_id", "luid", "device_id", "phys", "eng", "engtype"},
		nil,
	)

	c.gpuAdapterMemoryDedicatedUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "adapter_memory_dedicated_bytes"),
		"Dedicated GPU memory usage in bytes.",
		[]string{"luid", "device_id", "phys"},
		nil,
	)
	c.gpuAdapterMemorySharedUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "adapter_memory_shared_bytes"),
		"Shared GPU memory usage in bytes.",
		[]string{"luid", "device_id", "phys"},
		nil,
	)
	c.gpuAdapterMemoryTotalCommitted = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "adapter_memory_committed_bytes"),
		"Total committed GPU memory in bytes.",
		[]string{"luid", "device_id", "phys"},
		nil,
	)

	c.gpuLocalAdapterMemoryUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "local_adapter_memory_bytes"),
		"Local adapter memory usage in bytes.",
		[]string{"luid", "device_id", "phys", "part"},
		nil,
	)

	c.gpuNonLocalAdapterMemoryUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "non_local_adapter_memory_bytes"),
		"Non-local adapter memory usage in bytes.",
		[]string{"luid", "device_id", "phys", "part"},
		nil,
	)

	c.gpuProcessMemoryDedicatedUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "process_memory_dedicated_bytes"),
		"Dedicated process memory usage in bytes.",
		[]string{"process_id", "luid", "device_id", "phys"},
		nil,
	)
	c.gpuProcessMemoryLocalUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "process_memory_local_bytes"),
		"Local process memory usage in bytes.",
		[]string{"process_id", "luid", "device_id", "phys"},
		nil,
	)
	c.gpuProcessMemoryNonLocalUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "process_memory_non_local_bytes"),
		"Non-local process memory usage in bytes.",
		[]string{"process_id", "luid", "device_id", "phys"},
		nil,
	)
	c.gpuProcessMemorySharedUsage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "process_memory_shared_bytes"),
		"Shared process memory usage in bytes.",
		[]string{"process_id", "luid", "device_id", "phys"},
		nil,
	)
	c.gpuProcessMemoryTotalCommitted = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "process_memory_committed_bytes"),
		"Total committed process memory in bytes.",
		[]string{"process_id", "luid", "device_id", "phys"},
		nil,
	)

	errs := make([]error, 0)

	c.gpuEnginePerfDataCollector, err = pdh.NewCollector[gpuEnginePerfDataCounterValues](logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "GPU Engine", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create GPU Engine perf data collector: %w", err))
	}

	c.gpuAdapterMemoryPerfDataCollector, err = pdh.NewCollector[gpuAdapterMemoryPerfDataCounterValues](logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "GPU Adapter Memory", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create GPU Adapter Memory perf data collector: %w", err))
	}

	c.gpuLocalAdapterMemoryPerfDataCollector, err = pdh.NewCollector[gpuLocalAdapterMemoryPerfDataCounterValues](logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "GPU Local Adapter Memory", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create GPU Local Adapter Memory perf data collector: %w", err))
	}

	c.gpuNonLocalAdapterMemoryPerfDataCollector, err = pdh.NewCollector[gpuNonLocalAdapterMemoryPerfDataCounterValues](logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "GPU Non Local Adapter Memory", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create GPU Non Local Adapter Memory perf data collector: %w", err))
	}

	c.gpuProcessMemoryPerfDataCollector, err = pdh.NewCollector[gpuProcessMemoryPerfDataCounterValues](logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "GPU Process Memory", pdh.InstancesAll)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to create GPU Process Memory perf data collector: %w", err))
	}

	gpus, err := gdi32.GetGPUDevices()
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to get GPU devices: %w", err))
	}

	for _, gpu := range gpus {
		if gpu.AdapterString == "" {
			continue
		}

		// Skip Microsoft Basic Render Driver
		// https://devicehunt.com/view/type/pci/vendor/1414/device/008C
		if gpu.DeviceID == `PCI\VEN_1414&DEV_008C&SUBSYS_00000000&REV_00` {
			continue
		}

		if c.gpuDeviceCache == nil {
			c.gpuDeviceCache = make(map[string]gpuDevice)
		}

		luidKey := fmt.Sprintf("0x%08X_0x%08X", gpu.LUID.HighPart, gpu.LUID.LowPart)

		deviceID := gpu.DeviceID

		cfgmgr32Devs, err := cfgmgr32.GetDevicesInstanceIDs(gpu.DeviceID)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get device instance IDs for device ID %s: %w", gpu.DeviceID, err))
		}

		var cfgmgr32Dev cfgmgr32.Device

		for _, dev := range cfgmgr32Devs {
			if dev.BusNumber == gpu.BusNumber && dev.DeviceNumber == gpu.DeviceNumber && dev.FunctionNumber == gpu.FunctionNumber {
				cfgmgr32Dev = dev

				break
			}
		}

		if cfgmgr32Dev.InstanceID == "" {
			errs = append(errs, fmt.Errorf("failed to find matching device for device ID %s", gpu.DeviceID))
		} else {
			deviceID = cfgmgr32Dev.InstanceID
		}

		c.gpuDeviceCache[luidKey] = gpuDevice{
			gdi32:    gpu,
			cfgmgr32: cfgmgr32Dev,
			ID:       deviceID,
		}

		logger.Debug("Found GPU device",
			slog.String("collector", Name),
			slog.String("name", gpu.AdapterString),
			slog.String("luid", luidKey),
			slog.String("device_id", deviceID),
			slog.String("name", gpu.AdapterString),
			slog.Uint64("bus_number", uint64(gpu.BusNumber)),
			slog.Uint64("device_number", uint64(gpu.DeviceNumber)),
			slog.Uint64("function_number", uint64(gpu.FunctionNumber)),
		)
	}

	return errors.Join(errs...)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0)

	c.collectGpuInfo(ch)

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

func (c *Collector) collectGpuInfo(ch chan<- prometheus.Metric) {
	for luid, gpu := range c.gpuDeviceCache {
		ch <- prometheus.MustNewConstMetric(
			c.gpuInfo,
			prometheus.GaugeValue,
			1.0,
			luid,
			gpu.ID,
			gpu.gdi32.AdapterString,
			gpu.gdi32.BusNumber.String(),
			gpu.gdi32.DeviceNumber.String(),
			gpu.gdi32.FunctionNumber.String(),
		)

		ch <- prometheus.MustNewConstMetric(
			c.gpuSharedSystemMemorySize,
			prometheus.GaugeValue,
			float64(gpu.gdi32.SharedSystemMemorySize),
			luid, gpu.ID,
		)

		ch <- prometheus.MustNewConstMetric(
			c.gpuDedicatedSystemMemorySize,
			prometheus.GaugeValue,
			float64(gpu.gdi32.DedicatedSystemMemorySize),
			luid, gpu.ID,
		)

		ch <- prometheus.MustNewConstMetric(
			c.gpuDedicatedVideoMemorySize,
			prometheus.GaugeValue,
			float64(gpu.gdi32.DedicatedVideoMemorySize),
			luid, gpu.ID,
		)
	}
}

func (c *Collector) collectGpuEngineMetrics(ch chan<- prometheus.Metric) error {
	// Collect the GPU Engine perf data.
	if err := c.gpuEnginePerfDataCollector.Collect(&c.gpuEnginePerfDataObject); err != nil {
		return fmt.Errorf("failed to collect GPU Engine perf data: %w", err)
	}

	// Iterate over the GPU Engine perf data and aggregate the values.
	for _, data := range c.gpuEnginePerfDataObject {
		instance := parseGPUCounterInstanceString(data.Name)

		device, ok := c.gpuDeviceCache[instance.Luid]
		if !ok {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.gpuEngineRunningTime,
			prometheus.CounterValue,
			data.RunningTime/10_000_000,
			instance.Pid, instance.Luid, device.ID, instance.Phys, instance.Eng, instance.Engtype,
		)
	}

	return nil
}

func (c *Collector) collectGpuAdapterMemoryMetrics(ch chan<- prometheus.Metric) error {
	// Collect the GPU Adapter Memory perf data.
	if err := c.gpuAdapterMemoryPerfDataCollector.Collect(&c.gpuAdapterMemoryPerfDataObject); err != nil {
		return fmt.Errorf("failed to collect GPU Adapter Memory perf data: %w", err)
	}

	for _, data := range c.gpuAdapterMemoryPerfDataObject {
		instance := parseGPUCounterInstanceString(data.Name)

		device, ok := c.gpuDeviceCache[instance.Luid]
		if !ok {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.gpuAdapterMemoryDedicatedUsage,
			prometheus.GaugeValue,
			data.DedicatedUsage,
			instance.Luid, device.ID, instance.Phys,
		)

		ch <- prometheus.MustNewConstMetric(
			c.gpuAdapterMemorySharedUsage,
			prometheus.GaugeValue,
			data.SharedUsage,
			instance.Luid, device.ID, instance.Phys,
		)

		ch <- prometheus.MustNewConstMetric(
			c.gpuAdapterMemoryTotalCommitted,
			prometheus.GaugeValue,
			data.TotalCommitted,
			instance.Luid, device.ID, instance.Phys,
		)
	}

	return nil
}

func (c *Collector) collectGpuLocalAdapterMemoryMetrics(ch chan<- prometheus.Metric) error {
	// Collect the GPU Local Adapter Memory perf data.
	if err := c.gpuLocalAdapterMemoryPerfDataCollector.Collect(&c.gpuLocalAdapterMemoryPerfDataObject); err != nil {
		return fmt.Errorf("failed to collect GPU Local Adapter Memory perf data: %w", err)
	}

	for _, data := range c.gpuLocalAdapterMemoryPerfDataObject {
		instance := parseGPUCounterInstanceString(data.Name)

		device, ok := c.gpuDeviceCache[instance.Luid]
		if !ok {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.gpuLocalAdapterMemoryUsage,
			prometheus.GaugeValue,
			data.LocalUsage,
			instance.Luid, device.ID, instance.Phys, instance.Part,
		)
	}

	return nil
}

func (c *Collector) collectGpuNonLocalAdapterMemoryMetrics(ch chan<- prometheus.Metric) error {
	// Collect the GPU Non Local Adapter Memory perf data.
	if err := c.gpuNonLocalAdapterMemoryPerfDataCollector.Collect(&c.gpuNonLocalAdapterMemoryPerfDataObject); err != nil {
		return fmt.Errorf("failed to collect GPU Non Local Adapter Memory perf data: %w", err)
	}

	for _, data := range c.gpuNonLocalAdapterMemoryPerfDataObject {
		instance := parseGPUCounterInstanceString(data.Name)

		device, ok := c.gpuDeviceCache[instance.Luid]
		if !ok {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.gpuNonLocalAdapterMemoryUsage,
			prometheus.GaugeValue,
			data.NonLocalUsage,
			instance.Luid, device.ID, instance.Phys, instance.Part,
		)
	}

	return nil
}

func (c *Collector) collectGpuProcessMemoryMetrics(ch chan<- prometheus.Metric) error {
	// Collect the GPU Process Memory perf data.
	if err := c.gpuProcessMemoryPerfDataCollector.Collect(&c.gpuProcessMemoryPerfDataObject); err != nil {
		return fmt.Errorf("failed to collect GPU Process Memory perf data: %w", err)
	}

	for _, data := range c.gpuProcessMemoryPerfDataObject {
		instance := parseGPUCounterInstanceString(data.Name)

		device, ok := c.gpuDeviceCache[instance.Luid]
		if !ok {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.gpuProcessMemoryDedicatedUsage,
			prometheus.GaugeValue,
			data.DedicatedUsage,
			instance.Pid, instance.Luid, device.ID, instance.Phys,
		)

		ch <- prometheus.MustNewConstMetric(
			c.gpuProcessMemoryLocalUsage,
			prometheus.GaugeValue,
			data.LocalUsage,
			instance.Pid, instance.Luid, device.ID, instance.Phys,
		)

		ch <- prometheus.MustNewConstMetric(
			c.gpuProcessMemoryNonLocalUsage,
			prometheus.GaugeValue,
			data.NonLocalUsage,
			instance.Pid, instance.Luid, device.ID, instance.Phys,
		)

		ch <- prometheus.MustNewConstMetric(
			c.gpuProcessMemorySharedUsage,
			prometheus.GaugeValue,
			data.SharedUsage,
			instance.Pid, instance.Luid, device.ID, instance.Phys,
		)

		ch <- prometheus.MustNewConstMetric(
			c.gpuProcessMemoryTotalCommitted,
			prometheus.GaugeValue,
			data.TotalCommitted,
			instance.Pid, instance.Luid, device.ID, instance.Phys,
		)
	}

	return nil
}
