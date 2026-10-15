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

package smb

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus-community/windows_exporter/internal/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "smb"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

type Collector struct {
	config Config

	perfDataCollector *pdh.Collector[perfDataCounterValues]
	perfDataObject    []perfDataCounterValues

	// counters32 extends the 32-bit request counters of each share.
	counters32 map[string]*shareCounters32
	generation uint64

	treeConnectCount     *prometheus.Desc
	currentOpenFileCount *prometheus.Desc
	receivedBytes        *prometheus.Desc
	writeRequests        *prometheus.Desc
	readRequests         *prometheus.Desc
	metadataRequests     *prometheus.Desc
	sentBytes            *prometheus.Desc
	filesOpened          *prometheus.Desc
}

// shareCounters32 holds the 32-bit counters of one share, extended beyond
// 2^32, and the 64-bit byte counters that show a restart of the share.
type shareCounters32 struct {
	generation    uint64
	receivedBytes float64
	sentBytes     float64
	readRequests  utils.Counter32
	writeRequests utils.Counter32
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
	c.currentOpenFileCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_shares_current_open_file_count"),
		"Current total count open files on the SMB Server Share",
		[]string{"share"},
		nil,
	)
	c.treeConnectCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_shares_tree_connect_count"),
		"Count of user connections to the SMB Server Share",
		[]string{"share"},
		nil,
	)
	c.receivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_shares_received_bytes_total"),
		"Received bytes on the SMB Server Share",
		[]string{"share"},
		nil,
	)
	c.writeRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_shares_write_requests_total"),
		"Writes requests on the SMB Server Share",
		[]string{"share"},
		nil,
	)
	c.readRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_shares_read_requests_total"),
		"Read requests on the SMB Server Share",
		[]string{"share"},
		nil,
	)
	c.metadataRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_shares_metadata_requests_total"),
		"Metadata requests on the SMB Server Share",
		[]string{"share"},
		nil,
	)
	c.sentBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_shares_sent_bytes_total"),
		"Sent bytes on the SMB Server Share",
		[]string{"share"},
		nil,
	)
	c.filesOpened = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_shares_files_opened_total"),
		"Files opened on the SMB Server Share",
		[]string{"share"},
		nil,
	)

	c.counters32 = make(map[string]*shareCounters32)

	var err error

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](logger.With(slog.String("collector", Name)), pdh.CounterTypeRaw, "SMB Server Shares", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create SMB Server Shares collector: %w", err)
	}

	return nil
}

// Collect collects smb metrics and sends them to prometheus.
func (c *Collector) Collect(ch chan<- prometheus.Metric, _ time.Duration) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect SMB Server Shares metrics: %w", err)
	}

	c.generation++

	for _, data := range c.perfDataObject {
		c.extendCounters32(&data)

		ch <- prometheus.MustNewConstMetric(
			c.currentOpenFileCount,
			prometheus.GaugeValue,
			data.CurrentOpenFileCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.treeConnectCount,
			prometheus.GaugeValue,
			data.TreeConnectCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.receivedBytes,
			prometheus.CounterValue,
			data.ReceivedBytes,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeRequests,
			prometheus.CounterValue,
			data.WriteRequests,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readRequests,
			prometheus.CounterValue,
			data.ReadRequests,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.metadataRequests,
			prometheus.CounterValue,
			data.MetadataRequests,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sentBytes,
			prometheus.CounterValue,
			data.SentBytes,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.filesOpened,
			prometheus.CounterValue,
			data.FilesOpened,
			data.Name,
		)
	}

	for name, counters := range c.counters32 {
		if counters.generation != c.generation {
			delete(c.counters32, name)
		}
	}

	return nil
}

// extendCounters32 replaces the raw 32-bit request counters with values that
// keep counting past 2^32. Read Requests/sec and Write Requests/sec are
// PERF_COUNTER_COUNTER (32-bit), the byte counters are 64-bit.
func (c *Collector) extendCounters32(data *perfDataCounterValues) {
	counters, ok := c.counters32[data.Name]
	if !ok {
		counters = &shareCounters32{
			readRequests:  utils.NewCounter32(data.ReadRequests),
			writeRequests: utils.NewCounter32(data.WriteRequests),
		}
		c.counters32[data.Name] = counters
	}

	reset := data.ReceivedBytes < counters.receivedBytes || data.SentBytes < counters.sentBytes

	counters.generation = c.generation
	counters.receivedBytes = data.ReceivedBytes
	counters.sentBytes = data.SentBytes

	data.ReadRequests = counters.readRequests.Update(data.ReadRequests, reset)
	data.WriteRequests = counters.writeRequests.Update(data.WriteRequests, reset)
}
