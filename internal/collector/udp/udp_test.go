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

package udp_test

import (
	"testing"

	"github.com/prometheus-community/windows_exporter/internal/collector/udp"
	"github.com/prometheus-community/windows_exporter/internal/utils/testutils"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func BenchmarkCollector(b *testing.B) {
	testutils.FuncBenchmarkCollector(b, udp.Name, udp.NewWithFlags)
}

func TestCollector(t *testing.T) {
	metrics := testutils.TestCollector(t, udp.New, nil)

	// All four datagram totals are cumulative raw PDH values and documented as
	// counters in docs/collector.udp.md; datagram_received_total used to be
	// published as a gauge.
	for _, name := range []string{
		"windows_udp_datagram_no_port_total",
		"windows_udp_datagram_received_errors_total",
		"windows_udp_datagram_received_total",
		"windows_udp_datagram_sent_total",
	} {
		require.Contains(t, metrics, name)
		require.Equal(t, dto.MetricType_COUNTER, metrics[name].GetType(), name)
	}
}
