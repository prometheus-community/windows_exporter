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
	"strconv"
	"time"

	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Collector) describeClrMemory() {
	c.allocatedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrMemory+"_allocated_bytes_total"),
		"Displays the total number of bytes allocated on the garbage collection heap.",
		[]string{"process", "process_id"},
		nil,
	)
	c.finalizationSurvivors = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrMemory+"_finalization_survivors"),
		"Displays the number of garbage-collected objects that survive a collection because they are waiting to be finalized.",
		[]string{"process", "process_id"},
		nil,
	)
	c.heapSize = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrMemory+"_heap_size_bytes"),
		"Displays the maximum bytes that can be allocated; it does not indicate the current number of bytes allocated.",
		[]string{"process", "process_id", "area"},
		nil,
	)
	c.promotedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrMemory+"_promoted_bytes"),
		"Displays the bytes that were promoted from the generation to the next one during the last GC. Memory is promoted when it survives a garbage collection.",
		[]string{"process", "process_id", "area"},
		nil,
	)
	c.numberGCHandles = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrMemory+"_number_gc_handles"),
		"Displays the current number of garbage collection handles in use. Garbage collection handles are handles to resources external to the common language runtime and the managed environment.",
		[]string{"process", "process_id"},
		nil,
	)
	c.numberCollections = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrMemory+"_collections_total"),
		"Displays the number of times the generation objects are garbage collected since the application started.",
		[]string{"process", "process_id", "area"},
		nil,
	)
	c.numberInducedGC = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrMemory+"_induced_gc_total"),
		"Displays the peak number of times garbage collection was performed because of an explicit call to GC.Collect.",
		[]string{"process", "process_id"},
		nil,
	)
	c.numberOfPinnedObjects = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrMemory+"_number_pinned_objects"),
		"Displays the number of pinned objects encountered in the last garbage collection.",
		[]string{"process", "process_id"},
		nil,
	)
	c.numberOfSinkBlocksInUse = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrMemory+"_number_sink_blocksinuse"),
		"Displays the current number of synchronization blocks in use. Synchronization blocks are per-object data structures allocated for storing synchronization information. They hold weak references to managed objects and must be scanned by the garbage collector.",
		[]string{"process", "process_id"},
		nil,
	)
	c.numberTotalCommittedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrMemory+"_committed_bytes"),
		"Displays the amount of virtual memory, in bytes, currently committed by the garbage collector. Committed memory is the physical memory for which space has been reserved in the disk paging file.",
		[]string{"process", "process_id"},
		nil,
	)
	c.numberTotalReservedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrMemory+"_reserved_bytes"),
		"Displays the amount of virtual memory, in bytes, currently reserved by the garbage collector. Reserved memory is the virtual memory space reserved for the application when no disk or main memory pages have been used.",
		[]string{"process", "process_id"},
		nil,
	)
	c.timeInGC = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, collectorClrMemory+"_gc_time_percent"),
		"Displays the percentage of time that was spent performing a garbage collection in the last sample.",
		[]string{"process", "process_id"},
		nil,
	)
}

func (c *Collector) buildClrMemory() error {
	c.describeClrMemory()

	var err error

	c.perfClrMemory, err = newPerfCollector[perfDataClrMemory](c.logger, ".NET CLR Memory")

	return err
}

type perfDataClrMemory struct {
	Name                      string
	AllocatedBytesPersec      float64 `perfdata:"Allocated Bytes/sec"`
	FinalizationSurvivors     float64 `perfdata:"Finalization Survivors"`
	Gen0PromotedBytesPerSec   float64 `perfdata:"Gen 0 Promoted Bytes/Sec"`
	Gen0heapsize              float64 `perfdata:"Gen 0 heap size"`
	Gen1PromotedBytesPerSec   float64 `perfdata:"Gen 1 Promoted Bytes/Sec"`
	Gen1heapsize              float64 `perfdata:"Gen 1 heap size"`
	Gen2heapsize              float64 `perfdata:"Gen 2 heap size"`
	LargeObjectHeapsize       float64 `perfdata:"Large Object Heap size"`
	NumberGCHandles           float64 `perfdata:"# GC Handles"`
	NumberGen0Collections     float64 `perfdata:"# Gen 0 Collections"`
	NumberGen1Collections     float64 `perfdata:"# Gen 1 Collections"`
	NumberGen2Collections     float64 `perfdata:"# Gen 2 Collections"`
	NumberInducedGC           float64 `perfdata:"# Induced GC"`
	NumberTotalcommittedBytes float64 `perfdata:"# Total committed Bytes"`
	NumberTotalreservedBytes  float64 `perfdata:"# Total reserved Bytes"`
	NumberofPinnedObjects     float64 `perfdata:"# of Pinned Objects"`
	NumberofSinkBlocksinuse   float64 `perfdata:"# of Sink Blocks in use"`
	PercentTimeinGC           float64 `perfdata:"% Time in GC"`
	PercentTimeinGC_base      float64 `perfdata:"% Time in GC,secondvalue"`
	ProcessID                 float64 `perfdata:"Process ID"`
}

func (c *Collector) collectClrMemory(ch chan<- prometheus.Metric, _ time.Duration) error {
	var dst []perfDataClrMemory
	if err := c.perfClrMemory.Collect(&dst); err != nil {
		return fmt.Errorf("failed to collect .NET CLR Memory: %w", err)
	}

	for _, process := range dst {
		if process.Name == "_Global_" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.allocatedBytes,
			prometheus.CounterValue,
			float64(process.AllocatedBytesPersec),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
		)

		ch <- prometheus.MustNewConstMetric(
			c.finalizationSurvivors,
			prometheus.GaugeValue,
			float64(process.FinalizationSurvivors),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
		)

		ch <- prometheus.MustNewConstMetric(
			c.heapSize,
			prometheus.GaugeValue,
			float64(process.Gen0heapsize),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
			"Gen0",
		)

		ch <- prometheus.MustNewConstMetric(
			c.promotedBytes,
			prometheus.GaugeValue,
			float64(process.Gen0PromotedBytesPerSec),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
			"Gen0",
		)

		ch <- prometheus.MustNewConstMetric(
			c.heapSize,
			prometheus.GaugeValue,
			float64(process.Gen1heapsize),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
			"Gen1",
		)

		ch <- prometheus.MustNewConstMetric(
			c.promotedBytes,
			prometheus.GaugeValue,
			float64(process.Gen1PromotedBytesPerSec),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
			"Gen1",
		)

		ch <- prometheus.MustNewConstMetric(
			c.heapSize,
			prometheus.GaugeValue,
			float64(process.Gen2heapsize),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
			"Gen2",
		)

		ch <- prometheus.MustNewConstMetric(
			c.heapSize,
			prometheus.GaugeValue,
			float64(process.LargeObjectHeapsize),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
			"LOH",
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberGCHandles,
			prometheus.GaugeValue,
			float64(process.NumberGCHandles),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberCollections,
			prometheus.CounterValue,
			float64(process.NumberGen0Collections),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
			"Gen0",
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberCollections,
			prometheus.CounterValue,
			float64(process.NumberGen1Collections),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
			"Gen1",
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberCollections,
			prometheus.CounterValue,
			float64(process.NumberGen2Collections),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
			"Gen2",
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberInducedGC,
			prometheus.CounterValue,
			float64(process.NumberInducedGC),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfPinnedObjects,
			prometheus.GaugeValue,
			float64(process.NumberofPinnedObjects),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberOfSinkBlocksInUse,
			prometheus.GaugeValue,
			float64(process.NumberofSinkBlocksinuse),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberTotalCommittedBytes,
			prometheus.GaugeValue,
			float64(process.NumberTotalcommittedBytes),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
		)

		ch <- prometheus.MustNewConstMetric(
			c.numberTotalReservedBytes,
			prometheus.GaugeValue,
			float64(process.NumberTotalreservedBytes),
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
		)

		if process.PercentTimeinGC_base <= 0 {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.timeInGC,
			prometheus.GaugeValue,
			100*process.PercentTimeinGC/process.PercentTimeinGC_base,
			process.Name,
			strconv.FormatUint(uint64(process.ProcessID), 10),
		)
	}

	return nil
}
