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

package netframework

import (
	"fmt"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Collector) describeClrJIT() {
	c.numberOfMethodsJitted = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrJIT+"_jit_methods_total"),
		"Displays the total number of methods JIT-compiled since the application started. This counter does not include pre-JIT-compiled methods.",
		[]string{"process"},
		nil,
	)
	c.timeInJit = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrJIT+"_jit_time_percent"),
		"Displays the percentage of time spent in JIT compilation. This counter is updated at the end of every JIT compilation phase. A JIT compilation phase occurs when a method and its dependencies are compiled.",
		[]string{"process"},
		nil,
	)
	c.standardJitFailures = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrJIT+"_jit_standard_failures_total"),
		"Displays the peak number of methods the JIT compiler has failed to compile since the application started. This failure can occur if the MSIL cannot be verified or if there is an internal error in the JIT compiler.",
		[]string{"process"},
		nil,
	)
	c.totalNumberOfILBytesJitted = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrJIT+"_jit_il_bytes_total"),
		"Displays the total number of Microsoft intermediate language (MSIL) bytes compiled by the just-in-time (JIT) compiler since the application started",
		[]string{"process"},
		nil,
	)
}

func (c *Collector) buildClrJIT() error {
	c.describeClrJIT()

	var err error

	c.perfClrJIT, err = newPerfCollector[perfDataClrJIT](c.logger, ".NET CLR Jit")

	return err
}

type perfDataClrJIT struct {
	Name                       string
	NumberofMethodsJitted      float64 `perfdata:"# of Methods JITted"`
	PercentTimeinJit           float64 `perfdata:"% Time in Jit"`
	StandardJitFailures        float64 `perfdata:"Standard Jit Failures"`
	TotalNumberofILBytesJitted float64 `perfdata:"Total # of IL Bytes Jitted"`
}

func (c *Collector) collectClrJIT(ch chan<- prometheus.Metric, _ time.Duration) error {
	var dst []perfDataClrJIT
	if err := c.perfClrJIT.Collect(&dst); err != nil {
		return fmt.Errorf("failed to collect .NET CLR Jit: %w", err)
	}

	for _, process := range dst {
		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.numberOfMethodsJitted,
			prometheus.CounterValue,
			float64(process.NumberofMethodsJitted),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.timeInJit,
			prometheus.GaugeValue,
			float64(process.PercentTimeinJit)/c.perfFrequency,
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.standardJitFailures,
			prometheus.GaugeValue,
			float64(process.StandardJitFailures),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.totalNumberOfILBytesJitted,
			prometheus.CounterValue,
			float64(process.TotalNumberofILBytesJitted),
			process.Name,
		)
	}

	return nil
}
