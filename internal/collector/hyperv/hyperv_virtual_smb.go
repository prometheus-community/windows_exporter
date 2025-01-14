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

package hyperv

import (
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorVirtualSMB Hyper-V Virtual SMB metrics
type collectorVirtualSMB struct {
	perfDataCollectorVirtualSMB *pdh.Collector
	perfDataObjectVirtualSMB    []perfDataCounterValuesVirtualSMB

	virtualSMBDirectMappedSections   *prometheus.Desc // \Hyper-V Virtual SMB(*)\Direct-Mapped Sections
	virtualSMBDirectMappedPages      *prometheus.Desc // \Hyper-V Virtual SMB(*)\Direct-Mapped Pages
	virtualSMBWriteBytesRDMA         *prometheus.Desc // \Hyper-V Virtual SMB(*)\Write Bytes/sec (RDMA)
	virtualSMBWriteBytes             *prometheus.Desc // \Hyper-V Virtual SMB(*)\Write Bytes/sec
	virtualSMBReadBytesRDMA          *prometheus.Desc // \Hyper-V Virtual SMB(*)\Read Bytes/sec (RDMA)
	virtualSMBReadBytes              *prometheus.Desc // \Hyper-V Virtual SMB(*)\Read Bytes/sec
	virtualSMBFlushRequests          *prometheus.Desc // \Hyper-V Virtual SMB(*)\Flush Requests/sec
	virtualSMBWriteRequestsRDMA      *prometheus.Desc // \Hyper-V Virtual SMB(*)\Write Requests/sec (RDMA)
	virtualSMBWriteRequests          *prometheus.Desc // \Hyper-V Virtual SMB(*)\Write Requests/sec
	virtualSMBReadRequestsRDMA       *prometheus.Desc // \Hyper-V Virtual SMB(*)\Read Requests/sec (RDMA)
	virtualSMBReadRequests           *prometheus.Desc // \Hyper-V Virtual SMB(*)\Read Requests/sec
	virtualSMBCurrentPendingRequests *prometheus.Desc // \Hyper-V Virtual SMB(*)\Current Pending Requests
	virtualSMBCurrentOpenFileCount   *prometheus.Desc // \Hyper-V Virtual SMB(*)\Current Open File Count
	virtualSMBTreeConnectCount       *prometheus.Desc // \Hyper-V Virtual SMB(*)\Tree Connect Count
	virtualSMBRequests               *prometheus.Desc // \Hyper-V Virtual SMB(*)\Requests/sec
	virtualSMBSentBytes              *prometheus.Desc // \Hyper-V Virtual SMB(*)\Sent Bytes/sec
	virtualSMBReceivedBytes          *prometheus.Desc // \Hyper-V Virtual SMB(*)\Received Bytes/sec
}

type perfDataCounterValuesVirtualSMB struct {
	Name string

	VirtualSMBDirectMappedSections   float64 `perfdata:"Direct-Mapped Sections"`
	VirtualSMBDirectMappedPages      float64 `perfdata:"Direct-Mapped Pages"`
	VirtualSMBWriteBytesRDMA         float64 `perfdata:"Write Bytes/sec (RDMA)"`
	VirtualSMBWriteBytes             float64 `perfdata:"Write Bytes/sec"`
	VirtualSMBReadBytesRDMA          float64 `perfdata:"Read Bytes/sec (RDMA)"`
	VirtualSMBReadBytes              float64 `perfdata:"Read Bytes/sec"`
	VirtualSMBFlushRequests          float64 `perfdata:"Flush Requests/sec"`
	VirtualSMBWriteRequestsRDMA      float64 `perfdata:"Write Requests/sec (RDMA)"`
	VirtualSMBWriteRequests          float64 `perfdata:"Write Requests/sec"`
	VirtualSMBReadRequestsRDMA       float64 `perfdata:"Read Requests/sec (RDMA)"`
	VirtualSMBReadRequests           float64 `perfdata:"Read Requests/sec"`
	VirtualSMBCurrentPendingRequests float64 `perfdata:"Current Pending Requests"`
	VirtualSMBCurrentOpenFileCount   float64 `perfdata:"Current Open File Count"`
	VirtualSMBTreeConnectCount       float64 `perfdata:"Tree Connect Count"`
	VirtualSMBRequests               float64 `perfdata:"Requests/sec"`
	VirtualSMBSentBytes              float64 `perfdata:"Sent Bytes/sec"`
	VirtualSMBReceivedBytes          float64 `perfdata:"Received Bytes/sec"`
}

func (c *Collector) buildVirtualSMB() error {
	var err error

	c.perfDataCollectorVirtualSMB, err = pdh.NewCollector[perfDataCounterValuesVirtualSMB](pdh.CounterTypeRaw, "Hyper-V Virtual SMB", pdh.InstancesAll)
	if err != nil {
		return fmt.Errorf("failed to create Hyper-V Virtual SMB collector: %w", err)
	}

	c.virtualSMBDirectMappedSections = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_direct_mapped_sections"),
		"Represents the number of direct-mapped sections in the virtual SMB",
		[]string{"instance"},
		nil,
	)
	c.virtualSMBDirectMappedPages = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_direct_mapped_pages"),
		"Represents the number of direct-mapped pages in the virtual SMB",
		[]string{"instance"},
		nil,
	)
	c.virtualSMBWriteBytesRDMA = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_write_bytes_rdma"),
		"Represents the number of bytes written per second using RDMA in the virtual SMB",
		[]string{"instance"},
		nil,
	)
	c.virtualSMBWriteBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_write_bytes"),
		"Represents the number of bytes written per second in the virtual SMB",
		[]string{"instance"},
		nil,
	)
	c.virtualSMBReadBytesRDMA = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_read_bytes_rdma"),
		"Represents the number of bytes read per second using RDMA in the virtual SMB",
		[]string{"instance"},
		nil,
	)
	c.virtualSMBReadBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_read_bytes"),
		"Represents the number of bytes read per second in the virtual SMB",
		[]string{"instance"},
		nil,
	)
	c.virtualSMBFlushRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_flush_requests"),
		"Represents the number of flush requests per second in the virtual SMB",
		[]string{"instance"},
		nil,
	)
	c.virtualSMBWriteRequestsRDMA = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_write_requests_rdma"),
		"Represents the number of write requests per second using RDMA in the virtual SMB",
		[]string{"instance"},
		nil,
	)
	c.virtualSMBWriteRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_write_requests"),
		"Represents the number of write requests per second in the virtual SMB",
		[]string{"instance"},
		nil,
	)
	c.virtualSMBReadRequestsRDMA = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_read_requests_rdma"),
		"Represents the number of read requests per second using RDMA in the virtual SMB",
		[]string{"instance"},
		nil,
	)
	c.virtualSMBReadRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_read_requests"),
		"Represents the number of read requests per second in the virtual SMB",
		[]string{"instance"},
		nil,
	)
	c.virtualSMBCurrentPendingRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_current_pending_requests"),
		"Represents the current number of pending requests in the virtual SMB",
		[]string{"instance"},
		nil,
	)
	c.virtualSMBCurrentOpenFileCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_current_open_file_count"),
		"Represents the current number of open files in the virtual SMB",
		[]string{"instance"},
		nil,
	)
	c.virtualSMBTreeConnectCount = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_tree_connect_count"),
		"Represents the number of tree connects in the virtual SMB",
		[]string{"instance"},
		nil,
	)
	c.virtualSMBRequests = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_requests"),
		"Represents the number of requests per second in the virtual SMB",
		[]string{"instance"},
		nil,
	)
	c.virtualSMBSentBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_sent_bytes"),
		"Represents the number of bytes sent per second in the virtual SMB",
		[]string{"instance"},
		nil,
	)
	c.virtualSMBReceivedBytes = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "virtual_smb_received_bytes"),
		"Represents the number of bytes received per second in the virtual SMB",
		[]string{"instance"},
		nil,
	)

	return nil
}

func (c *Collector) collectVirtualSMB(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorVirtualSMB.Collect(&c.perfDataObjectVirtualSMB)
	if err != nil {
		return fmt.Errorf("failed to collect Hyper-V Virtual SMB metrics: %w", err)
	}

	for _, data := range c.perfDataObjectVirtualSMB {
		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBDirectMappedSections,
			prometheus.GaugeValue,
			data.VirtualSMBDirectMappedSections,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBDirectMappedPages,
			prometheus.GaugeValue,
			data.VirtualSMBDirectMappedPages,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBWriteBytesRDMA,
			prometheus.CounterValue,
			data.VirtualSMBWriteBytesRDMA,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBWriteBytes,
			prometheus.CounterValue,
			data.VirtualSMBWriteBytes,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBReadBytesRDMA,
			prometheus.CounterValue,
			data.VirtualSMBReadBytesRDMA,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBReadBytes,
			prometheus.CounterValue,
			data.VirtualSMBReadBytes,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBFlushRequests,
			prometheus.CounterValue,
			data.VirtualSMBFlushRequests,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBWriteRequestsRDMA,
			prometheus.CounterValue,
			data.VirtualSMBWriteRequestsRDMA,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBWriteRequests,
			prometheus.CounterValue,
			data.VirtualSMBWriteRequests,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBReadRequestsRDMA,
			prometheus.CounterValue,
			data.VirtualSMBReadRequestsRDMA,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBReadRequests,
			prometheus.CounterValue,
			data.VirtualSMBReadRequests,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBCurrentPendingRequests,
			prometheus.GaugeValue,
			data.VirtualSMBCurrentPendingRequests,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBCurrentOpenFileCount,
			prometheus.GaugeValue,
			data.VirtualSMBCurrentOpenFileCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBTreeConnectCount,
			prometheus.GaugeValue,
			data.VirtualSMBTreeConnectCount,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBRequests,
			prometheus.CounterValue,
			data.VirtualSMBRequests,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBSentBytes,
			prometheus.CounterValue,
			data.VirtualSMBSentBytes,
			data.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBReceivedBytes,
			prometheus.CounterValue,
			data.VirtualSMBReceivedBytes,
			data.Name,
		)
	}

	return nil
}
