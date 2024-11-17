//go:build windows

package utils

type Counter struct {
	lastValue  uint32
	totalValue float64
}

// NewCounter creates a new Counter that accepts uint32 values and returns float64 values.
// It resolve the overflow issue of uint32 by using the difference between the last value and the current value.
func NewCounter(lastValue uint32) Counter {
	return Counter{
		lastValue:  lastValue,
		totalValue: 0,
	}
}

func (c *Counter) AddValue(value uint32) {
	c.totalValue += float64(value - c.lastValue)
}

func (c *Counter) Value() float64 {
	return c.totalValue
}
