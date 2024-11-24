//go:build windows

package perfdata

import "github.com/prometheus/client_golang/prometheus"

const (
	InstanceEmpty = "------"
	InstanceTotal = "_Total"
)

type CounterValues struct {
	Type        prometheus.ValueType
	FirstValue  float64
	SecondValue float64
}
