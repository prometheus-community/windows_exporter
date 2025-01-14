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

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

type collectorADAccessProcesses struct {
	perfDataCollectorADAccessProcesses *pdh.Collector
	perfDataObjectADAccessProcesses    []perfDataCounterValuesADAccessProcesses

	ldapReadTime                    *prometheus.Desc
	ldapSearchTime                  *prometheus.Desc
	ldapTimeoutErrorsPerSec         *prometheus.Desc
	ldapWriteTime                   *prometheus.Desc
	longRunningLDAPOperationsPerMin *prometheus.Desc
}

type perfDataCounterValuesADAccessProcesses struct {
	Name string

	LdapReadTime                    float64 `perfdata:"LDAP Read Time"`
	LdapSearchTime                  float64 `perfdata:"LDAP Search Time"`
	LdapWriteTime                   float64 `perfdata:"LDAP Write Time"`
	LdapTimeoutErrorsPerSec         float64 `perfdata:"LDAP Timeout Errors/sec"`
	LongRunningLDAPOperationsPerMin float64 `perfdata:"Long Running LDAP Operations/min"`
}

func (c *Collector) buildADAccessProcesses() error {
	var err error

	c.perfDataCollectorADAccessProcesses, err = pdh.NewCollector[perfDataCounterValuesADAccessProcesses](pdh.CounterTypeRaw, "MSExchange ADAccess Processes", pdh.InstancesAll)
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
	err := c.perfDataCollectorADAccessProcesses.Collect(&c.perfDataObjectADAccessProcesses)
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange ADAccess Processes metrics: %w", err)
	}

	labelUseCount := make(map[string]int)

	for _, data := range c.perfDataObjectADAccessProcesses {
		labelName := c.toLabelName(data.Name)

		// Since we're not including the PID suffix from the instance names in the label names, we get an occasional duplicate.
		// This seems to affect about 4 instances only of this object.
		labelUseCount[labelName]++
		if labelUseCount[labelName] > 1 {
			labelName = fmt.Sprintf("%s_%d", labelName, labelUseCount[labelName])
		}

		ch <- prometheus.MustNewConstMetric(
			c.ldapReadTime,
			prometheus.CounterValue,
			utils.MilliSecToSec(data.LdapReadTime),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ldapSearchTime,
			prometheus.CounterValue,
			utils.MilliSecToSec(data.LdapSearchTime),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ldapWriteTime,
			prometheus.CounterValue,
			utils.MilliSecToSec(data.LdapWriteTime),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ldapTimeoutErrorsPerSec,
			prometheus.CounterValue,
			data.LdapTimeoutErrorsPerSec,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.longRunningLDAPOperationsPerMin,
			prometheus.CounterValue,
			data.LongRunningLDAPOperationsPerMin*60,
			labelName,
		)
	}

	return nil
}
