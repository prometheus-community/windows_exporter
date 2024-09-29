package cpu

// Processor performance counters.
const (
	c1TimeSeconds            = "% C1 Time"
	c2TimeSeconds            = "% C2 Time"
	c3TimeSeconds            = "% C3 Time"
	c1TransitionsTotal       = "C1 Transitions/sec"
	c2TransitionsTotal       = "C2 Transitions/sec"
	c3TransitionsTotal       = "C3 Transitions/sec"
	clockInterruptsTotal     = "Clock Interrupts/sec"
	dpcQueuedPerSecond       = "DPCs Queued/sec"
	dpcTimeSeconds           = "% DPC Time"
	idleBreakEventsTotal     = "Idle Break Events/sec"
	idleTimeSeconds          = "% Idle Time"
	interruptsTotal          = "Interrupts/sec"
	interruptTimeSeconds     = "% Interrupt Time"
	parkingStatus            = "Parking Status"
	performanceLimitPercent  = "% Performance Limit"
	priorityTimeSeconds      = "% Priority Time"
	privilegedTimeSeconds    = "% Privileged Time"
	privilegedUtilitySeconds = "% Privileged Utility"
	processorFrequencyMHz    = "Processor Frequency"
	processorPerformance     = "% Processor Performance"
	processorTimeSeconds     = "% Processor Time"
	processorUtilityRate     = "% Processor Utility"
	userTimeSeconds          = "% User Time"
)

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
