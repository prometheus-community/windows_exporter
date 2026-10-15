// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
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

// Counter32 extends a raw 32-bit performance counter, such as a
// PERF_COUNTER_COUNTER or the numerator of a PERF_AVERAGE_TIMER, beyond 2^32.
// Unlike Counter, it starts at the raw value and follows a restart of the
// source, so its value equals the raw value until the first wrap. It can only
// detect one wrap between two samples.
type Counter32 struct {
	lastValue    uint32
	totalValue   float64
	restartBelow uint32
}

// NewCounter32 creates a Counter32 from the first raw value. If restartBelow is
// not zero, a decrease from a raw value below restartBelow is taken as a
// restart of the source instead of a wrap. Use it for counters that cannot
// grow by 2^32-restartBelow between two samples, such as request counts.
func NewCounter32(value float64, restartBelow uint32) Counter32 {
	raw := toUint32(value)

	return Counter32{
		lastValue:    raw,
		totalValue:   float64(raw),
		restartBelow: restartBelow,
	}
}

// Update adds a raw value and returns the extended value. reset reports that
// the source restarted, for example because a share was reconnected; the raw
// value is then taken as is. Any other decrease is a wrap, unless it starts
// below restartBelow.
func (c *Counter32) Update(value float64, reset bool) float64 {
	raw := toUint32(value)

	if reset || (raw < c.lastValue && c.lastValue < c.restartBelow) {
		c.totalValue = float64(raw)
	} else {
		c.totalValue += float64(raw - c.lastValue)
	}

	c.lastValue = raw

	return c.totalValue
}

// toUint32 converts through int64, so a sign-extended value keeps its low
// 32 bits on every architecture.
func toUint32(value float64) uint32 {
	return uint32(int64(value))
}
