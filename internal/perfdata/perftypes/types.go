package perftypes

import "github.com/prometheus/client_golang/prometheus"

const EmptyInstance = "------"

var TotalInstance = []string{"_Total"}

type CounterValues struct {
	Type        prometheus.ValueType
	FirstValue  float64
	SecondValue float64
}
