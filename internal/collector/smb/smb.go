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
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "smb"

type Config struct{}

var ConfigDefaults = Config{}

type Collector struct {
	config Config

	perfDataCollector *perfdata.Collector

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
	var err error

	c.perfDataCollector, err = perfdata.NewCollector("SMB Server Shares", perfdata.InstancesAll, []string{
		currentOpenFileCount,
		treeConnectCount,
		receivedBytes,
		writeRequests,
		readRequests,
		metadataRequests,
		sentBytes,
		filesOpened,
	})
	if err != nil {
		return fmt.Errorf("failed to create SMB Server Shares collector: %w", err)
	}

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
		prometheus.BuildFQName(types.Namespace, Name, "server_shares_filed_opened_count_total"),
		"Files opened on the SMB Server Share",
		[]string{"share"},
		nil,
	)

	return nil
}

// Collect collects smb metrics and sends them to prometheus.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect SMB Server Shares metrics: %w", err)
	}

	for share, data := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.currentOpenFileCount,
			prometheus.CounterValue,
			data[currentOpenFileCount].FirstValue,
			share,
		)

		ch <- prometheus.MustNewConstMetric(
			c.treeConnectCount,
			prometheus.CounterValue,
			data[treeConnectCount].FirstValue,
			share,
		)

		ch <- prometheus.MustNewConstMetric(
			c.receivedBytes,
			prometheus.CounterValue,
			data[receivedBytes].FirstValue,
			share,
		)

		ch <- prometheus.MustNewConstMetric(
			c.writeRequests,
			prometheus.CounterValue,
			data[writeRequests].FirstValue,
			share,
		)

		ch <- prometheus.MustNewConstMetric(
			c.readRequests,
			prometheus.CounterValue,
			data[readRequests].FirstValue,
			share,
		)

		ch <- prometheus.MustNewConstMetric(
			c.metadataRequests,
			prometheus.CounterValue,
			data[metadataRequests].FirstValue,
			share,
		)

		ch <- prometheus.MustNewConstMetric(
			c.sentBytes,
			prometheus.CounterValue,
			data[sentBytes].FirstValue,
			share,
		)

		ch <- prometheus.MustNewConstMetric(
			c.filesOpened,
			prometheus.CounterValue,
			data[filesOpened].FirstValue,
			share,
		)
	}

	return nil
}
