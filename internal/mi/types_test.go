// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

package mi_test

import (
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/stretchr/testify/require"
)

func TestNewInterval(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		duration time.Duration
		expected mi.Interval
	}{
		{
			name:     "zero",
			duration: 0,
			expected: mi.Interval{},
		},
		{
			name:     "sub-minute",
			duration: 9*time.Second + 500*time.Millisecond,
			expected: mi.Interval{Seconds: 9, Microseconds: 500000},
		},
		{
			name:     "over a minute",
			duration: 90 * time.Second,
			expected: mi.Interval{Minutes: 1, Seconds: 30},
		},
		{
			name:     "hours minutes seconds",
			duration: 2*time.Hour + 3*time.Minute + 4*time.Second + 500*time.Millisecond,
			expected: mi.Interval{Hours: 2, Minutes: 3, Seconds: 4, Microseconds: 500000},
		},
		{
			name:     "over a day",
			duration: 26 * time.Hour,
			expected: mi.Interval{Days: 1, Hours: 2},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.expected, *mi.NewInterval(tc.duration))
		})
	}
}
