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

package smb

import (
	"fmt"
	"log/slog"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "smb"

type Config struct{}

//nolint:gochecknoglobals
var ConfigDefaults = Config{}

type Collector struct {
	config Config

	perfDataCollector *pdh.Collector
	perfDataObject    []perfDataCounterValues

	treeConnectCount     *prometheus.Desc
	currentOpenFileCount *prometheus.Desc
	receivedBytes        *prometheus.Desc
	writeRequests        *prometheus.Desc
	readRequests         *prometheus.Desc
	metadataRequests     *prometheus.Desc
	sentBytes            *prometheus.Desc
	filesOpened          *prometheus.Desc
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

func (c *Collector) Build(_ *slog.Logger, _ *mi.Session) error {
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
		prometheus.BuildFQName(types.Namespace, Name, "server_shares_write_requests_count_total"),
		"Writes requests on the SMB Server Share",
		[]string{"share"},
		nil,
	)
	c.readRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_shares_read_requests_count_total"),
		"Read requests on the SMB Server Share",
		[]string{"share"},
		nil,
	)
	c.metadataRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "server_shares_metadata_requests_count_total"),
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
		prometheus.BuildFQName(types.Namespace, Name, "server_shares_files_opened_count_total"),
		"Files opened on the SMB Server Share",
		[]string{"share"},
		nil,
	)

	var err error

	c.perfDataCollector, err = pdh.NewCollector[perfDataCounterValues](pdh.CounterTypeRaw, "SMB Server Shares", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create SMB Server Shares collector: %w", err)
	}

	return nil
}

// Collect collects smb metrics and sends them to prometheus.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollector.Collect(&c.perfDataObject)
	if err != nil {
		return fmt.Errorf("failed to collect SMB Server Shares metrics: %w", err)
	}

	for _, data := range c.perfDataObject {
		ch <- prometheus.MustNewConstMetric(
			c.currentOpenFileCount,
			prometheus.CounterValue,
			data.CurrentOpenFileCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.treeConnectCount,
			prometheus.CounterValue,
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

	return nil
}
