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

package exchange

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	types "github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	ldapReadTime                    = "LDAP Read Time"
	ldapSearchTime                  = "LDAP Search Time"
	ldapWriteTime                   = "LDAP Write Time"
	ldapTimeoutErrorsPerSec         = "LDAP Timeout Errors/sec"
	longRunningLDAPOperationsPerMin = "Long Running LDAP Operations/min"
)

func (c *Collector) buildADAccessProcesses() error {
	counters := []string{
		ldapReadTime,
		ldapSearchTime,
		ldapWriteTime,
		ldapTimeoutErrorsPerSec,
		longRunningLDAPOperationsPerMin,
	}

	var err error

	c.perfDataCollectorADAccessProcesses, err = perfdata.NewCollector("MSExchange ADAccess Processes", perfdata.InstancesAll, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange ADAccess Processes collector: %w", err)
	}

	c.ldapReadTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_read_time_sec"),
		"Time (sec) to send an LDAP read request and receive a response",
		[]string{"name"},
		nil,
	)
	c.ldapSearchTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_search_time_sec"),
		"Time (sec) to send an LDAP search request and receive a response",
		[]string{"name"},
		nil,
	)
	c.ldapWriteTime = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_write_time_sec"),
		"Time (sec) to send an LDAP Add/Modify/Delete request and receive a response",
		[]string{"name"},
		nil,
	)
	c.ldapTimeoutErrorsPerSec = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_timeout_errors_total"),
		"Total number of LDAP timeout errors",
		[]string{"name"},
		nil,
	)
	c.longRunningLDAPOperationsPerMin = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "ldap_long_running_ops_per_sec"),
		"Long Running LDAP operations per second",
		[]string{"name"},
		nil,
	)

	return nil
}

func (c *Collector) collectADAccessProcesses(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorADAccessProcesses.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange ADAccess Processes metrics: %w", err)
	}

	if len(perfData) == 0 {
		return fmt.Errorf("failed to collect MSExchange ADAccess Processes metrics: %w", types.ErrNoData)
	}

	labelUseCount := make(map[string]int)

	for name, data := range perfData {
		labelName := c.toLabelName(name)

		// Since we're not including the PID suffix from the instance names in the label names, we get an occasional duplicate.
		// This seems to affect about 4 instances only of this object.
		labelUseCount[labelName]++
		if labelUseCount[labelName] > 1 {
			labelName = fmt.Sprintf("%s_%d", labelName, labelUseCount[labelName])
		}

		ch <- prometheus.MustNewConstMetric(
			c.ldapReadTime,
			prometheus.CounterValue,
			c.msToSec(data[ldapReadTime].FirstValue),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ldapSearchTime,
			prometheus.CounterValue,
			c.msToSec(data[ldapSearchTime].FirstValue),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ldapWriteTime,
			prometheus.CounterValue,
			c.msToSec(data[ldapWriteTime].FirstValue),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ldapTimeoutErrorsPerSec,
			prometheus.CounterValue,
			data[ldapTimeoutErrorsPerSec].FirstValue,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.longRunningLDAPOperationsPerMin,
			prometheus.CounterValue,
			data[longRunningLDAPOperationsPerMin].FirstValue*60,
			labelName,
		)
	}

	return nil
}
