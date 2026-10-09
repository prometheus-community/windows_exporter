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

package dhcp

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestDenialCounterValues(t *testing.T) {
	t.Parallel()

	c := New(&Config{CollectorsEnabled: []string{subCollectorServerMetrics}})
	// Build creates descriptors before attempting to connect to the optional DHCP provider.
	_ = c.Build(slog.New(slog.DiscardHandler), nil)
	if c.perfDataCollector != nil {
		c.perfDataCollector.Close()
	}

	c.perfDataCollector = denialFixture{
		"Denied due to match.":     7,
		"Denied due to non-match.": 11,
	}
	metrics := make(chan prometheus.Metric, 100)
	require.NoError(t, c.collectServerMetrics(metrics))
	close(metrics)

	found := make(map[*prometheus.Desc]float64)

	for metric := range metrics {
		var value dto.Metric
		require.NoError(t, metric.Write(&value))
		found[metric.Desc()] = value.GetCounter().GetValue()
	}

	require.InDelta(t, 7, found[c.deniedDueToMatch], 1e-9)
	require.InDelta(t, 11, found[c.deniedDueToNonMatch], 1e-9)
}

type denialFixture map[string]float64

func (f denialFixture) Collect(dst *[]perfDataCounterValues) error {
	var row perfDataCounterValues

	value := reflect.ValueOf(&row).Elem()
	for _, field := range reflect.VisibleFields(value.Type()) {
		value.FieldByIndex(field.Index).SetFloat(f[field.Tag.Get("perfdata")])
	}

	*dst = []perfDataCounterValues{row}

	return nil
}

func (f denialFixture) Close() {}
