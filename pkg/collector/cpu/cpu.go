//go:build windows

package cpu

import (
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/winversion"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "cpu"

type Config struct{}

var ConfigDefaults = Config{}

type Collector struct {
	logger log.Logger

	CStateSecondsTotal *prometheus.Desc
	TimeTotal          *prometheus.Desc
	InterruptsTotal    *prometheus.Desc
	DPCsTotal          *prometheus.Desc

	ClockInterruptsTotal     *prometheus.Desc
	IdleBreakEventsTotal     *prometheus.Desc
	ParkingStatus            *prometheus.Desc
	ProcessorFrequencyMHz    *prometheus.Desc
	ProcessorMaxFrequencyMHz *prometheus.Desc
	ProcessorPerformance     *prometheus.Desc
	ProcessorMPerf           *prometheus.Desc
	ProcessorRTC             *prometheus.Desc
	ProcessorUtility         *prometheus.Desc
	ProcessorPrivUtility     *prometheus.Desc
}

func New(logger log.Logger, _ *Config) *Collector {
	c := &Collector{}
	c.SetLogger(logger)

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	if winversion.WindowsVersionFloat > 6.05 {
		return []string{"Processor Information"}, nil
	}
	return []string{"Processor"}, nil
}

func (c *Collector) Build() error {
	c.CStateSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cstate_seconds_total"),
		"Time spent in low-power idle state",
		[]string{"core", "state"},
		nil,
	)
	c.TimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "time_total"),
		"Time that processor spent in different modes (dpc, idle, interrupt, privileged, user)",
		[]string{"core", "mode"},
		nil,
	)
	c.InterruptsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "interrupts_total"),
		"Total number of received and serviced hardware interrupts",
		[]string{"core"},
		nil,
	)
	c.DPCsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dpcs_total"),
		"Total number of received and serviced deferred procedure calls (DPCs)",
		[]string{"core"},
		nil,
	)

	// For Windows 2008 (version 6.0) or earlier we only have the "Processor"
	// class. As of Windows 2008 R2 (version 6.1) the more detailed
	// "Processor Information" set is available (although some of the counters
	// are added in later versions, so we aren't guaranteed to get all of
	// them).
	// Value 6.05 was selected to split between Windows versions.
	if winversion.WindowsVersionFloat < 6.05 {
		return nil
	}

	c.CStateSecondsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "cstate_seconds_total"),
		"Time spent in low-power idle state",
		[]string{"core", "state"},
		nil,
	)
	c.TimeTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "time_total"),
		"Time that processor spent in different modes (dpc, idle, interrupt, privileged, user)",
		[]string{"core", "mode"},
		nil,
	)
	c.InterruptsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "interrupts_total"),
		"Total number of received and serviced hardware interrupts",
		[]string{"core"},
		nil,
	)
	c.DPCsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "dpcs_total"),
		"Total number of received and serviced deferred procedure calls (DPCs)",
		[]string{"core"},
		nil,
	)
	c.ClockInterruptsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "clock_interrupts_total"),
		"Total number of received and serviced clock tick interrupts",
		[]string{"core"},
		nil,
	)
	c.IdleBreakEventsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "idle_break_events_total"),
		"Total number of time processor was woken from idle",
		[]string{"core"},
		nil,
	)
	c.ParkingStatus = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "parking_status"),
		"Parking Status represents whether a processor is parked or not",
		[]string{"core"},
		nil,
	)
	c.ProcessorFrequencyMHz = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "core_frequency_mhz"),
		"Core frequency in megahertz",
		[]string{"core"},
		nil,
	)
	c.ProcessorPerformance = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processor_performance_total"),
		"Processor Performance is the average performance of the processor while it is executing instructions, as a percentage of the nominal performance of the processor. On some processors, Processor Performance may exceed 100%",
		[]string{"core"},
		nil,
	)
	c.ProcessorMPerf = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processor_mperf_total"),
		"Processor MPerf is the number of TSC ticks incremented while executing instructions",
		[]string{"core"},
		nil,
	)
	c.ProcessorRTC = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processor_rtc_total"),
		"Processor RTC represents the number of RTC ticks made since the system booted. It should consistently be 64e6, and can be used to properly derive Processor Utility Rate",
		[]string{"core"},
		nil,
	)
	c.ProcessorUtility = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processor_utility_total"),
		"Processor Utility represents is the amount of time the core spends executing instructions",
		[]string{"core"},
		nil,
	)
	c.ProcessorPrivUtility = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "processor_privileged_utility_total"),
		"Processor Privileged Utility represents is the amount of time the core has spent executing instructions inside the kernel",
		[]string{"core"},
		nil,
	)

	return nil
}

func (c *Collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if winversion.WindowsVersionFloat > 6.05 {
		return c.CollectFull(ctx, ch)
	}

	return c.CollectBasic(ctx, ch)
}

type perflibProcessor struct {
	Name                  string
	C1Transitions         float64 `perflib:"C1 Transitions/sec"`
	C2Transitions         float64 `perflib:"C2 Transitions/sec"`
	C3Transitions         float64 `perflib:"C3 Transitions/sec"`
	DPCRate               float64 `perflib:"DPC Rate"`
	DPCsQueued            float64 `perflib:"DPCs Queued/sec"`
	Interrupts            float64 `perflib:"Interrupts/sec"`
	PercentC1Time         float64 `perflib:"% C1 Time"`
	PercentC2Time         float64 `perflib:"% C2 Time"`
	PercentC3Time         float64 `perflib:"% C3 Time"`
	PercentDPCTime        float64 `perflib:"% DPC Time"`
	PercentIdleTime       float64 `perflib:"% Idle Time"`
	PercentInterruptTime  float64 `perflib:"% Interrupt Time"`
	PercentPrivilegedTime float64 `perflib:"% Privileged Time"`
	PercentProcessorTime  float64 `perflib:"% Processor Time"`
	PercentUserTime       float64 `perflib:"% User Time"`
}

func (c *Collector) CollectBasic(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	data := make([]perflibProcessor, 0)
	err := perflib.UnmarshalObject(ctx.PerfObjects["Processor"], &data, c.logger)
	if err != nil {
		return err
	}

	for _, cpu := range data {
		if strings.Contains(strings.ToLower(cpu.Name), "_total") {
			continue
		}
		core := cpu.Name

		ch <- prometheus.MustNewConstMetric(
			c.CStateSecondsTotal,
			prometheus.CounterValue,
			cpu.PercentC1Time,
			core, "c1",
		)
		ch <- prometheus.MustNewConstMetric(
			c.CStateSecondsTotal,
			prometheus.CounterValue,
			cpu.PercentC2Time,
			core, "c2",
		)
		ch <- prometheus.MustNewConstMetric(
			c.CStateSecondsTotal,
			prometheus.CounterValue,
			cpu.PercentC3Time,
			core, "c3",
		)

		ch <- prometheus.MustNewConstMetric(
			c.TimeTotal,
			prometheus.CounterValue,
			cpu.PercentIdleTime,
			core, "idle",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TimeTotal,
			prometheus.CounterValue,
			cpu.PercentInterruptTime,
			core, "interrupt",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TimeTotal,
			prometheus.CounterValue,
			cpu.PercentDPCTime,
			core, "dpc",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TimeTotal,
			prometheus.CounterValue,
			cpu.PercentPrivilegedTime,
			core, "privileged",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TimeTotal,
			prometheus.CounterValue,
			cpu.PercentUserTime,
			core, "user",
		)

		ch <- prometheus.MustNewConstMetric(
			c.InterruptsTotal,
			prometheus.CounterValue,
			cpu.Interrupts,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.DPCsTotal,
			prometheus.CounterValue,
			cpu.DPCsQueued,
			core,
		)
	}

	return nil
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

func (c *Collector) CollectFull(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	data := make([]perflibProcessorInformation, 0)
	err := perflib.UnmarshalObject(ctx.PerfObjects["Processor Information"], &data, c.logger)
	if err != nil {
		return err
	}

	for _, cpu := range data {
		if strings.Contains(strings.ToLower(cpu.Name), "_total") {
			continue
		}
		core := cpu.Name

		ch <- prometheus.MustNewConstMetric(
			c.CStateSecondsTotal,
			prometheus.CounterValue,
			cpu.C1TimeSeconds,
			core, "c1",
		)
		ch <- prometheus.MustNewConstMetric(
			c.CStateSecondsTotal,
			prometheus.CounterValue,
			cpu.C2TimeSeconds,
			core, "c2",
		)
		ch <- prometheus.MustNewConstMetric(
			c.CStateSecondsTotal,
			prometheus.CounterValue,
			cpu.C3TimeSeconds,
			core, "c3",
		)

		ch <- prometheus.MustNewConstMetric(
			c.TimeTotal,
			prometheus.CounterValue,
			cpu.IdleTimeSeconds,
			core, "idle",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TimeTotal,
			prometheus.CounterValue,
			cpu.InterruptTimeSeconds,
			core, "interrupt",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TimeTotal,
			prometheus.CounterValue,
			cpu.DPCTimeSeconds,
			core, "dpc",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TimeTotal,
			prometheus.CounterValue,
			cpu.PrivilegedTimeSeconds,
			core, "privileged",
		)
		ch <- prometheus.MustNewConstMetric(
			c.TimeTotal,
			prometheus.CounterValue,
			cpu.UserTimeSeconds,
			core, "user",
		)

		ch <- prometheus.MustNewConstMetric(
			c.InterruptsTotal,
			prometheus.CounterValue,
			cpu.InterruptsTotal,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.DPCsTotal,
			prometheus.CounterValue,
			cpu.DPCsQueuedTotal,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ClockInterruptsTotal,
			prometheus.CounterValue,
			cpu.ClockInterruptsTotal,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.IdleBreakEventsTotal,
			prometheus.CounterValue,
			cpu.IdleBreakEventsTotal,
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ParkingStatus,
			prometheus.GaugeValue,
			cpu.ParkingStatus,
			core,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ProcessorFrequencyMHz,
			prometheus.GaugeValue,
			cpu.ProcessorFrequencyMHz,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ProcessorPerformance,
			prometheus.CounterValue,
			cpu.ProcessorPerformance,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ProcessorMPerf,
			prometheus.CounterValue,
			cpu.ProcessorMPerf,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ProcessorRTC,
			prometheus.CounterValue,
			cpu.ProcessorRTC,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ProcessorUtility,
			prometheus.CounterValue,
			cpu.ProcessorUtilityRate,
			core,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ProcessorPrivUtility,
			prometheus.CounterValue,
			cpu.PrivilegedUtilitySeconds,
			core,
		)
	}

	return nil
}
