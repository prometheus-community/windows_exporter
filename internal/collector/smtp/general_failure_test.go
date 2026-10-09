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

package smtp

import (
	"log/slog"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestGeneralFailureMetric(t *testing.T) {
	t.Parallel()

	c := New(nil)
	// Descriptors are initialized before connecting to the optional SMTP provider.
	_ = c.Build(slog.New(slog.DiscardHandler), nil)
	if c.perfDataCollector != nil {
		c.perfDataCollector.Close()
	}

	c.perfDataCollector = generalFailureFixture{}
	metrics := make(chan prometheus.Metric, 100)
	require.NoError(t, c.Collect(metrics, 0))
	close(metrics)

	found := false

	for metric := range metrics {
		if metric.Desc() != c.badMailedMessagesGeneralFailureTotal {
			continue
		}

		var value dto.Metric
		require.NoError(t, metric.Write(&value))
		require.InDelta(t, 42, value.GetCounter().GetValue(), 1e-9)
		require.Equal(t, "mail", value.GetLabel()[0].GetValue())

		found = true
	}

	require.True(t, found)
}

type generalFailureFixture struct{}

func (generalFailureFixture) Collect(dst *[]perfDataCounterValues) error {
	*dst = []perfDataCounterValues{{Name: "mail", BadmailedMessagesGeneralFailureTotal: 42}}

	return nil
}
func (generalFailureFixture) Close() {}

func TestCloseBeforeProviderInitialization(t *testing.T) {
	t.Parallel()
	require.NoError(t, New(nil).Close())
}
