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

func (c *Collector) describeClrLocksAndThreads() {
	c.currentQueueLength = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrLocksAndThreads+"_current_queue_length"),
		"Displays the total number of threads that are currently waiting to acquire a managed lock in the application.",
		[]string{"process"},
		nil,
	)
	c.numberOfCurrentLogicalThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrLocksAndThreads+"_current_logical_threads"),
		"Displays the number of current managed thread objects in the application. This counter maintains the count of both running and stopped threads. ",
		[]string{"process"},
		nil,
	)
	c.numberOfCurrentPhysicalThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrLocksAndThreads+"_physical_threads_current"),
		"Displays the number of native operating system threads created and owned by the common language runtime to act as underlying threads for managed thread objects. This counter's value does not include the threads used by the runtime in its internal operations; it is a subset of the threads in the operating system process.",
		[]string{"process"},
		nil,
	)
	c.numberOfCurrentRecognizedThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrLocksAndThreads+"_recognized_threads_current"),
		"Displays the number of threads that are currently recognized by the runtime. These threads are associated with a corresponding managed thread object. The runtime does not create these threads, but they have run inside the runtime at least once.",
		[]string{"process"},
		nil,
	)
	c.numberOfTotalRecognizedThreads = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrLocksAndThreads+"_recognized_threads_total"),
		"Displays the total number of threads that have been recognized by the runtime since the application started. These threads are associated with a corresponding managed thread object. The runtime does not create these threads, but they have run inside the runtime at least once.",
		[]string{"process"},
		nil,
	)
	c.queueLengthPeak = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrLocksAndThreads+"_queue_length_total"),
		"Displays the total number of threads that waited to acquire a managed lock since the application started.",
		[]string{"process"},
		nil,
	)
	c.totalNumberOfContentions = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrLocksAndThreads+"_contentions_total"),
		"Displays the total number of times that threads in the runtime have attempted to acquire a managed lock unsuccessfully.",
		[]string{"process"},
		nil,
	)
}

func (c *Collector) buildClrLocksAndThreads() error {
	c.describeClrLocksAndThreads()

	var err error

	c.perfClrLocksAndThreads, err = newPerfCollector[perfDataClrLocksAndThreads](c.logger, ".NET CLR LocksAndThreads")

	return err
}

type perfDataClrLocksAndThreads struct {
	Name                             string
	CurrentQueueLength               float64 `perfdata:"Current Queue Length"`
	NumberofcurrentlogicalThreads    float64 `perfdata:"# of current logical Threads"`
	NumberofcurrentphysicalThreads   float64 `perfdata:"# of current physical Threads"`
	Numberofcurrentrecognizedthreads float64 `perfdata:"# of current recognized threads"`
	Numberoftotalrecognizedthreads   float64 `perfdata:"# of total recognized threads"`
	QueueLengthPeak                  float64 `perfdata:"Queue Length Peak"`
	TotalNumberofContentions         float64 `perfdata:"Total # of Contentions"`
}

func (c *Collector) collectClrLocksAndThreads(ch chan<- prometheus.Metric, _ time.Duration) error {
	var dst []perfDataClrLocksAndThreads
	if err := c.perfClrLocksAndThreads.Collect(&dst); err != nil {
		return fmt.Errorf("failed to collect .NET CLR LocksAndThreads: %w", err)
	}

	for _, process := range dst {
		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.currentQueueLength,
			prometheus.GaugeValue,
			float64(process.CurrentQueueLength),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfCurrentLogicalThreads,
			prometheus.GaugeValue,
			float64(process.NumberofcurrentlogicalThreads),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfCurrentPhysicalThreads,
			prometheus.GaugeValue,
			float64(process.NumberofcurrentphysicalThreads),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfCurrentRecognizedThreads,
			prometheus.GaugeValue,
			float64(process.Numberofcurrentrecognizedthreads),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfTotalRecognizedThreads,
			prometheus.CounterValue,
			float64(process.Numberoftotalrecognizedthreads),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.queueLengthPeak,
			prometheus.CounterValue,
			float64(process.QueueLengthPeak),
			process.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.totalNumberOfContentions,
			prometheus.CounterValue,
			float64(process.TotalNumberofContentions),
			process.Name,
		)
	}

	return nil
}
