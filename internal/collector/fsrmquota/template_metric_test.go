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

package fsrmquota

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestQuotaTemplateMetric(t *testing.T) {
	t.Parallel()

	c := New(nil)
	c.buildDescriptors()

	metrics := make(chan prometheus.Metric, 20)
	c.emitQuotas(metrics, []msftFSRMQuota{{Path: "C:\\data", Template: "Standard"}})
	close(metrics)

	found := false

	for metric := range metrics {
		if metric.Desc() != c.template {
			continue
		}

		var value dto.Metric
		require.NoError(t, metric.Write(&value))
		require.InDelta(t, 1, value.GetGauge().GetValue(), 1e-9)

		labels := make(map[string]string)
		for _, label := range value.GetLabel() {
			labels[label.GetName()] = label.GetValue()
		}

		require.Equal(t, map[string]string{"path": "C:\\data", "template": "Standard"}, labels)

		found = true
	}

	require.True(t, found)
}
