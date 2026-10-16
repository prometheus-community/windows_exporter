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

package cpu

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "cpu"

// maxTimeDrop is the largest decrease of a processor time counter, in
// seconds, that Collect hides. Windows sometimes reports a per-processor time
// one clock tick (1/64 s) lower than in the previous sample. A larger drop is
// passed through as a counter reset.
const maxTimeDrop = 1.0

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

type Collector struct {
	config Config

	perfDataCollector *pdh.Collector[perfDataCounterValues]
	perfDataObject    []perfDataCounterValues

	mu sync.Mutex

	processorRTCValues   map[string]utils.Counter
	processorMPerfValues map[string]utils.Counter
	// timeValues are the processor times of the last scrape by core and mode.
	timeValues map[timeKey]float64

	logicalProcessors          *prometheus.Desc
	cStateSecondsTotal         *prometheus.Desc
	timeTotal                  *prometheus.Desc
	interruptsTotal            *prometheus.Desc
	dpcsTotal                  *prometheus.Desc
	clockInterruptsTotal       *prometheus.Desc
	idleBreakEventsTotal       *prometheus.Desc
	parkingStatus              *prometheus.Desc
	processorFrequencyHertz    *prometheus.Desc
	processorPerformance       *prometheus.Desc
	processorMPerf             *prometheus.Desc
	processorRTC               *prometheus.Desc
	processorUtility           *prometheus.Desc
	processorPrivilegedUtility *prometheus.Desc
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
	c.perfDataCollector.Close()

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.mu = sync.Mutex{}

	c.buildDescriptors()

	var err error

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "Processor Information", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Processor Information collector: %w", err)
	}

	return nil
}

// buildDescriptors creates the descriptors and the per-core state.
func (c *Collector) buildDescriptors() {
	c.logicalProcessors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "logical_processor"),
		"Total number of logical processors",
		nil,
		nil,
	)
	c.cStateSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cstate_seconds_total"),
		"Time spent in low-power idle state",
		[]string{"core", "state"},
		nil,
	)
	c.timeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "time_seconds_total"),
		"Time that processor spent in different modes (dpc, idle, interrupt, privileged, user)",
		[]string{"core", "mode"},
		nil,
	)
	c.interruptsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "interrupts_total"),
		"Total number of received and serviced hardware interrupts",
		[]string{"core"},
		nil,
	)
	c.dpcsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dpcs_total"),
		"Total number of received and serviced deferred procedure calls (DPCs)",
		[]string{"core"},
		nil,
	)
	c.clockInterruptsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clock_interrupts_total"),
		"Total number of received and serviced clock tick interrupts",
		[]string{"core"},
		nil,
	)
	c.idleBreakEventsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "idle_break_events_total"),
		"Total number of time processor was woken from idle",
		[]string{"core"},
		nil,
	)
	c.parkingStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "parking_status"),
		"Parking Status represents whether a processor is parked or not",
		[]string{"core"},
		nil,
	)
	c.processorFrequencyHertz = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "core_frequency_hertz"),
		"Core frequency in hertz",
		[]string{"core"},
		nil,
	)
	c.processorPerformance = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processor_performance_total"),
		"Processor Performance is the average performance of the processor while it is executing instructions, as a percentage of the nominal performance of the processor. On some processors, Processor Performance may exceed 100%",
		[]string{"core"},
		nil,
	)
	c.processorMPerf = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processor_mperf_total"),
		"Processor MPerf is the number of TSC ticks incremented while executing instructions",
		[]string{"core"},
		nil,
	)
	c.processorRTC = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processor_rtc_total"),
		"Processor RTC represents the number of RTC ticks made since the system booted. It should consistently be 64e6, and can be used to properly derive Processor Utility Rate",
		[]string{"core"},
		nil,
	)
	c.processorUtility = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processor_utility_total"),
		"Processor Utility represents is the amount of time the core spends executing instructions",
		[]string{"core"},
		nil,
	)
	c.processorPrivilegedUtility = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processor_privileged_utility_total"),
		"Processor Privileged Utility represents is the amount of time the core has spent executing instructions inside the kernel",
		[]string{"core"},
		nil,
	)

	c.processorRTCValues = map[string]utils.Counter{}
	c.processorMPerfValues = map[string]utils.Counter{}
	c.timeValues = map[timeKey]float64{}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	c.mu.Lock() // Lock is needed to prevent concurrent access to the per-core state maps.
	defer c.mu.Unlock()

	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect Processor Information metrics: %w", err)
	}

	c.publish(ch)

	return nil
}

// publish sends the metrics of c.perfDataObject and updates the per-core
// state. The caller holds c.mu.
func (c *Collector) publish(ch chan<- prometheus.Metric) {
	var coreCount float64

	timeValues := make(map[timeKey]float64, len(c.timeValues))

	for _, coreData := range c.perfDataObject {
		core := coreData.Name
		coreCount++

		var (
			counterProcessorRTCValues   utils.Counter
			counterProcessorMPerfValues utils.Counter
			ok                          bool
		)

		if counterProcessorRTCValues, ok = c.processorRTCValues[core]; ok {
			counterProcessorRTCValues.AddValue(uint32(coreData.ProcessorUtilityRateSecondValue))
		} else {
			counterProcessorRTCValues = utils.NewCounter(uint32(coreData.ProcessorUtilityRateSecondValue))
		}

		c.processorRTCValues[core] = counterProcessorRTCValues

		if counterProcessorMPerfValues, ok = c.processorMPerfValues[core]; ok {
			counterProcessorMPerfValues.AddValue(uint32(coreData.ProcessorPerformanceSecondValue))
		} else {
			counterProcessorMPerfValues = utils.NewCounter(uint32(coreData.ProcessorPerformanceSecondValue))
		}

		c.processorMPerfValues[core] = counterProcessorMPerfValues

		ch <- prometheus.MustNewConstMetric(
			c.cStateSecondsTotal,
			prometheus.CounterValue,
			c.monotonicTime(timeValues, core, "c1", coreData.C1TimeSeconds),
			core, "c1",
		)

		ch <- prometheus.MustNewConstMetric(
			c.cStateSecondsTotal,
			prometheus.CounterValue,
			c.monotonicTime(timeValues, core, "c2", coreData.C2TimeSeconds),
			core, "c2",
		)

		ch <- prometheus.MustNewConstMetric(
			c.cStateSecondsTotal,
			prometheus.CounterValue,
			c.monotonicTime(timeValues, core, "c3", coreData.C3TimeSeconds),
			core, "c3",
		)

		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			c.monotonicTime(timeValues, core, "idle", coreData.IdleTimeSeconds),
			core, "idle",
		)

		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			c.monotonicTime(timeValues, core, "interrupt", coreData.InterruptTimeSeconds),
			core, "interrupt",
		)

		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			c.monotonicTime(timeValues, core, "dpc", coreData.DpcTimeSeconds),
			core, "dpc",
		)

		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			c.monotonicTime(timeValues, core, "privileged", coreData.PrivilegedTimeSeconds),
			core, "privileged",
		)

		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			c.monotonicTime(timeValues, core, "user", coreData.UserTimeSeconds),
			core, "user",
		)

		ch <- prometheus.MustNewConstMetric(
			c.interruptsTotal,
			prometheus.CounterValue,
			coreData.InterruptsTotal,
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dpcsTotal,
			prometheus.CounterValue,
			coreData.DpcQueuedPerSecond,
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.clockInterruptsTotal,
			prometheus.CounterValue,
			coreData.ClockInterruptsTotal,
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.idleBreakEventsTotal,
			prometheus.CounterValue,
			coreData.IdleBreakEventsTotal,
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.parkingStatus,
			prometheus.GaugeValue,
			coreData.ParkingStatus,
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.processorFrequencyHertz,
			prometheus.GaugeValue,
			coreData.ProcessorFrequencyMHz*1e6,
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.processorPerformance,
			prometheus.CounterValue,
			coreData.ProcessorPerformance,
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.processorMPerf,
			prometheus.CounterValue,
			counterProcessorMPerfValues.Value(),
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.processorRTC,
			prometheus.CounterValue,
			counterProcessorRTCValues.Value(),
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.processorUtility,
			prometheus.CounterValue,
			coreData.ProcessorUtilityRate,
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.processorPrivilegedUtility,
			prometheus.CounterValue,
			coreData.PrivilegedUtilitySeconds,
			core,
		)
	}

	c.timeValues = timeValues

	ch <- prometheus.MustNewConstMetric(
		c.logicalProcessors,
		prometheus.GaugeValue,
		coreCount,
	)
}

type timeKey struct {
	core, mode string
}

// monotonicTime returns the processor time of the core and mode, but at least
// the value of the last scrape if it dropped by less than maxTimeDrop. Rate
// functions read every decrease of a counter as a reset, so a drop of one
// clock tick would show up as a spike. The returned value is stored in
// timeValues.
func (c *Collector) monotonicTime(timeValues map[timeKey]float64, core, mode string, value float64) float64 {
	key := timeKey{core: core, mode: mode}

	if last, ok := c.timeValues[key]; ok && value < last && last-value < maxTimeDrop {
		value = last
	}

	timeValues[key] = value

	return value
}
