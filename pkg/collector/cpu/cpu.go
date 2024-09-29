//go:build windows

package cpu

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/pkg/perfdata"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
)

const Name = "cpu"

type Config struct{}

var ConfigDefaults = Config{}

type Collector struct {
	config Config

	perfDataCollector *perfdata.Collector

	logicalProcessors          *prometheus.Desc
	cStateSecondsTotal         *prometheus.Desc
	timeTotal                  *prometheus.Desc
	interruptsTotal            *prometheus.Desc
	dpcsTotal                  *prometheus.Desc
	clockInterruptsTotal       *prometheus.Desc
	idleBreakEventsTotal       *prometheus.Desc
	parkingStatus              *prometheus.Desc
	processorFrequencyMHz      *prometheus.Desc
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

func (c *Collector) GetPerfCounter(_ *slog.Logger) ([]string, error) {
	if utils.PDHEnabled() {
		return []string{}, nil
	}

	return []string{"Processor Information"}, nil
}

func (c *Collector) Close(_ *slog.Logger) error {
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *wmi.Client) error {
	if utils.PDHEnabled() {
		counters := []string{
			c1TimeSeconds,
			c2TimeSeconds,
			c3TimeSeconds,
			c1TransitionsTotal,
			c2TransitionsTotal,
			c3TransitionsTotal,
			clockInterruptsTotal,
			dpcQueuedPerSecond,
			dpcTimeSeconds,
			idleBreakEventsTotal,
			idleTimeSeconds,
			interruptsTotal,
			interruptTimeSeconds,
			parkingStatus,
			performanceLimitPercent,
			priorityTimeSeconds,
			privilegedTimeSeconds,
			privilegedUtilitySeconds,
			processorFrequencyMHz,
			processorPerformance,
			processorTimeSeconds,
			processorUtilityRate,
			userTimeSeconds,
		}

		var err error

		c.perfDataCollector, err = perfdata.NewCollector("Processor Information", []string{"*"}, counters)
		if err != nil {
			return fmt.Errorf("failed to create Processor Information collector: %w", err)
		}
	}

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
		prometheus.BuildFQName(types.Namespace, Name, "time_total"),
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
	c.cStateSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cstate_seconds_total"),
		"Time spent in low-power idle state",
		[]string{"core", "state"},
		nil,
	)
	c.timeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "time_total"),
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
	c.processorFrequencyMHz = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "core_frequency_mhz"),
		"Core frequency in megahertz",
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

	return nil
}

func (c *Collector) Collect(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	if utils.PDHEnabled() {
		return c.collectPDH(ch)
	}

	logger = logger.With(slog.String("collector", Name))

	return c.collectFull(ctx, logger, ch)
}

func (c *Collector) collectFull(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	data := make([]perflibProcessorInformation, 0)

	err := perflib.UnmarshalObject(ctx.PerfObjects["Processor Information"], &data, logger)
	if err != nil {
		return err
	}

	var coreCount float64

	for _, cpu := range data {
		if strings.Contains(strings.ToLower(cpu.Name), "_total") {
			continue
		}

		core := cpu.Name

		coreCount++

		ch <- prometheus.MustNewConstMetric(
			c.cStateSecondsTotal,
			prometheus.CounterValue,
			cpu.C1TimeSeconds,
			core, "c1",
		)
		ch <- prometheus.MustNewConstMetric(
			c.cStateSecondsTotal,
			prometheus.CounterValue,
			cpu.C2TimeSeconds,
			core, "c2",
		)
		ch <- prometheus.MustNewConstMetric(
			c.cStateSecondsTotal,
			prometheus.CounterValue,
			cpu.C3TimeSeconds,
			core, "c3",
		)

		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			cpu.IdleTimeSeconds,
			core, "idle",
		)
		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			cpu.InterruptTimeSeconds,
			core, "interrupt",
		)
		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			cpu.DPCTimeSeconds,
			core, "dpc",
		)
		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			cpu.PrivilegedTimeSeconds,
			core, "privileged",
		)
		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			cpu.UserTimeSeconds,
			core, "user",
		)

		ch <- prometheus.MustNewConstMetric(
			c.interruptsTotal,
			prometheus.CounterValue,
			cpu.InterruptsTotal,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.dpcsTotal,
			prometheus.CounterValue,
			cpu.DPCsQueuedTotal,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.clockInterruptsTotal,
			prometheus.CounterValue,
			cpu.ClockInterruptsTotal,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.idleBreakEventsTotal,
			prometheus.CounterValue,
			cpu.IdleBreakEventsTotal,
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.parkingStatus,
			prometheus.GaugeValue,
			cpu.ParkingStatus,
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.processorFrequencyMHz,
			prometheus.GaugeValue,
			cpu.ProcessorFrequencyMHz,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.processorPerformance,
			prometheus.CounterValue,
			cpu.ProcessorPerformance,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.processorMPerf,
			prometheus.CounterValue,
			cpu.ProcessorMPerf,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.processorRTC,
			prometheus.CounterValue,
			cpu.ProcessorRTC,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.processorUtility,
			prometheus.CounterValue,
			cpu.ProcessorUtilityRate,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.processorPrivilegedUtility,
			prometheus.CounterValue,
			cpu.PrivilegedUtilitySeconds,
			core,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.logicalProcessors,
		prometheus.GaugeValue,
		coreCount,
	)

	return nil
}

func (c *Collector) collectPDH(ch chan<- prometheus.Metric) error {
	data, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect Processor Information metrics: %w", err)
	}

	var coreCount float64

	for core, coreData := range data {
		coreCount++

		ch <- prometheus.MustNewConstMetric(
			c.cStateSecondsTotal,
			prometheus.CounterValue,
			coreData[c1TimeSeconds].FirstValue,
			core, "c1",
		)
		ch <- prometheus.MustNewConstMetric(
			c.cStateSecondsTotal,
			prometheus.CounterValue,
			coreData[c2TimeSeconds].FirstValue,
			core, "c2",
		)
		ch <- prometheus.MustNewConstMetric(
			c.cStateSecondsTotal,
			prometheus.CounterValue,
			coreData[c3TimeSeconds].FirstValue,
			core, "c3",
		)

		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			coreData[idleTimeSeconds].FirstValue,
			core, "idle",
		)
		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			coreData[interruptTimeSeconds].FirstValue,
			core, "interrupt",
		)
		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			coreData[dpcTimeSeconds].FirstValue,
			core, "dpc",
		)
		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			coreData[privilegedTimeSeconds].FirstValue,
			core, "privileged",
		)
		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			coreData[userTimeSeconds].FirstValue,
			core, "user",
		)

		ch <- prometheus.MustNewConstMetric(
			c.interruptsTotal,
			prometheus.CounterValue,
			coreData[interruptsTotal].FirstValue,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.dpcsTotal,
			prometheus.CounterValue,
			coreData[dpcQueuedPerSecond].FirstValue,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.clockInterruptsTotal,
			prometheus.CounterValue,
			coreData[clockInterruptsTotal].FirstValue,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.idleBreakEventsTotal,
			prometheus.CounterValue,
			coreData[idleBreakEventsTotal].FirstValue,
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.parkingStatus,
			prometheus.GaugeValue,
			coreData[parkingStatus].FirstValue,
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.processorFrequencyMHz,
			prometheus.GaugeValue,
			coreData[processorFrequencyMHz].FirstValue,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.processorPerformance,
			prometheus.CounterValue,
			coreData[processorPerformance].FirstValue,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.processorMPerf,
			prometheus.CounterValue,
			coreData[processorPerformance].SecondValue,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.processorRTC,
			prometheus.CounterValue,
			coreData[processorUtilityRate].SecondValue,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.processorUtility,
			prometheus.CounterValue,
			coreData[processorUtilityRate].FirstValue,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.processorPrivilegedUtility,
			prometheus.CounterValue,
			coreData[privilegedUtilitySeconds].FirstValue,
			core,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.logicalProcessors,
		prometheus.GaugeValue,
		coreCount,
	)

	return nil
}
