// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

// Package multiplier applies the optional multiplier of a user-defined metric
// of the wmi, performancecounter and registry collectors. It converts values
// to base units, for example milliseconds to seconds with 0.001.
package multiplier

import (
	"fmt"
	"math"
)

// maxExactInteger bounds the divisor. Beyond it, several integers round to the
// same reciprocal, so the rounded divisor may not be the one the user meant.
const maxExactInteger = 1 << 53

// Multiplier scales values. The zero value leaves them unchanged.
type Multiplier struct {
	factor  float64
	divisor float64
}

// New validates the multiplier of the configuration; nil means none. It must
// be a positive finite number, so that counters stay monotonic.
func New(multiplier *float64) (Multiplier, error) {
	if multiplier == nil {
		return Multiplier{}, nil
	}

	m := *multiplier

	if math.IsNaN(m) || math.IsInf(m, 0) || m <= 0 {
		return Multiplier{}, fmt.Errorf("multiplier must be a positive finite number, got %v", m)
	}

	if m == 1 {
		return Multiplier{}, nil
	}

	// x * 0.001 is not always the correctly rounded x / 1000, because 0.001
	// has no exact binary representation: 9 * 0.001 is 0.009000000000000001.
	// Divide by the reciprocal instead when it is an integer.
	if m < 1 {
		if d := math.Round(1 / m); d <= maxExactInteger && 1/d == m {
			return Multiplier{divisor: d}, nil
		}
	}

	return Multiplier{factor: m}, nil
}

// Apply returns v scaled by the multiplier.
func (m Multiplier) Apply(v float64) float64 {
	switch {
	case m.divisor != 0:
		return v / m.divisor
	case m.factor != 0:
		return v * m.factor
	default:
		return v
	}
}
