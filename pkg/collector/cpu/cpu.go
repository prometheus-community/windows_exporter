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
			C1TimeSeconds,
			C2TimeSeconds,
			C3TimeSeconds,
			C1TransitionsTotal,
			C2TransitionsTotal,
			C3TransitionsTotal,
			ClockInterruptsTotal,
			DPCsQueuedTotal,
			DPCTimeSeconds,
			IdleBreakEventsTotal,
			IdleTimeSeconds,
			InterruptsTotal,
			InterruptTimeSeconds,
			ParkingStatus,
			PerformanceLimitPercent,
			PriorityTimeSeconds,
			PrivilegedTimeSeconds,
			PrivilegedUtilitySeconds,
			ProcessorFrequencyMHz,
			ProcessorPerformance,
			ProcessorTimeSeconds,
			ProcessorUtilityRate,
			UserTimeSeconds,
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
	logger = logger.With(slog.String("collector", Name))

	if utils.PDHEnabled() {
		return c.collectPDH(ch)
	}

	return c.collectFull(ctx, logger, ch)
}

type perflibProcessorInformation struct {
	Name                     string
	C1TimeSeconds            float64 `perflib:"% C1 Time"`
	C2TimeSeconds            float64 `perflib:"% C2 Time"`
	C3TimeSeconds            float64 `perflib:"% C3 Time"`
	C1TransitionsTotal       float64 `perflib:"C1 Transitions/sec"`
	C2TransitionsTotal       float64 `perflib:"C2 Transitions/sec"`
	C3TransitionsTotal       float64 `perflib:"C3 Transitions/sec"`
	ClockInterruptsTotal     float64 `perflib:"Clock Interrupts/sec"`
	DPCsQueuedTotal          float64 `perflib:"DPCs Queued/sec"`
	DPCTimeSeconds           float64 `perflib:"% DPC Time"`
	IdleBreakEventsTotal     float64 `perflib:"Idle Break Events/sec"`
	IdleTimeSeconds          float64 `perflib:"% Idle Time"`
	InterruptsTotal          float64 `perflib:"Interrupts/sec"`
	InterruptTimeSeconds     float64 `perflib:"% Interrupt Time"`
	ParkingStatus            float64 `perflib:"Parking Status"`
	PerformanceLimitPercent  float64 `perflib:"% Performance Limit"`
	PriorityTimeSeconds      float64 `perflib:"% Priority Time"`
	PrivilegedTimeSeconds    float64 `perflib:"% Privileged Time"`
	PrivilegedUtilitySeconds float64 `perflib:"% Privileged Utility"`
	ProcessorFrequencyMHz    float64 `perflib:"Processor Frequency"`
	ProcessorPerformance     float64 `perflib:"% Processor Performance"`
	ProcessorMPerf           float64 `perflib:"% Processor Performance,secondvalue"`
	ProcessorTimeSeconds     float64 `perflib:"% Processor Time"`
	ProcessorUtilityRate     float64 `perflib:"% Processor Utility"`
	ProcessorRTC             float64 `perflib:"% Processor Utility,secondvalue"`
	UserTimeSeconds          float64 `perflib:"% User Time"`
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
			coreData[C1TimeSeconds].FirstValue,
			core, "c1",
		)
		ch <- prometheus.MustNewConstMetric(
			c.cStateSecondsTotal,
			prometheus.CounterValue,
			coreData[C2TimeSeconds].FirstValue,
			core, "c2",
		)
		ch <- prometheus.MustNewConstMetric(
			c.cStateSecondsTotal,
			prometheus.CounterValue,
			coreData[C3TimeSeconds].FirstValue,
			core, "c3",
		)

		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			coreData[IdleTimeSeconds].FirstValue,
			core, "idle",
		)
		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			coreData[InterruptTimeSeconds].FirstValue,
			core, "interrupt",
		)
		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			coreData[DPCTimeSeconds].FirstValue,
			core, "dpc",
		)
		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			coreData[PrivilegedTimeSeconds].FirstValue,
			core, "privileged",
		)
		ch <- prometheus.MustNewConstMetric(
			c.timeTotal,
			prometheus.CounterValue,
			coreData[UserTimeSeconds].FirstValue,
			core, "user",
		)

		ch <- prometheus.MustNewConstMetric(
			c.interruptsTotal,
			prometheus.CounterValue,
			coreData[InterruptsTotal].FirstValue,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.dpcsTotal,
			prometheus.CounterValue,
			coreData[DPCsQueuedTotal].FirstValue,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.clockInterruptsTotal,
			prometheus.CounterValue,
			coreData[ClockInterruptsTotal].FirstValue,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.idleBreakEventsTotal,
			prometheus.CounterValue,
			coreData[IdleBreakEventsTotal].FirstValue,
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.parkingStatus,
			prometheus.GaugeValue,
			coreData[ParkingStatus].FirstValue,
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.processorFrequencyMHz,
			prometheus.GaugeValue,
			coreData[ProcessorFrequencyMHz].FirstValue,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.processorPerformance,
			prometheus.CounterValue,
			coreData[ProcessorPerformance].FirstValue,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.processorMPerf,
			prometheus.CounterValue,
			coreData[ProcessorPerformance].SecondValue,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.processorRTC,
			prometheus.CounterValue,
			coreData[ProcessorUtilityRate].SecondValue,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.processorUtility,
			prometheus.CounterValue,
			coreData[ProcessorUtilityRate].FirstValue,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.processorPrivilegedUtility,
			prometheus.CounterValue,
			coreData[PrivilegedUtilitySeconds].FirstValue,
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
