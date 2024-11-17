//go:build windows

package perfdata

import "github.com/prometheus/client_golang/prometheus"

const EmptyInstance = "------"

type CounterValues struct {
	Type        prometheus.ValueType
	FirstValue  float64
	SecondValue float64
}
