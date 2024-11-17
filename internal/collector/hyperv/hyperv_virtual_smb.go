package hyperv

import (
	"errors"
	"fmt"

	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// collectorVirtualSMB Hyper-V Virtual SMB metrics
type collectorVirtualSMB struct {
	perfDataCollectorVirtualSMB *perfdata.Collector

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

const (
	virtualSMBDirectMappedSections   = "Direct-Mapped Sections"
	virtualSMBDirectMappedPages      = "Direct-Mapped Pages"
	virtualSMBWriteBytesRDMA         = "Write Bytes/sec (RDMA)"
	virtualSMBWriteBytes             = "Write Bytes/sec"
	virtualSMBReadBytesRDMA          = "Read Bytes/sec (RDMA)"
	virtualSMBReadBytes              = "Read Bytes/sec"
	virtualSMBFlushRequests          = "Flush Requests/sec"
	virtualSMBWriteRequestsRDMA      = "Write Requests/sec (RDMA)"
	virtualSMBWriteRequests          = "Write Requests/sec"
	virtualSMBReadRequestsRDMA       = "Read Requests/sec (RDMA)"
	virtualSMBReadRequests           = "Read Requests/sec"
	virtualSMBCurrentPendingRequests = "Current Pending Requests"
	virtualSMBCurrentOpenFileCount   = "Current Open File Count"
	virtualSMBTreeConnectCount       = "Tree Connect Count"
	virtualSMBRequests               = "Requests/sec"
	virtualSMBSentBytes              = "Sent Bytes/sec"
	virtualSMBReceivedBytes          = "Received Bytes/sec"
)

func (c *Collector) buildVirtualSMB() error {
	var err error

	c.perfDataCollectorVirtualSMB, err = perfdata.NewCollector("Hyper-V Virtual SMB", perfdata.InstanceAll, []string{
		virtualSMBDirectMappedSections,
		virtualSMBDirectMappedPages,
		virtualSMBWriteBytesRDMA,
		virtualSMBWriteBytes,
		virtualSMBReadBytesRDMA,
		virtualSMBReadBytes,
		virtualSMBFlushRequests,
		virtualSMBWriteRequestsRDMA,
		virtualSMBWriteRequests,
		virtualSMBReadRequestsRDMA,
		virtualSMBReadRequests,
		virtualSMBCurrentPendingRequests,
		virtualSMBCurrentOpenFileCount,
		virtualSMBTreeConnectCount,
		virtualSMBRequests,
		virtualSMBSentBytes,
		virtualSMBReceivedBytes,
	})
	if err != nil && !errors.Is(err, perfdata.ErrNoData) {
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
	data, err := c.perfDataCollectorVirtualSMB.Collect()
	if err != nil && !errors.Is(err, perfdata.ErrNoData) {
		return fmt.Errorf("failed to collect Hyper-V Virtual SMB metrics: %w", err)
	}

	for name, smbData := range data {
		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBDirectMappedSections,
			prometheus.GaugeValue,
			smbData[virtualSMBDirectMappedSections].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBDirectMappedPages,
			prometheus.GaugeValue,
			smbData[virtualSMBDirectMappedPages].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBWriteBytesRDMA,
			prometheus.CounterValue,
			smbData[virtualSMBWriteBytesRDMA].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBWriteBytes,
			prometheus.CounterValue,
			smbData[virtualSMBWriteBytes].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBReadBytesRDMA,
			prometheus.CounterValue,
			smbData[virtualSMBReadBytesRDMA].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBReadBytes,
			prometheus.CounterValue,
			smbData[virtualSMBReadBytes].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBFlushRequests,
			prometheus.CounterValue,
			smbData[virtualSMBFlushRequests].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBWriteRequestsRDMA,
			prometheus.CounterValue,
			smbData[virtualSMBWriteRequestsRDMA].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBWriteRequests,
			prometheus.CounterValue,
			smbData[virtualSMBWriteRequests].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBReadRequestsRDMA,
			prometheus.CounterValue,
			smbData[virtualSMBReadRequestsRDMA].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBReadRequests,
			prometheus.CounterValue,
			smbData[virtualSMBReadRequests].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBCurrentPendingRequests,
			prometheus.GaugeValue,
			smbData[virtualSMBCurrentPendingRequests].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBCurrentOpenFileCount,
			prometheus.GaugeValue,
			smbData[virtualSMBCurrentOpenFileCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBTreeConnectCount,
			prometheus.GaugeValue,
			smbData[virtualSMBTreeConnectCount].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBRequests,
			prometheus.CounterValue,
			smbData[virtualSMBRequests].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBSentBytes,
			prometheus.CounterValue,
			smbData[virtualSMBSentBytes].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.virtualSMBReceivedBytes,
			prometheus.CounterValue,
			smbData[virtualSMBReceivedBytes].FirstValue,
			name,
		)
	}

	return nil
}
