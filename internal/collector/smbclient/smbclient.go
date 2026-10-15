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

package smbclient

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name = "smbclient"
)

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

type Collector struct {
	config Config
	logger *slog.Logger

	perfDataCollector *pdh.Collector[perfDataCounterValues]
	perfDataObject    []perfDataCounterValues

	// unexpectedInstances holds instance names already logged as unparsable.
	unexpectedInstances map[string]struct{}

	// counters32 extends the 32-bit counters of each share.
	counters32 map[string]*shareCounters32
	generation uint64

	readBytesTotal                            *prometheus.Desc
	readBytesTransmittedViaSMBDirectTotal     *prometheus.Desc
	readRequestQueueSecsTotal                 *prometheus.Desc
	readRequestsTransmittedViaSMBDirectTotal  *prometheus.Desc
	readSecsTotal                             *prometheus.Desc
	readsTotal                                *prometheus.Desc
	turboIOReadsTotal                         *prometheus.Desc
	TurboIOWritesTotal                        *prometheus.Desc
	writeBytesTotal                           *prometheus.Desc
	writeBytesTransmittedViaSMBDirectTotal    *prometheus.Desc
	writeRequestQueueSecsTotal                *prometheus.Desc
	writeRequestsTransmittedViaSMBDirectTotal *prometheus.Desc
	writeSecsTotal                            *prometheus.Desc
	writesTotal                               *prometheus.Desc

	creditStallsTotal     *prometheus.Desc
	currentDataQueued     *prometheus.Desc
	dataBytesTotal        *prometheus.Desc
	dataRequestsTotal     *prometheus.Desc
	metadataRequestsTotal *prometheus.Desc
	requestQueueSecsTotal *prometheus.Desc
	requestSecs           *prometheus.Desc
}

// shareCounters32 holds the 32-bit counters of one share, extended beyond
// 2^32, and the 64-bit byte counter that shows a reconnect of the share.
type shareCounters32 struct {
	generation uint64
	restart    [numRestartSignals]float64
	counters   [numTimers32 + numCounts32]utils.Counter32
}

const (
	numTimers32       = 3
	numCounts32       = 9
	numRestartSignals = 4

	// restartBelow treats a drop of a request count from below 2^31 as a
	// share restart: a wrap from there needs 2^31 requests between two
	// scrapes. Timer sums grow by 10^7 per second for each outstanding
	// request, so they can wrap from anywhere and don't use it.
	restartBelow = 1 << 31

	staleGenerations = 3
)

// restartSignals returns 64-bit counters that only drop when the share
// restarts, for example after an idle disconnect and a reconnect.
func restartSignals(data *perfDataCounterValues) [numRestartSignals]float64 {
	return [numRestartSignals]float64{
		data.DataBytesPerSec,
		data.ReadBytesPerSec,
		data.WriteBytesPerSec,
		data.AvgDataQueueLength,
	}
}

// fields32 returns the 32-bit counters: first the PERF_AVERAGE_TIMER
// numerators, then the request counters (PERF_COUNTER_COUNTER).
func fields32(data *perfDataCounterValues) [numTimers32 + numCounts32]*float64 {
	return [numTimers32 + numCounts32]*float64{
		&data.AvgSecPerRead,
		&data.AvgSecPerWrite,
		&data.AvgSecPerDataRequest,
		&data.CreditStallsPerSec,
		&data.DataRequestsPerSec,
		&data.MetadataRequestsPerSec,
		&data.ReadRequestsTransmittedViaSMBDirectPerSec,
		&data.ReadRequestsPerSec,
		&data.TurboIOReadsPerSec,
		&data.TurboIOWritesPerSec,
		&data.WriteRequestsTransmittedViaSMBDirectPerSec,
		&data.WriteRequestsPerSec,
	}
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	c.perfDataCollector.Close()

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	c.logger = logger.With(slog.String("collector", Name))
	c.unexpectedInstances = make(map[string]struct{})
	c.counters32 = make(map[string]*shareCounters32)

	// desc creates a new prometheus description
	desc := func(metricName string, description string, labels []string) *prometheus.Desc {
		return prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, Name, metricName),
			description,
			labels,
			nil,
		)
	}

	c.requestQueueSecsTotal = desc("data_queue_seconds_total",
		"Seconds requests waited on queue on this share",
		[]string{"server", "share"},
	)
	c.readRequestQueueSecsTotal = desc("read_queue_seconds_total",
		"Seconds read requests waited on queue on this share",
		[]string{"server", "share"},
	)
	c.writeRequestQueueSecsTotal = desc("write_queue_seconds_total",
		"Seconds write requests waited on queue on this share",
		[]string{"server", "share"},
	)
	c.requestSecs = desc("request_seconds_total",
		"Seconds waiting for requests on this share",
		[]string{"server", "share"},
	)
	c.creditStallsTotal = desc("stalls_total",
		"The number of requests delayed based on insufficient credits on this share",
		[]string{"server", "share"},
	)
	c.currentDataQueued = desc("requests_queued",
		"The point in time number of requests outstanding on this share",
		[]string{"server", "share"},
	)
	c.dataBytesTotal = desc("data_bytes_total",
		"The bytes read or written on this share",
		[]string{"server", "share"},
	)
	c.dataRequestsTotal = desc("requests_total",
		"The requests on this share",
		[]string{"server", "share"},
	)
	c.metadataRequestsTotal = desc("metadata_requests_total",
		"The metadata requests on this share",
		[]string{"server", "share"},
	)
	c.readBytesTransmittedViaSMBDirectTotal = desc("read_bytes_via_smbdirect_total",
		"The bytes read from this share via RDMA direct placement",
		[]string{"server", "share"},
	)
	c.readBytesTotal = desc("read_bytes_total",
		"The bytes read on this share",
		[]string{"server", "share"},
	)
	c.readRequestsTransmittedViaSMBDirectTotal = desc("read_requests_via_smbdirect_total",
		"The read requests on this share via RDMA direct placement",
		[]string{"server", "share"},
	)
	c.readsTotal = desc("read_requests_total",
		"The read requests on this share",
		[]string{"server", "share"},
	)
	c.turboIOReadsTotal = desc("turbo_io_reads_total",
		"The read requests that go through Turbo I/O",
		[]string{"server", "share"},
	)
	c.TurboIOWritesTotal = desc("turbo_io_writes_total",
		"The write requests that go through Turbo I/O",
		[]string{"server", "share"},
	)
	c.writeBytesTransmittedViaSMBDirectTotal = desc("write_bytes_via_smbdirect_total",
		"The written bytes to this share via RDMA direct placement",
		[]string{"server", "share"},
	)
	c.writeBytesTotal = desc("write_bytes_total",
		"The bytes written on this share",
		[]string{"server", "share"},
	)
	c.writeRequestsTransmittedViaSMBDirectTotal = desc("write_requests_via_smbdirect_total",
		"The write requests to this share via RDMA direct placement",
		[]string{"server", "share"},
	)
	c.writesTotal = desc("write_requests_total",
		"The write requests on this share",
		[]string{"server", "share"},
	)
	c.readSecsTotal = desc("read_seconds_total",
		"Seconds waiting for read requests on this share",
		[]string{"server", "share"},
	)
	c.writeSecsTotal = desc("write_seconds_total",
		"Seconds waiting for write requests on this share",
		[]string{"server", "share"},
	)

	var err error

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](c.logger, pdh.CounterTypeRaw, "SMB Client Shares", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create SMB Client Shares collector: %w", err)
	}

	return nil
}

// Collect collects smb client metrics and sends them to prometheus.
func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect SMB Client Shares metrics: %w", err)
	}

	c.generation++

	for _, data := range c.perfDataObject {
		serverValue, shareValue, ok := parseInstanceName(data.Name)
		if !ok {
			if _, logged := c.unexpectedInstances[data.Name]; !logged {
				c.unexpectedInstances[data.Name] = struct{}{}

				c.logger.Warn("skipping SMB Client Shares instance with unexpected name",
					slog.String("instance", data.Name),
				)
			}

			continue
		}

		c.extendCounters32(&data)

		// Request time spent on queue. Convert from ticks to seconds.
		ch <- prometheus.MustNewConstMetric(
			c.requestQueueSecsTotal,
			prometheus.CounterValue,
			data.AvgDataQueueLength*pdh.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		// Read time spent on queue. Convert from ticks to seconds.
		ch <- prometheus.MustNewConstMetric(
			c.readRequestQueueSecsTotal,
			prometheus.CounterValue,
			data.AvgReadQueueLength*pdh.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readSecsTotal,
			prometheus.CounterValue,
			data.AvgSecPerRead*pdh.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeSecsTotal,
			prometheus.CounterValue,
			data.AvgSecPerWrite*pdh.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.requestSecs,
			prometheus.CounterValue,
			data.AvgSecPerDataRequest*pdh.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		// Write time spent on queue. Convert from  ticks to seconds.
		ch <- prometheus.MustNewConstMetric(
			c.writeRequestQueueSecsTotal,
			prometheus.CounterValue,
			data.AvgWriteQueueLength*pdh.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.creditStallsTotal,
			prometheus.CounterValue,
			data.CreditStallsPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.currentDataQueued,
			prometheus.GaugeValue,
			data.CurrentDataQueueLength,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataBytesTotal,
			prometheus.CounterValue,
			data.DataBytesPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataRequestsTotal,
			prometheus.CounterValue,
			data.DataRequestsPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.metadataRequestsTotal,
			prometheus.CounterValue,
			data.MetadataRequestsPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readBytesTransmittedViaSMBDirectTotal,
			prometheus.CounterValue,
			data.ReadBytesTransmittedViaSMBDirectPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readBytesTotal,
			prometheus.CounterValue,
			data.ReadBytesPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readRequestsTransmittedViaSMBDirectTotal,
			prometheus.CounterValue,
			data.ReadRequestsTransmittedViaSMBDirectPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readsTotal,
			prometheus.CounterValue,
			data.ReadRequestsPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.turboIOReadsTotal,
			prometheus.CounterValue,
			data.TurboIOReadsPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TurboIOWritesTotal,
			prometheus.CounterValue,
			data.TurboIOWritesPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeBytesTransmittedViaSMBDirectTotal,
			prometheus.CounterValue,
			data.WriteBytesTransmittedViaSMBDirectPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeBytesTotal,
			prometheus.CounterValue,
			data.WriteBytesPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeRequestsTransmittedViaSMBDirectTotal,
			prometheus.CounterValue,
			data.WriteRequestsTransmittedViaSMBDirectPerSec,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writesTotal,
			prometheus.CounterValue,
			data.WriteRequestsPerSec,
			serverValue, shareValue,
		)
	}

	c.pruneCounters32()

	return nil
}

// pruneCounters32 drops the state of shares missing from the last
// staleGenerations scrapes. A share missing from a single sample keeps its
// state, so its extended counters don't fall back to the raw value.
func (c *Collector) pruneCounters32() {
	for name, counters := range c.counters32 {
		if c.generation-counters.generation >= staleGenerations {
			delete(c.counters32, name)
		}
	}
}

// extendCounters32 replaces the raw 32-bit counters with values that keep
// counting past 2^32.
func (c *Collector) extendCounters32(data *perfDataCounterValues) {
	fields := fields32(data)

	signals := restartSignals(data)

	counters, ok := c.counters32[data.Name]
	if !ok {
		counters = &shareCounters32{restart: signals}

		for i, field := range fields {
			var below uint32
			if i >= numTimers32 {
				below = restartBelow
			}

			counters.counters[i] = utils.NewCounter32(*field, below)
		}

		c.counters32[data.Name] = counters
	}

	var reset bool

	for i, signal := range signals {
		reset = reset || signal < counters.restart[i]
	}

	counters.generation = c.generation
	counters.restart = signals

	for i, field := range fields {
		*field = counters.counters[i].Update(*field, reset)
	}
}

// parseInstanceName splits an SMB Client Shares instance name such as
// \\server\share into its server and share parts.
func parseInstanceName(name string) (string, string, bool) {
	parsed := strings.FieldsFunc(name, func(r rune) bool { return r == '\\' })
	if len(parsed) != 2 {
		return "", "", false
	}

	return parsed[0], parsed[1], true
}
