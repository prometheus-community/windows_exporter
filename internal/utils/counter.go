// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	c.lastValue = value
}

func (c *Counter) Value() float64 {
	return c.totalValue
}
