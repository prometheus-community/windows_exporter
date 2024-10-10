package exchange

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	v1 "github.com/prometheus-community/windows_exporter/internal/perfdata/v1"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	ldapReadTime                    = "LDAP Read Time"
	ldapSearchTime                  = "LDAP Search Time"
	ldapWriteTime                   = "LDAP Write Time"
	ldapTimeoutErrorsPerSec         = "LDAP Timeout Errors/sec"
	longRunningLDAPOperationsPerMin = "Long Running LDAP Operations/min"
)

// Perflib: [19108] MSExchange ADAccess Processes.
type perflibADAccessProcesses struct {
	Name string

	LDAPReadTime                    float64 `perflib:"LDAP Read Time"`
	LDAPSearchTime                  float64 `perflib:"LDAP Search Time"`
	LDAPWriteTime                   float64 `perflib:"LDAP Write Time"`
	LDAPTimeoutErrorsPerSec         float64 `perflib:"LDAP Timeout Errors/sec"`
	LongRunningLDAPOperationsPerMin float64 `perflib:"Long Running LDAP Operations/min"`
}

func (c *Collector) buildADAccessProcesses() error {
	counters := []string{
		ldapReadTime,
		ldapSearchTime,
		ldapWriteTime,
		ldapTimeoutErrorsPerSec,
		longRunningLDAPOperationsPerMin,
	}

	var err error

	c.perfDataCollectorADAccessProcesses, err = perfdata.NewCollector(perfdata.V1, "MSExchange ADAccess Processes", perfdata.AllInstances, counters)
	if err != nil {
		return fmt.Errorf("failed to create MSExchange ADAccess Processes collector: %w", err)
	}

	return nil
}

func (c *Collector) collectADAccessProcesses(ctx *types.ScrapeContext, logger *slog.Logger, ch chan<- prometheus.Metric) error {
	var data []perflibADAccessProcesses

	if err := v1.UnmarshalObject(ctx.PerfObjects["MSExchange ADAccess Processes"], &data, logger); err != nil {
		return err
	}

	labelUseCount := make(map[string]int)

	for _, proc := range data {
		labelName := c.toLabelName(proc.Name)
		if strings.HasSuffix(labelName, "_total") {
			continue
		}

		// Since we're not including the PID suffix from the instance names in the label names, we get an occasional duplicate.
		// This seems to affect about 4 instances only of this object.
		labelUseCount[labelName]++
		if labelUseCount[labelName] > 1 {
			labelName = fmt.Sprintf("%s_%d", labelName, labelUseCount[labelName])
		}
		ch <- prometheus.MustNewConstMetric(
			c.ldapReadTime,
			prometheus.CounterValue,
			c.msToSec(proc.LDAPReadTime),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ldapSearchTime,
			prometheus.CounterValue,
			c.msToSec(proc.LDAPSearchTime),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ldapWriteTime,
			prometheus.CounterValue,
			c.msToSec(proc.LDAPWriteTime),
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.ldapTimeoutErrorsPerSec,
			prometheus.CounterValue,
			proc.LDAPTimeoutErrorsPerSec,
			labelName,
		)
		ch <- prometheus.MustNewConstMetric(
			c.longRunningLDAPOperationsPerMin,
			prometheus.CounterValue,
			proc.LongRunningLDAPOperationsPerMin*60,
			labelName,
		)
	}

	return nil
}

func (c *Collector) collectPDHADAccessProcesses(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorADAccessProcesses.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect MSExchange ADAccess Processes metrics: %w", err)
	}

	if len(perfData) == 0 {
		return errors.New("perflib query for MSExchange ADAccess Processes returned empty result set")
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
