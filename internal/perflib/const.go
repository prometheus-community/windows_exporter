package perflib

// Based on https://github.com/leoluk/perflib_exporter/blob/master/collector/mapper.go

const (
	PERF_COUNTER_COUNTER       = 0x10410400
	PERF_100NSEC_TIMER         = 0x20510500
	PERF_PRECISION_100NS_TIMER = 0x20570500
	PERF_ELAPSED_TIME          = 0x30240500
)
