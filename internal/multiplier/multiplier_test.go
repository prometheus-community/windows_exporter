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

package multiplier_test

import (
	"math"
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/multiplier"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		multiplier *float64
		value      float64
		want       float64
	}{
		{"unset", nil, 42, 42},
		{"one", new(1.0), 42, 42},
		{"hours to seconds", new(3600.0), 2, 7200},
		{"megabytes to bytes", new(1048576.0), 3, 3145728},
		{"milliseconds to seconds", new(0.001), 9, 0.009},
		{"percent to ratio", new(0.01), 7, 0.07},
		{"ticks to seconds", new(1e-7), 115960642, 11.5960642},
		{"non-integer reciprocal", new(0.4), 5, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m, err := multiplier.New(tc.multiplier)
			require.NoError(t, err)
			require.Equal(t, tc.want, m.Apply(tc.value)) //nolint:testifylint // The result must be exact.
		})
	}
}

func TestNewInvalid(t *testing.T) {
	t.Parallel()

	for _, m := range []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		_, err := multiplier.New(&m)
		require.Error(t, err, "multiplier %v", m)
	}
}

// TestApplyDividesExactly checks that a multiplier that is the reciprocal of
// an integer gives the correctly rounded quotient. A plain multiplication
// misses it for some values, for example 9 * 0.001.
func TestApplyDividesExactly(t *testing.T) {
	t.Parallel()

	for _, divisor := range []float64{10, 100, 1000, 1024, 1e6, 1e7, 1e9} {
		m, err := multiplier.New(new(1 / divisor))
		require.NoError(t, err)

		for i := range 100000 {
			v := float64(i)
			if got := m.Apply(v); got != v/divisor {
				t.Fatalf("Apply(%v) with multiplier 1/%v = %v, want %v", v, divisor, got, v/divisor)
			}
		}
	}
}
