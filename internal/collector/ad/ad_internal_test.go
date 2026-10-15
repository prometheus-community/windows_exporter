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

package ad

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestCollectMetrics(t *testing.T) {
	t.Parallel()

	// Every counter gets a distinct value, so a metric that reads the wrong
	// field shows up as a wrong value.
	data := perfDataCounterValues{
		AtqEstimatedQueueDelay:                                  250,
		AtqRequestLatency:                                       1500,
		DsPercentSearchesFromLSA:                                9,
		DsPercentWritesFromLSA:                                  12,
		DsPercentReadsFromLSA:                                   70,
		DsSecurityDescriptorSubOperationsPerSec:                 17,
		DsSecurityDescriptorPropagationsEvents:                  3,
		DsSecurityDescriptorPropagatorAverageExclusionTime:      42,
		DsSecurityDescriptorPropagatorRuntimeQueue:              5,
		LdapBindTime:                                            20,
		SamAccountGroupEvaluationLatency:                        30,
		SamResourceGroupEvaluationLatency:                       40,
		SamMachineCreationAttemptsPerSec:                        7,
		SamSuccessfulComputerCreationsPerSecIncludesAllRequests: 4,
		SamUserCreationAttemptsPerSec:                           11,
		SamSuccessfulUserCreationsPerSec:                        10,
	}

	c := New(nil)
	c.buildDescriptors()

	registry := prometheus.NewPedanticRegistry()
	require.NoError(t, registry.Register(metricCollector{collector: c, data: data}))

	families, err := registry.Gather()
	require.NoError(t, err)

	byName := make(map[string]*dto.MetricFamily, len(families))
	for _, family := range families {
		require.NotEmpty(t, family.GetHelp(), "metric %s has no HELP", family.GetName())
		byName[family.GetName()] = family
	}

	for _, tc := range []struct {
		name       string
		labels     map[string]string
		metricType dto.MetricType
		value      float64
	}{
		{
			name:       "windows_ad_directory_operations_total",
			labels:     map[string]string{"operation": "write", "origin": "local_security_authority"},
			metricType: dto.MetricType_COUNTER,
			value:      12,
		},
		{
			name:       "windows_ad_directory_operations_total",
			labels:     map[string]string{"operation": "search", "origin": "local_security_authority"},
			metricType: dto.MetricType_COUNTER,
			value:      9,
		},
		{
			name:       "windows_ad_directory_operations_total",
			labels:     map[string]string{"operation": "read", "origin": "local_security_authority"},
			metricType: dto.MetricType_COUNTER,
			value:      70,
		},
		{name: "windows_ad_sam_computer_creation_requests_total", metricType: dto.MetricType_COUNTER, value: 7},
		{name: "windows_ad_sam_computer_creation_successful_requests_total", metricType: dto.MetricType_COUNTER, value: 4},
		{name: "windows_ad_sam_user_creation_requests_total", metricType: dto.MetricType_COUNTER, value: 11},
		{name: "windows_ad_sam_user_creation_successful_requests_total", metricType: dto.MetricType_COUNTER, value: 10},
		{name: "windows_ad_security_descriptor_propagation_suboperations_total", metricType: dto.MetricType_COUNTER, value: 17},
		{name: "windows_ad_security_descriptor_propagation_events_queued", metricType: dto.MetricType_GAUGE, value: 3},
		{name: "windows_ad_security_descriptor_propagation_average_exclusion_time", metricType: dto.MetricType_GAUGE, value: 42},
		{name: "windows_ad_security_descriptor_propagation_items_queued", metricType: dto.MetricType_GAUGE, value: 5},
		{name: "windows_ad_atq_estimated_delay_seconds", metricType: dto.MetricType_GAUGE, value: 0.25},
		{name: "windows_ad_atq_average_request_latency_seconds", metricType: dto.MetricType_GAUGE, value: 1.5},
		{name: "windows_ad_ldap_last_bind_time_seconds", metricType: dto.MetricType_GAUGE, value: 0.02},
		{
			name:       "windows_ad_sam_group_evaluation_latency_seconds",
			labels:     map[string]string{"evaluation_type": "account_group"},
			metricType: dto.MetricType_GAUGE,
			value:      0.03,
		},
		{
			name:       "windows_ad_sam_group_evaluation_latency_seconds",
			labels:     map[string]string{"evaluation_type": "resource_group"},
			metricType: dto.MetricType_GAUGE,
			value:      0.04,
		},
	} {
		family, ok := byName[tc.name]
		require.True(t, ok, "metric %s not published", tc.name)
		require.Equal(t, tc.metricType, family.GetType(), tc.name)

		metric := findMetric(t, family, tc.labels)

		actual := metric.GetGauge().GetValue()
		if tc.metricType == dto.MetricType_COUNTER {
			actual = metric.GetCounter().GetValue()
		}

		require.InDelta(t, tc.value, actual, 1e-9, "%s%v", tc.name, tc.labels)
	}

	for _, name := range []string{
		"windows_ad_atq_average_request_latency",
		"windows_ad_sam_group_evaluation_latency",
		"windows_ad_security_descriptor_propagation_access_wait_total_seconds",
		"windows_ad_security_descriptor_propagation_events_total",
		"windows_ad_security_descriptor_propagation_items_queued_total",
	} {
		require.NotContains(t, byName, name, "renamed metric %s is still published", name)
	}
}

func findMetric(t *testing.T, family *dto.MetricFamily, labels map[string]string) *dto.Metric {
	t.Helper()

	for _, metric := range family.GetMetric() {
		matched := 0

		for _, label := range metric.GetLabel() {
			if value, ok := labels[label.GetName()]; ok && value == label.GetValue() {
				matched++
			}
		}

		if matched == len(labels) {
			return metric
		}
	}

	require.FailNow(t, "series not found", "%s%v", family.GetName(), labels)

	return nil
}

type metricCollector struct {
	collector *Collector
	data      perfDataCounterValues
}

func (c metricCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c metricCollector) Collect(ch chan<- prometheus.Metric) {
	c.collector.collectMetrics(ch, c.data)
}
