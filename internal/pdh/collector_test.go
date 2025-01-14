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

package pdh_test

import (
	"testing"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type process struct {
	Name        string
	ThreadCount float64 `perfdata:"Thread Count"`
}

func TestCollector(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		object    string
		instances []string
	}{
		{
			object:    "Process",
			instances: []string{"*"},
		},
	} {
		t.Run(tc.object, func(t *testing.T) {
			t.Parallel()

			performanceData, err := pdh.NewCollector[process](pdh.CounterTypeRaw, tc.object, tc.instances)
			require.NoError(t, err)

			time.Sleep(100 * time.Millisecond)

			var data []process

			err = performanceData.Collect(&data)
			require.NoError(t, err)
			require.NotEmpty(t, data)

			err = performanceData.Collect(&data)
			require.NoError(t, err)
			require.NotEmpty(t, data)

			for _, instance := range data {
				if instance.Name == "Idle" || instance.Name == "Secure System" {
					continue
				}

				assert.NotZerof(t, instance.ThreadCount, "object: %s, instance: %s, counter: %s", tc.object, instance, instance.ThreadCount)
			}
		})
	}
}
