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

package smbclient

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
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

	perfDataCollector *pdh.Collector

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
	return nil
}

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
	var err error

	c.perfDataCollector, err = pdh.NewCollector("SMB Client Shares", pdh.InstancesAll, []string{
		AvgDataQueueLength,
		AvgReadQueueLength,
		AvgSecPerRead,
		AvgSecPerWrite,
		AvgSecPerDataRequest,
		AvgWriteQueueLength,
		CreditStallsPerSec,
		CurrentDataQueueLength,
		DataBytesPerSec,
		DataRequestsPerSec,
		MetadataRequestsPerSec,
		ReadBytesTransmittedViaSMBDirectPerSec,
		ReadBytesPerSec,
		ReadRequestsTransmittedViaSMBDirectPerSec,
		ReadRequestsPerSec,
		TurboIOReadsPerSec,
		TurboIOWritesPerSec,
		WriteBytesTransmittedViaSMBDirectPerSec,
		WriteBytesPerSec,
		WriteRequestsTransmittedViaSMBDirectPerSec,
		WriteRequestsPerSec,
	}, false)
	if err != nil {
		return fmt.Errorf("failed to create SMB Client Shares collector: %w", err)
	}

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

	return nil
}

// Collect collects smb client metrics and sends them to prometheus.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect SMB Client Shares metrics: %w", err)
	}

	for name, data := range perfData {
		parsed := strings.FieldsFunc(name, func(r rune) bool { return r == '\\' })
		serverValue := parsed[0]
		shareValue := parsed[1]

		// Request time spent on queue. Convert from ticks to seconds.
		ch <- prometheus.MustNewConstMetric(
			c.requestQueueSecsTotal,
			prometheus.CounterValue,
			data[AvgDataQueueLength].FirstValue*pdh.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		// Read time spent on queue. Convert from ticks to seconds.
		ch <- prometheus.MustNewConstMetric(
			c.readRequestQueueSecsTotal,
			prometheus.CounterValue,
			data[AvgReadQueueLength].FirstValue*pdh.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readSecsTotal,
			prometheus.CounterValue,
			data[AvgSecPerRead].FirstValue*pdh.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeSecsTotal,
			prometheus.CounterValue,
			data[AvgSecPerWrite].FirstValue*pdh.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.requestSecs,
			prometheus.CounterValue,
			data[AvgSecPerDataRequest].FirstValue*pdh.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		// Write time spent on queue. Convert from  ticks to seconds.
		ch <- prometheus.MustNewConstMetric(
			c.writeRequestQueueSecsTotal,
			prometheus.CounterValue,
			data[AvgWriteQueueLength].FirstValue*pdh.TicksToSecondScaleFactor,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.creditStallsTotal,
			prometheus.CounterValue,
			data[CreditStallsPerSec].FirstValue,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.currentDataQueued,
			prometheus.GaugeValue,
			data[CurrentDataQueueLength].FirstValue,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataBytesTotal,
			prometheus.CounterValue,
			data[DataBytesPerSec].FirstValue,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.dataRequestsTotal,
			prometheus.CounterValue,
			data[DataRequestsPerSec].FirstValue,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.metadataRequestsTotal,
			prometheus.CounterValue,
			data[MetadataRequestsPerSec].FirstValue,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readBytesTransmittedViaSMBDirectTotal,
			prometheus.CounterValue,
			data[ReadBytesTransmittedViaSMBDirectPerSec].FirstValue,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readBytesTotal,
			prometheus.CounterValue,
			data[ReadBytesPerSec].FirstValue,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readRequestsTransmittedViaSMBDirectTotal,
			prometheus.CounterValue,
			data[ReadRequestsTransmittedViaSMBDirectPerSec].FirstValue,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readsTotal,
			prometheus.CounterValue,
			data[ReadRequestsPerSec].FirstValue,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.turboIOReadsTotal,
			prometheus.CounterValue,
			data[TurboIOReadsPerSec].FirstValue,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TurboIOWritesTotal,
			prometheus.CounterValue,
			data[TurboIOWritesPerSec].FirstValue,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeBytesTransmittedViaSMBDirectTotal,
			prometheus.CounterValue,
			data[WriteBytesTransmittedViaSMBDirectPerSec].FirstValue,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeBytesTotal,
			prometheus.CounterValue,
			data[WriteBytesPerSec].FirstValue,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeRequestsTransmittedViaSMBDirectTotal,
			prometheus.CounterValue,
			data[WriteRequestsTransmittedViaSMBDirectPerSec].FirstValue,
			serverValue, shareValue,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writesTotal,
			prometheus.CounterValue,
			data[WriteRequestsPerSec].FirstValue,
			serverValue, shareValue,
		)
	}

	return nil
}
