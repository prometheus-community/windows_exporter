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

package adcs

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestCollectMetrics(t *testing.T) {
	t.Parallel()

	values := perfDataCounterValues{
		Name:                                         "WebServer",
		RequestsPerSecond:                            17,
		RequestProcessingTime:                        1500,
		RetrievalsPerSecond:                          19,
		RetrievalProcessingTime:                      2250,
		FailedRequestsPerSecond:                      3,
		IssuedRequestsPerSecond:                      11,
		PendingRequestsPerSecond:                     2,
		RequestCryptographicSigningTime:              500,
		RequestPolicyModuleProcessingTime:            750,
		ChallengeResponsesPerSecond:                  13,
		ChallengeResponseProcessingTime:              1250,
		SignedCertificateTimestampListsPerSecond:     23,
		SignedCertificateTimestampListProcessingTime: 1750,
	}
	expected := map[string]struct {
		metricType dto.MetricType
		value      float64
	}{
		"windows_adcs_requests_total":                                            {metricType: dto.MetricType_COUNTER, value: 17},
		"windows_adcs_request_processing_time_seconds":                           {metricType: dto.MetricType_GAUGE, value: 1.5},
		"windows_adcs_retrievals_total":                                          {metricType: dto.MetricType_COUNTER, value: 19},
		"windows_adcs_retrievals_processing_time_seconds":                        {metricType: dto.MetricType_GAUGE, value: 2.25},
		"windows_adcs_failed_requests_total":                                     {metricType: dto.MetricType_COUNTER, value: 3},
		"windows_adcs_issued_requests_total":                                     {metricType: dto.MetricType_COUNTER, value: 11},
		"windows_adcs_pending_requests_total":                                    {metricType: dto.MetricType_COUNTER, value: 2},
		"windows_adcs_request_cryptographic_signing_time_seconds":                {metricType: dto.MetricType_GAUGE, value: 0.5},
		"windows_adcs_request_policy_module_processing_time_seconds":             {metricType: dto.MetricType_GAUGE, value: 0.75},
		"windows_adcs_challenge_responses_total":                                 {metricType: dto.MetricType_COUNTER, value: 13},
		"windows_adcs_challenge_response_processing_time_seconds":                {metricType: dto.MetricType_GAUGE, value: 1.25},
		"windows_adcs_signed_certificate_timestamp_lists_total":                  {metricType: dto.MetricType_COUNTER, value: 23},
		"windows_adcs_signed_certificate_timestamp_list_processing_time_seconds": {metricType: dto.MetricType_GAUGE, value: 1.75},
	}

	for _, tc := range []struct {
		name string
		rows []perfDataCounterValues
	}{
		{name: "no templates"},
		{name: "zero counters", rows: []perfDataCounterValues{{Name: "ZeroTemplate"}}},
		{name: "independent templates", rows: []perfDataCounterValues{values, {Name: "ZeroTemplate"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := New(nil)
			c.buildDescriptors()

			registry := prometheus.NewPedanticRegistry()
			require.NoError(t, registry.Register(metricCollector{collector: c, rows: tc.rows}))
			families, err := registry.Gather()
			require.NoError(t, err)

			if len(tc.rows) == 0 {
				require.Empty(t, families)

				return
			}

			require.Len(t, families, len(expected))

			for _, family := range families {
				want, ok := expected[family.GetName()]
				require.True(t, ok, "unexpected metric %s", family.GetName())
				require.Equal(t, want.metricType, family.GetType())
				require.Len(t, family.GetMetric(), len(tc.rows))
				templates := make(map[string]bool, len(tc.rows))

				for _, metric := range family.GetMetric() {
					require.Len(t, metric.GetLabel(), 1)
					label := metric.GetLabel()[0]
					require.Equal(t, "cert_template", label.GetName())
					require.False(t, templates[label.GetValue()], "duplicate template sample")
					templates[label.GetValue()] = true

					expectedValue := 0.0
					if label.GetValue() == values.Name {
						expectedValue = want.value
					} else {
						require.Equal(t, "ZeroTemplate", label.GetValue())
					}

					actual := metric.GetGauge().GetValue()
					if want.metricType == dto.MetricType_COUNTER {
						actual = metric.GetCounter().GetValue()
					}

					require.InDelta(t, expectedValue, actual, 0, "%s for template %s", family.GetName(), label.GetValue())
				}
			}
		})
	}
}

type metricCollector struct {
	collector *Collector
	rows      []perfDataCounterValues
}

func (c metricCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c metricCollector) Collect(ch chan<- prometheus.Metric) {
	c.collector.collectMetrics(ch, c.rows)
}
