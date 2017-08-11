// returns data points from
// Win32_PerfRawData_HvStats_HyperVHypervisor
// Win32_PerfRawData_HvStats_HyperVHypervisorLogicalProcessor

package collector

import (
	"log"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	Factories["hyperv"] = NewHypervCollector
}

// A HypervCollector is a Prometheus collector for WMI
// Win32_PerfRawData_HvStats_HyperVHypervisorLogicalProcessor
// and Win32_PerfRawData_HvStats_HyperVHypervisor metrics.
type HypervCollector struct {
	// Win32_PerfRawData_HvStats_HyperVHypervisor
	LogicalProcessors      *prometheus.Desc
	MonitoredNotifications *prometheus.Desc
	Partitions             *prometheus.Desc
	TotalPages             *prometheus.Desc
	VirtualProcessors      *prometheus.Desc

	//Win32_PerfRawData_HvStats_HyperVHypervisorLogicalProcessor
	C1TransitionsPersec                *prometheus.Desc
	C2TransitionsPersec                *prometheus.Desc
	C3TransitionsPersec                *prometheus.Desc
	ContextSwitchesPersec              *prometheus.Desc
	Frequency                          *prometheus.Desc
	HardwareInterruptsPersec           *prometheus.Desc
	InterProcessorInterruptsPersec     *prometheus.Desc
	InterProcessorInterruptsSentPersec *prometheus.Desc
	MonitorTransitionCost              *prometheus.Desc
	ParkingStatus                      *prometheus.Desc
	PercentC1Time                      *prometheus.Desc
	PercentC2Time                      *prometheus.Desc
	PercentC3Time                      *prometheus.Desc
	PercentGuestRunTime                *prometheus.Desc
	PercentHypervisorRunTime           *prometheus.Desc
	PercentIdleTime                    *prometheus.Desc
	PercentofMaxFrequency              *prometheus.Desc
	PercentTotalRunTime                *prometheus.Desc
	ProcessorStateFlags                *prometheus.Desc
	RootVpIndex                        *prometheus.Desc
	SchedulerInterruptsPersec          *prometheus.Desc
	TimerInterruptsPersec              *prometheus.Desc
	TotalInterruptsPersec              *prometheus.Desc
}

func NewHypervCollector() (Collector, error) {
	const subsystem = "hypervisor"

	//func NewDesc(fqName, help string, variableLabels []string, constLabels Labels) *Desc
	//func BuildFQName(namespace, subsystem, name string) string

	return &HypervCollector{
		LogicalProcessors: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Logical_Processors_Number"),
			"The number of logical processors present in the system.",
			nil,
			nil,
		),

		MonitoredNotifications: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Monitored_Notifications_Number"),
			"The number of monitored notifications registered with the hypervisor.",
			nil,
			nil,
		),

		Partitions: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Partitions_Number"),
			"The number of partitions (virtual machines) present in the system.",
			nil,
			nil,
		),

		TotalPages: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Total_Pages_Number"),
			"The number of bootstrap and deposited pages in the hypervisor.",
			nil,
			nil,
		),

		VirtualProcessors: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Virtual_Processors_Number"),
			"The number of virtual processors present in the system.",
			nil,
			nil,
		),

		C1TransitionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "C1_Transitions_Persec"),
			"C1 Transitions/sec is the rate that CPU enters the C1 low-power idle state.",
			nil,
			nil,
		),

		C2TransitionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "C2_Transitions_Persec"),
			"C2 Transitions/sec is the rate that CPU enters the C2 low-power idle state.",
			nil,
			nil,
		),

		C3TransitionsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "C3_Transitions_Persec"),
			"C3 Transitions/sec is the rate that CPU enters the C3 low-power idle state.",
			nil,
			nil,
		),

		ContextSwitchesPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Context_Switches_Persec"),
			"The rate of virtual processor context switches on the processor.",
			nil,
			nil,
		),

		Frequency: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Processor_Frequency"),
			"Processor Frequency is the frequency of the current processor in megahertz.",
			nil,
			nil,
		),

		HardwareInterruptsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Hardware_Interrupts_Persec"),
			"The rate of hardware interrupts on the processor (excluding hypervisor interrupts).",
			nil,
			nil,
		),

		InterProcessorInterruptsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Inter_Processor_Interrupts_Persec"),
			"The rate of hypervisor inter-processor interrupts delivered to the processor.",
			nil,
			nil,
		),

		InterProcessorInterruptsSentPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Inter_Processor_Interrupts_Sent_Persec"),
			"The rate of hypervisor inter-processor interrupts sent by the processor.",
			nil,
			nil,
		),

		MonitorTransitionCost: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Monitor_Transition_Cost"),
			"The hardware cost of transitions into the hypervisor.",
			nil,
			nil,
		),

		ParkingStatus: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Parking_Status"),
			"Parking Status represents whether a processor is parked or not.",
			nil,
			nil,
		),

		PercentC1Time: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Percent_C1_Time"),
			"% C1 Time is the percentage of time the processor spends in the C1 low-power idle state. % C1 Time is a subset of the total processor idle time.",
			nil,
			nil,
		),

		PercentC2Time: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Percent_C2_Time"),
			"% C2 Time is the percentage of time the processor spends in the C2 low-power idle state. % C2 Time is a subset of the total processor idle time.",
			nil,
			nil,
		),

		PercentC3Time: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Percent_C3_Time"),
			"% C3 Time is the percentage of time the processor spends in the C3 low-power idle state. % C3 Time is a subset of the total processor idle time.",
			nil,
			nil,
		),

		PercentGuestRunTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Percent_Guest_Run_Time"),
			"The percentage of time spent by the processor in guest code.",
			nil,
			nil,
		),

		PercentHypervisorRunTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Percent_Hypervisor_Run_Time"),
			"The percentage of time spent by the processor in hypervisor code.",
			nil,
			nil,
		),

		PercentIdleTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Percent_Idle_Time"),
			"The percentage of time spent by the processor in an idle state.",
			nil,
			nil,
		),

		PercentofMaxFrequency: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Percent_of_Max_Frequency"),
			"% of Maximum Frequency is the percentage of the current processor's maximum frequency.",
			nil,
			nil,
		),

		PercentTotalRunTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Percent_Total_Run_Time"),
			"The percentage of time spent by the processor in guest and hypervisor code.",
			nil,
			nil,
		),

		ProcessorStateFlags: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Processor_State_Flags"),
			"Processor State Flags",
			nil,
			nil,
		),

		RootVpIndex: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Root_Virtual_Processor_Index"),
			"Index of the root virtual processor that is affinity bound to this logical processor. A value that is greater than the maximum possible root VP index indicates no binding.",
			nil,
			nil,
		),

		SchedulerInterruptsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Scheduler_Interrupts_Persec"),
			"The rate of hypervisor scheduler interrupts on the processor.",
			nil,
			nil,
		),

		TimerInterruptsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Timer_Interrupts_Persec"),
			"The rate of hypervisor timer interrupts on the processor.",
			nil,
			nil,
		),

		TotalInterruptsPersec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "Total_Interrupts_Persec"),
			"The rate of hardware and hypervisor interrupts/sec.",
			nil,
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *HypervCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collectStats(ch); err != nil {
		log.Println("[ERROR] failed collecting hyperv stats metrics:", desc, err)
		return err
	}

	if desc, err := c.collectProce(ch); err != nil {
		log.Println("[ERROR] failed collecting hyperv processor metrics:", desc, err)
		return err
	}

	return nil
}

type Win32_PerfRawData_HvStats_HyperVHypervisor struct {
	LogicalProcessors      uint64
	MonitoredNotifications uint64
	Partitions             uint64
	TotalPages             uint64
	VirtualProcessors      uint64
}

type Win32_PerfRawData_HvStats_HyperVHypervisorLogicalProcessor struct {
	C1TransitionsPersec                uint64
	C2TransitionsPersec                uint64
	C3TransitionsPersec                uint64
	ContextSwitchesPersec              uint64
	Frequency                          uint64
	HardwareInterruptsPersec           uint64
	InterProcessorInterruptsPersec     uint64
	InterProcessorInterruptsSentPersec uint64
	MonitorTransitionCost              uint64
	ParkingStatus                      uint64
	PercentC1Time                      uint64
	PercentC2Time                      uint64
	PercentC3Time                      uint64
	PercentGuestRunTime                uint64
	PercentHypervisorRunTime           uint64
	PercentIdleTime                    uint64
	PercentofMaxFrequency              uint64
	PercentTotalRunTime                uint64
	ProcessorStateFlags                uint64
	RootVpIndex                        uint64
	SchedulerInterruptsPersec          uint64
	TimerInterruptsPersec              uint64
	TotalInterruptsPersec              uint64
}

func (c *HypervCollector) collectStats(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_HvStats_HyperVHypervisor
	q := wmi.CreateQuery(&dst, "")
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	// func MustNewConstMetric(desc *Desc, valueType ValueType,
	// value float64, labelValues ...string) Metric

	ch <- prometheus.MustNewConstMetric(
		c.LogicalProcessors,
		prometheus.GaugeValue,
		float64(dst[0].LogicalProcessors),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MonitoredNotifications,
		prometheus.GaugeValue,
		float64(dst[0].MonitoredNotifications),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Partitions,
		prometheus.GaugeValue,
		float64(dst[0].Partitions),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TotalPages,
		prometheus.GaugeValue,
		float64(dst[0].TotalPages),
	)

	ch <- prometheus.MustNewConstMetric(
		c.VirtualProcessors,
		prometheus.GaugeValue,
		float64(dst[0].VirtualProcessors),
	)

	return nil, nil
}

func (c *HypervCollector) collectProce(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []Win32_PerfRawData_HvStats_HyperVHypervisorLogicalProcessor
	q := wmi.CreateQuery(&dst, "")
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}

	ch <- prometheus.MustNewConstMetric(
		c.C1TransitionsPersec,
		prometheus.GaugeValue,
		float64(dst[0].C1TransitionsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.C2TransitionsPersec,
		prometheus.GaugeValue,
		float64(dst[0].C2TransitionsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.C3TransitionsPersec,
		prometheus.GaugeValue,
		float64(dst[0].C3TransitionsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ContextSwitchesPersec,
		prometheus.CounterValue,
		float64(dst[0].ContextSwitchesPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.Frequency,
		prometheus.CounterValue,
		float64(dst[0].Frequency),
	)

	ch <- prometheus.MustNewConstMetric(
		c.HardwareInterruptsPersec,
		prometheus.GaugeValue,
		float64(dst[0].HardwareInterruptsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.InterProcessorInterruptsPersec,
		prometheus.GaugeValue,
		float64(dst[0].InterProcessorInterruptsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.MonitorTransitionCost,
		prometheus.GaugeValue,
		float64(dst[0].MonitorTransitionCost),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ParkingStatus,
		prometheus.GaugeValue,
		float64(dst[0].ParkingStatus),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentC1Time,
		prometheus.GaugeValue,
		float64(dst[0].PercentC1Time),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentC2Time,
		prometheus.GaugeValue,
		float64(dst[0].PercentC2Time),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentC3Time,
		prometheus.GaugeValue,
		float64(dst[0].PercentC3Time),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentGuestRunTime,
		prometheus.GaugeValue,
		float64(dst[0].PercentGuestRunTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentHypervisorRunTime,
		prometheus.GaugeValue,
		float64(dst[0].PercentHypervisorRunTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentIdleTime,
		prometheus.GaugeValue,
		float64(dst[0].PercentIdleTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentofMaxFrequency,
		prometheus.GaugeValue,
		float64(dst[0].PercentofMaxFrequency),
	)

	ch <- prometheus.MustNewConstMetric(
		c.PercentTotalRunTime,
		prometheus.GaugeValue,
		float64(dst[0].PercentTotalRunTime),
	)

	ch <- prometheus.MustNewConstMetric(
		c.ProcessorStateFlags,
		prometheus.GaugeValue,
		float64(dst[0].ProcessorStateFlags),
	)

	ch <- prometheus.MustNewConstMetric(
		c.RootVpIndex,
		prometheus.GaugeValue,
		float64(dst[0].RootVpIndex),
	)

	ch <- prometheus.MustNewConstMetric(
		c.SchedulerInterruptsPersec,
		prometheus.GaugeValue,
		float64(dst[0].SchedulerInterruptsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TimerInterruptsPersec,
		prometheus.GaugeValue,
		float64(dst[0].TimerInterruptsPersec),
	)

	ch <- prometheus.MustNewConstMetric(
		c.TotalInterruptsPersec,
		prometheus.GaugeValue,
		float64(dst[0].TotalInterruptsPersec),
	)

	return nil, nil
}
