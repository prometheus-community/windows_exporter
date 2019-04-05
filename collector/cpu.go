// +build windows

package collector

import (
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

func init() {
	Factories["cpu"] = newCPUCollector
}

// A function to get Windows version from registry
func getWindowsVersion() float64 {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		log.Warn("Couldn't open registry", err)
		return 0
	}
	defer func() {
		err = k.Close()
		if err != nil {
			log.Warnf("Failed to close registry key: %v", err)
		}
	}()

	currentv, _, err := k.GetStringValue("CurrentVersion")
	if err != nil {
		log.Warn("Couldn't open registry to determine current Windows version:", err)
		return 0
	}

	currentv_flt, err := strconv.ParseFloat(currentv, 64)

	log.Debugf("Detected Windows version %f\n", currentv_flt)

	return currentv_flt
}

type cpuCollectorBasic struct {
	CStateSecondsTotal *prometheus.Desc
	TimeTotal          *prometheus.Desc
	InterruptsTotal    *prometheus.Desc
	DPCsTotal          *prometheus.Desc
}
type cpuCollectorFull struct {
	CStateSecondsTotal       *prometheus.Desc
	TimeTotal                *prometheus.Desc
	InterruptsTotal          *prometheus.Desc
	DPCsTotal                *prometheus.Desc
	ClockInterruptsTotal     *prometheus.Desc
	IdleBreakEventsTotal     *prometheus.Desc
	ParkingStatus            *prometheus.Desc
	ProcessorFrequencyMHz    *prometheus.Desc
	ProcessorMaxFrequencyMHz *prometheus.Desc
	ProcessorPerformance     *prometheus.Desc
}

// newCPUCollector constructs a new cpuCollector, appropriate for the running OS
func newCPUCollector() (Collector, error) {
	const subsystem = "cpu"

	version := getWindowsVersion()
	// Windows version by number https://docs.microsoft.com/en-us/windows/desktop/sysinfo/operating-system-version
	// For Windows 2008 or earlier Windows version is 6.0 or lower, where we only have the older "Processor" counters
	// For Windows 2008 R2 or later Windows version is 6.1 or higher, so we can use "ProcessorInformation" counters
	// Value 6.05 was selected just to split between Windows versions
	if version < 6.05 {
		return &cpuCollectorBasic{
			CStateSecondsTotal: prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, subsystem, "cstate_seconds_total"),
				"Time spent in low-power idle state",
				[]string{"core", "state"},
				nil,
			),
			TimeTotal: prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, subsystem, "time_total"),
				"Time that processor spent in different modes (idle, user, system, ...)",
				[]string{"core", "mode"},
				nil,
			),
			InterruptsTotal: prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, subsystem, "interrupts_total"),
				"Total number of received and serviced hardware interrupts",
				[]string{"core"},
				nil,
			),
			DPCsTotal: prometheus.NewDesc(
				prometheus.BuildFQName(Namespace, subsystem, "dpcs_total"),
				"Total number of received and serviced deferred procedure calls (DPCs)",
				[]string{"core"},
				nil,
			),
		}, nil
	}

	return &cpuCollectorFull{
		CStateSecondsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "cstate_seconds_total"),
			"Time spent in low-power idle state",
			[]string{"core", "state"},
			nil,
		),
		TimeTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "time_total"),
			"Time that processor spent in different modes (idle, user, system, ...)",
			[]string{"core", "mode"},
			nil,
		),
		InterruptsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "interrupts_total"),
			"Total number of received and serviced hardware interrupts",
			[]string{"core"},
			nil,
		),
		DPCsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "dpcs_total"),
			"Total number of received and serviced deferred procedure calls (DPCs)",
			[]string{"core"},
			nil,
		),
		ClockInterruptsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "clock_interrupts_total"),
			"Total number of received and serviced clock tick interrupts",
			[]string{"core"},
			nil,
		),
		IdleBreakEventsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "idle_break_events_total"),
			"Total number of time processor was woken from idle",
			[]string{"core"},
			nil,
		),
		ParkingStatus: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "parking_status"),
			"Parking Status represents whether a processor is parked or not",
			[]string{"core"},
			nil,
		),
		ProcessorFrequencyMHz: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "core_frequency_mhz"),
			"Core frequency in megahertz",
			[]string{"core"},
			nil,
		),
		ProcessorPerformance: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "processor_performance"),
			"Processor Performance is the average performance of the processor while it is executing instructions, as a percentage of the nominal performance of the processor. On some processors, Processor Performance may exceed 100%",
			[]string{"core"},
			nil,
		),
	}, nil
}

type perflibProcessor struct {
	Name                  string
	C1Transitions         float64 `perflib:"C1 Transitions/sec"`
	C2Transitions         float64 `perflib:"C2 Transitions/sec"`
	C3Transitions         float64 `perflib:"C3 Transitions/sec"`
	DPCRate               float64 `perflib:"DPC Rate"`
	DPCsQueued            float64 `perflib:"DPCs Queued/sec"`
	Interrupts            float64 `perflib:"Interrupts/sec"`
	PercentC2Time         float64 `perflib:"% C1 Time"`
	PercentC3Time         float64 `perflib:"% C2 Time"`
	PercentC1Time         float64 `perflib:"% C3 Time"`
	PercentDPCTime        float64 `perflib:"% DPC Time"`
	PercentIdleTime       float64 `perflib:"% Idle Time"`
	PercentInterruptTime  float64 `perflib:"% Interrupt Time"`
	PercentPrivilegedTime float64 `perflib:"% Privileged Time"`
	PercentProcessorTime  float64 `perflib:"% Processor Time"`
	PercentUserTime       float64 `perflib:"% User Time"`
}

func (c *cpuCollectorBasic) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	data := make([]perflibProcessor, 0)
	err := unmarshalObject(ctx.perfObjects["Processor"], &data)
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
	ProcessorTimeSeconds     float64 `perflib:"% Processor Time"`
	ProcessorUtilityRate     float64 `perflib:"% Processor Utility"`
	UserTimeSeconds          float64 `perflib:"% User Time"`
}

func (c *cpuCollectorFull) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	data := make([]perflibProcessorInformation, 0)
	err := unmarshalObject(ctx.perfObjects["Processor Information"], &data)
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
			prometheus.GaugeValue,
			cpu.ProcessorPerformance,
			core,
		)
	}

	return nil
}
