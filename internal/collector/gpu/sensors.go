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
	"log/slog"
	"strconv"

	"github.com/prometheus-community/windows_exporter/internal/headers/gdi32"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/windows"
)

// maxNodes limits the probing of engine nodes. Probing stops at the first node ordinal the driver rejects.
const maxNodes = 64

// sensorMetrics holds the descriptors of the sensor metrics, read from dxgkrnl via D3DKMTQueryAdapterInfo.
type sensorMetrics struct {
	temperature        *prometheus.Desc
	temperatureWarning *prometheus.Desc
	temperatureMax     *prometheus.Desc
	fanSpeed           *prometheus.Desc
	fanSpeedMax        *prometheus.Desc
	powerUsage         *prometheus.Desc
	memoryFrequency    *prometheus.Desc
	memoryFrequencyMax *prometheus.Desc
	memoryBandwidth    *prometheus.Desc
	pcieBandwidth      *prometheus.Desc
	engineFrequency    *prometheus.Desc
	engineFrequencyMax *prometheus.Desc
}

// gpuSensors holds the static adapter data and the sensors the driver reported as supported during discovery.
type gpuSensors struct {
	driverVersion string
	wddmVersion   string
	architecture  string

	physicalAdapters []physicalAdapterSensors
}

type physicalAdapterSensors struct {
	index uint32
	phys  string

	caps gdi32.D3DKMT_ADAPTER_PERFDATACAPS

	// perfData is true if KMTQAITYPE_ADAPTERPERFDATA is supported.
	perfData        bool
	temperature     bool
	fanSpeed        bool
	powerUsage      bool
	memoryFrequency bool
	memoryBandwidth bool
	pcieBandwidth   bool

	// nodes are the engine node ordinals that report a frequency.
	nodes []uint32
}

func (s *gpuSensors) hasDynamicSensors() bool {
	for _, physicalAdapter := range s.physicalAdapters {
		if physicalAdapter.perfData || len(physicalAdapter.nodes) > 0 {
			return true
		}
	}

	return false
}

func newSensorMetrics() sensorMetrics {
	labels := []string{"luid", "device_id", "phys"}

	return sensorMetrics{
		temperature: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "temperature_celsius"),
			"Main temperature sensor reading of the physical GPU in degrees Celsius.",
			labels,
			nil,
		),
		temperatureWarning: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "temperature_warning_celsius"),
			"Temperature in degrees Celsius at which the physical GPU starts throttling.",
			labels,
			nil,
		),
		temperatureMax: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "temperature_max_celsius"),
			"Maximum temperature in degrees Celsius before the physical GPU takes damage.",
			labels,
			nil,
		),
		fanSpeed: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "fan_speed_rpm"),
			"Current speed of the main fan of the physical GPU in revolutions per minute.",
			labels,
			nil,
		),
		fanSpeedMax: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "fan_speed_max_rpm"),
			"Maximum speed of the main fan of the physical GPU in revolutions per minute.",
			labels,
			nil,
		),
		powerUsage: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "power_usage_ratio"),
			"Current power draw of the physical GPU as a ratio of its maximum power (TDP).",
			labels,
			nil,
		),
		memoryFrequency: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "memory_frequency_hertz"),
			"Current clock frequency of the physical GPU memory in hertz.",
			labels,
			nil,
		),
		memoryFrequencyMax: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "memory_frequency_max_hertz"),
			"Maximum clock frequency of the physical GPU memory in hertz, while not overclocked.",
			labels,
			nil,
		),
		memoryBandwidth: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "memory_bandwidth_bytes_total"),
			"Total amount of memory transferred by the physical GPU in bytes.",
			labels,
			nil,
		),
		pcieBandwidth: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "pcie_bandwidth_bytes_total"),
			"Total amount of memory transferred over PCIe by the physical GPU in bytes.",
			labels,
			nil,
		),
		engineFrequency: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "engine_frequency_hertz"),
			"Current clock frequency of the GPU engine in hertz.",
			append(labels, "eng"),
			nil,
		),
		engineFrequencyMax: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, "engine_frequency_max_hertz"),
			"Maximum clock frequency of the GPU engine in hertz, while not overclocked.",
			append(labels, "eng"),
			nil,
		),
	}
}

// discoverSensors reads the static adapter data and probes which sensors the driver supports.
// The queries depend on WDDM 2.4+ and driver support. Failures are logged at debug level and
// mark the sensor as unsupported.
func discoverSensors(logger *slog.Logger, luid windows.LUID) gpuSensors {
	var sensors gpuSensors

	hAdapter, err := gdi32.OpenAdapterFromLUID(luid)
	if err != nil {
		logger.Debug("failed to open GPU adapter for sensor discovery", slog.Any("err", err))

		return sensors
	}

	defer func() {
		if err := gdi32.CloseAdapter(hAdapter); err != nil {
			logger.Debug("failed to close GPU adapter", slog.Any("err", err))
		}
	}()

	if version, err := gdi32.QueryKMDDriverVersion(hAdapter); err != nil {
		logger.Debug("GPU driver version not supported", slog.Any("err", err))
	} else {
		sensors.driverVersion = gdi32.FormatKMDDriverVersion(version)
	}

	if version, err := gdi32.QueryWDDMVersion(hAdapter); err != nil {
		logger.Debug("GPU WDDM version not supported", slog.Any("err", err))
	} else {
		sensors.wddmVersion = gdi32.FormatWDDMVersion(version)
	}

	physicalAdapterCount, err := gdi32.QueryPhysicalAdapterCount(hAdapter)
	if err != nil || physicalAdapterCount == 0 {
		logger.Debug("GPU physical adapter count not supported, assuming 1", slog.Any("err", err))

		physicalAdapterCount = 1
	}

	for index := range physicalAdapterCount {
		physicalAdapter := discoverPhysicalAdapterSensors(logger.With(slog.Uint64("phys", uint64(index))), hAdapter, index)

		if index == 0 {
			if version, err := gdi32.QueryGPUVersion(hAdapter, index); err != nil {
				logger.Debug("GPU version not supported", slog.Any("err", err))
			} else {
				sensors.architecture = windows.UTF16ToString(version.GpuArchitecture[:])
			}
		}

		sensors.physicalAdapters = append(sensors.physicalAdapters, physicalAdapter)
	}

	return sensors
}

func discoverPhysicalAdapterSensors(logger *slog.Logger, hAdapter gdi32.D3DKMT_HANDLE, index uint32) physicalAdapterSensors {
	physicalAdapter := physicalAdapterSensors{
		index: index,
		phys:  strconv.FormatUint(uint64(index), 10),
	}

	caps, err := gdi32.QueryAdapterPerfDataCaps(hAdapter, index)
	if err != nil {
		logger.Debug("GPU adapter perf data caps not supported", slog.Any("err", err))
	} else {
		physicalAdapter.caps = caps
	}

	// Drivers report 0 for unsupported values, so a sensor counts as supported
	// if either the caps or the first reading are non-zero.
	perfData, err := gdi32.QueryAdapterPerfData(hAdapter, index)
	if err != nil {
		logger.Debug("GPU adapter perf data not supported", slog.Any("err", err))
	} else {
		physicalAdapter.perfData = true
		physicalAdapter.temperature = perfData.Temperature > 0 || caps.TemperatureWarning > 0 || caps.TemperatureMax > 0
		physicalAdapter.fanSpeed = perfData.FanRPM > 0 || caps.MaxFanRPM > 0
		physicalAdapter.powerUsage = perfData.Power > 0
		physicalAdapter.memoryFrequency = perfData.MemoryFrequency > 0 || perfData.MaxMemoryFrequency > 0
		physicalAdapter.memoryBandwidth = perfData.MemoryBandwidth > 0 || caps.MaxMemoryBandwidth > 0
		physicalAdapter.pcieBandwidth = perfData.PCIEBandwidth > 0 || caps.MaxPCIEBandwidth > 0
	}

	for node := range uint32(maxNodes) {
		nodePerfData, err := gdi32.QueryNodePerfData(hAdapter, index, node)
		if err != nil {
			// The driver rejects node ordinals past the last node.
			logger.Debug("GPU node perf data not supported",
				slog.Uint64("eng", uint64(node)),
				slog.Any("err", err),
			)

			break
		}

		if nodePerfData.MaxFrequency > 0 {
			physicalAdapter.nodes = append(physicalAdapter.nodes, node)
		}
	}

	return physicalAdapter
}

func (c *Collector) collectGpuSensorMetrics(ch chan<- prometheus.Metric) {
	for luid, device := range c.gpuDeviceCache {
		if !device.sensors.hasDynamicSensors() {
			continue
		}

		logger := c.logger.With(slog.String("luid", luid))

		hAdapter, err := gdi32.OpenAdapterFromLUID(device.gdi32.LUID)
		if err != nil {
			logger.Debug("failed to open GPU adapter", slog.Any("err", err))

			continue
		}

		for _, physicalAdapter := range device.sensors.physicalAdapters {
			c.collectPhysicalAdapterSensorMetrics(ch, logger, hAdapter, luid, device.ID, physicalAdapter)
		}

		if err := gdi32.CloseAdapter(hAdapter); err != nil {
			logger.Debug("failed to close GPU adapter", slog.Any("err", err))
		}
	}
}

func (c *Collector) collectPhysicalAdapterSensorMetrics(
	ch chan<- prometheus.Metric,
	logger *slog.Logger,
	hAdapter gdi32.D3DKMT_HANDLE,
	luid, deviceID string,
	physicalAdapter physicalAdapterSensors,
) {
	labels := []string{luid, deviceID, physicalAdapter.phys}
	caps := physicalAdapter.caps

	if caps.TemperatureWarning > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.sensorMetrics.temperatureWarning,
			prometheus.GaugeValue,
			float64(caps.TemperatureWarning)/10,
			labels...,
		)
	}

	if caps.TemperatureMax > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.sensorMetrics.temperatureMax,
			prometheus.GaugeValue,
			float64(caps.TemperatureMax)/10,
			labels...,
		)
	}

	if caps.MaxFanRPM > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.sensorMetrics.fanSpeedMax,
			prometheus.GaugeValue,
			float64(caps.MaxFanRPM),
			labels...,
		)
	}

	if physicalAdapter.perfData {
		perfData, err := gdi32.QueryAdapterPerfData(hAdapter, physicalAdapter.index)
		if err != nil {
			logger.Debug("failed to query GPU adapter perf data",
				slog.String("phys", physicalAdapter.phys),
				slog.Any("err", err),
			)
		} else {
			c.collectAdapterPerfData(ch, physicalAdapter, perfData, labels)
		}
	}

	for _, node := range physicalAdapter.nodes {
		nodePerfData, err := gdi32.QueryNodePerfData(hAdapter, physicalAdapter.index, node)
		if err != nil {
			logger.Debug("failed to query GPU node perf data",
				slog.String("phys", physicalAdapter.phys),
				slog.Uint64("eng", uint64(node)),
				slog.Any("err", err),
			)

			continue
		}

		eng := strconv.FormatUint(uint64(node), 10)

		ch <- prometheus.MustNewConstMetric(
			c.sensorMetrics.engineFrequency,
			prometheus.GaugeValue,
			float64(nodePerfData.Frequency),
			append(labels, eng)...,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sensorMetrics.engineFrequencyMax,
			prometheus.GaugeValue,
			float64(nodePerfData.MaxFrequency),
			append(labels, eng)...,
		)
	}
}

func (c *Collector) collectAdapterPerfData(
	ch chan<- prometheus.Metric,
	physicalAdapter physicalAdapterSensors,
	perfData gdi32.D3DKMT_ADAPTER_PERFDATA,
	labels []string,
) {
	if physicalAdapter.temperature {
		ch <- prometheus.MustNewConstMetric(
			c.sensorMetrics.temperature,
			prometheus.GaugeValue,
			float64(perfData.Temperature)/10,
			labels...,
		)
	}

	if physicalAdapter.fanSpeed {
		ch <- prometheus.MustNewConstMetric(
			c.sensorMetrics.fanSpeed,
			prometheus.GaugeValue,
			float64(perfData.FanRPM),
			labels...,
		)
	}

	if physicalAdapter.powerUsage {
		ch <- prometheus.MustNewConstMetric(
			c.sensorMetrics.powerUsage,
			prometheus.GaugeValue,
			float64(perfData.Power)/1000,
			labels...,
		)
	}

	if physicalAdapter.memoryFrequency {
		ch <- prometheus.MustNewConstMetric(
			c.sensorMetrics.memoryFrequency,
			prometheus.GaugeValue,
			float64(perfData.MemoryFrequency),
			labels...,
		)

		if perfData.MaxMemoryFrequency > 0 {
			ch <- prometheus.MustNewConstMetric(
				c.sensorMetrics.memoryFrequencyMax,
				prometheus.GaugeValue,
				float64(perfData.MaxMemoryFrequency),
				labels...,
			)
		}
	}

	if physicalAdapter.memoryBandwidth {
		ch <- prometheus.MustNewConstMetric(
			c.sensorMetrics.memoryBandwidth,
			prometheus.CounterValue,
			float64(perfData.MemoryBandwidth),
			labels...,
		)
	}

	if physicalAdapter.pcieBandwidth {
		ch <- prometheus.MustNewConstMetric(
			c.sensorMetrics.pcieBandwidth,
			prometheus.CounterValue,
			float64(perfData.PCIEBandwidth),
			labels...,
		)
	}
}
