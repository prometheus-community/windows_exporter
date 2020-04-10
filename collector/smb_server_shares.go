// +build windows

package collector

import (
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	registerCollector("smb_server_shares", NewSmbServerSharesCollector, "SMB Server Shares")
}

var (
	shareWhitelist = kingpin.Flag(
		"collector.smb_server_shares.share-whitelist",
		"Regexp of shares to whitelist. Share name must both match whitelist and not match blacklist to be included.",
	).Default(".+").String()
	shareBlacklist = kingpin.Flag(
		"collector.smb_server_shares.share-blacklist",
		"Regexp of shares to blacklist. Share name must both match whitelist and not match blacklist to be included.",
	).Default("").String()
)

// A SmbServerSharesCollector is a Prometheus collector for perflib SmbServerShares metrics
type SmbServerSharesCollector struct {
	CurrentOpenFileCount        *prometheus.Desc
	TotalFileOpenCount          *prometheus.Desc
	FilesOpenedPerSec           *prometheus.Desc
	CurrentDurableOpenFileCount *prometheus.Desc

	shareWhitelistPattern *regexp.Regexp
	shareBlacklistPattern *regexp.Regexp
}

// NewSmbServerSharesCollector ...
func NewSmbServerSharesCollector() (Collector, error) {
	const subsystem = "smb_server_shares"

	return &SmbServerSharesCollector{
		CurrentOpenFileCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_open_file_count"),
			"The number of file handles that are currently open in this share (SmbServerShares.CurrentOpenFileCount)",
			[]string{"share"},
			nil,
		),
		TotalFileOpenCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "file_open_count"),
			"The number of files that have been opened by the SMB File Server on behalf of its clients on this share since the server started. (SmbServerShares.TotalFileOpenCount)",
			[]string{"share"},
			nil,
		),
		FilesOpenedPerSec: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "files_opened_total"),
			"The rate, in seconds, at which files are being opened for the SMB File Server’s clients on this share. (SmbServerShares.FilesOpenedPerSec)",
			[]string{"share"},
			nil,
		),
		CurrentDurableOpenFileCount: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "current_durable_open_file_count"),
			"The number of durable file handles that are currently open on this share (SmbServerShares.CurrentDurableOpenFileCount)",
			[]string{"share"},
			nil,
		),

		shareWhitelistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *shareWhitelist)),
		shareBlacklistPattern: regexp.MustCompile(fmt.Sprintf("^(?:%s)$", *shareBlacklist)),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *SmbServerSharesCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting smb_server_shares metrics:", desc, err)
		return err
	}
	return nil
}

// :9432/dump?query=4352#4352
type SmbServerShares struct {
	Name string
	/*
		ReceivedBytesPerSec                    float64 `perflib:"Received Bytes/sec"`
		RequestsPerSec                         float64 `perflib:"Requests/sec"`
		TreeConnectCount                       float64 `perflib:"Tree Connect Count"`*/
	CurrentOpenFileCount float64 `perflib:"Current Open File Count"`
	/*
		SentBytesPerSec                        float64 `perflib:"Sent Bytes/sec"`
		TransferredBytesPerSec                 float64 `perflib:"Transferred Bytes/sec"`
		CurrentPendingRequests                 float64 `perflib:"Current Pending Requests"`
		AvgSecPerRequest                       float64 `perflib:"Avg. sec/Request"`
		WriteRequestsPerSec                    float64 `perflib:"Write Requests/sec"`
		AvgSecPerWrite                         float64 `perflib:"Avg. sec/Write"`
		WriteBytesPerSec                       float64 `perflib:"Write Bytes/sec"`
		ReadRequestsPerSec                     float64 `perflib:"Read Requests/sec"`
		AvgSecPerRead                          float64 `perflib:"Avg. sec/Read"`
		ReadBytesPerSec                        float64 `perflib:"Read Bytes/sec"` */
	TotalFileOpenCount          float64 `perflib:"Total File Open Count"`
	FilesOpenedPerSec           float64 `perflib:"Files Opened/sec"`
	CurrentDurableOpenFileCount float64 `perflib:"Current Durable Open File Count"`
	/*
		TotalDurableHandleReopenCount          float64 `perflib:"Total Durable Handle Reopen Count"`
		TotalFailedHandleReopenCount           float64 `perflib:"Total Failed Durable Handle Reopen Count"`
		PercentResilientHandles                float64 `perflib:"% Resilient Handles"`
		TotalResilientHandleReopenCount        float64 `perflib:"Total Resilient Handle Reopen Count"`
		TotalFailedResilientHandleReopenCount  float64 `perflib:"Total Failed Resilient Handle Reopen Count"`
		PercentPersistentHandles               float64 `perflib:"% Persistent Handles"`
		TotalPersistentHandleReopenCount       float64 `perflib:"Total Persistent Handle Reopen Count"`
		TotalFailedPersistentHandleReopenCount float64 `perflib:"Total Failed Persistent Handle Reopen Count"`
		MetadataRequestsPerSec                 float64 `perflib:"Metadata Requests/sec"`
		AvgSecPerDataRequest                   float64 `perflib:"Avg. sec/Data Request"`
		AvgDataBytesPerRequest                 float64 `perflib:"Avg. Data Bytes/Request"`
		AvgBytesPerRead                        float64 `perflib:"Avg. Bytes/Read"`
		AvgBytesPerWrite                       float64 `perflib:"Avg. Bytes/Write"`
		AvgReadQueueLength                     float64 `perflib:"Avg. Read Queue Length"`
		AvgWriteQueueLength                    float64 `perflib:"Avg. Write Queue Length"`
		AvgDataQueueLength                     float64 `perflib:"Avg. Data Queue Length"`
		DataBytesPerSec                        float64 `perflib:"Data Bytes/sec"`
		DataRequestsPerSec                     float64 `perflib:"Data Requests/sec"`
		CurrentDataQueueLength                 float64 `perflib:"Current Data Queue Length"` */
}

func (c *SmbServerSharesCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []SmbServerShares
	if err := unmarshalObject(ctx.perfObjects["SMB Server Shares"], &dst); err != nil {
		return nil, err
	}

	for _, share := range dst {
		if share.Name == "_Total" ||
			c.shareBlacklistPattern.MatchString(share.Name) ||
			!c.shareWhitelistPattern.MatchString(share.Name) {
			continue
		}
		/*
			ch <- prometheus.MustNewConstMetric(
				c.ReceivedBytesPerSec,
				prometheus.CounterValue
				share.ReceivedBytesPerSec,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.RequestsPerSec,
				prometheus.CounterValue,
				share.RequestsPerSec,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.TreeConnectCount,
				prometheus.GaugeValue,
				share.TreeConnectCount,
				share.Name,
			) */

		ch <- prometheus.MustNewConstMetric(
			c.CurrentOpenFileCount,
			prometheus.GaugeValue,
			share.CurrentOpenFileCount,
			share.Name,
		)
		/*
			ch <- prometheus.MustNewConstMetric(
				c.SentBytesPerSec,
				prometheus.CounterValue,
				share.SentBytesPerSec,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.TransferredBytesPerSec,
				prometheus.CounterValue,
				share.TransferredBytesPerSec,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.CurrentPendingRequests,
				prometheus.GaugeValue,
				share.CurrentPendingRequests,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.AvgSecPerRequest,
				prometheus.CounterValue,
				share.AvgSecPerRequest,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.WriteRequestsPerSec,
				prometheus.CounterValue,
				share.WriteRequestsPerSec,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.AvgSecPerWrite,
				prometheus.CounterValue,
				share.AvgSecPerWrite,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.WriteLatency,
				prometheus.CounterValue,
				share.WriteBytesPerSec*ticksToSecondsScaleFactor,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.ReadRequestsPerSec,
				prometheus.CounterValue,
				share.ReadRequestsPerSec,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.AvgSecPerRead,
				prometheus.CounterValue,
				share.AvgSecPerRead*ticksToSecondsScaleFactor,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.ReadBytesPerSec,
				prometheus.CounterValue,
				share.ReadBytesPerSec,
				share.Name,
			) */

		ch <- prometheus.MustNewConstMetric(
			c.TotalFileOpenCount,
			prometheus.GaugeValue,
			share.TotalFileOpenCount,
			share.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FilesOpenedPerSec,
			prometheus.CounterValue,
			share.FilesOpenedPerSec,
			share.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CurrentDurableOpenFileCount,
			prometheus.GaugeValue,
			share.CurrentDurableOpenFileCount,
			share.Name,
		)
		/*
			ch <- prometheus.MustNewConstMetric(
				c.TotalDurableHandleReopenCount,
				prometheus.CounterValue,
				share.TotalDurableHandleReopenCount,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.TotalFailedHandleReopenCount,
				prometheus.CounterValue,
				share.TotalFailedHandleReopenCount,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.PercentResilientHandles,
				prometheus.CounterValue,
				share.PercentResilientHandles,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.TotalResilientHandleReopenCount,
				prometheus.GaugeValue,
				share.TotalResilientHandleReopenCount,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.TotalFailedResilientHandleReopenCount,
				prometheus.GaugeValue,
				share.TotalFailedResilientHandleReopenCount,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.PercentPersistentHandles,
				prometheus.CounterValue,
				share.PercentPersistentHandles,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.TotalPersistentHandleReopenCount,
				prometheus.GaugeValue,
				share.TotalPersistentHandleReopenCount,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.TotalFailedPersistentHandleReopenCount,
				prometheus.GaugeValue,
				share.TotalFailedPersistentHandleReopenCount,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.MetadataRequestsPerSec,
				prometheus.CounterValue,
				share.MetadataRequestsPerSec,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.AvgSecPerDataRequest,
				prometheus.CounterValue,
				share.AvgSecPerDataRequest,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.AvgDataBytesPerRequest,
				prometheus.CounterValue,
				share.AvgDataBytesPerRequest,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.AvgBytesPerRead,
				prometheus.CounterValue,
				share.AvgBytesPerRead,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.AvgBytesPerWrite,
				prometheus.CounterValue,
				share.AvgBytesPerWrite,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.AvgReadQueueLength,
				prometheus.CounterValue,
				share.AvgReadQueueLength,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.AvgWriteQueueLength,
				prometheus.CounterValue,
				share.AvgWriteQueueLength,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.AvgDataQueueLength,
				prometheus.CounterValue,
				share.AvgDataQueueLength,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.DataBytesPerSec,
				prometheus.CounterValue,
				share.DataBytesPerSec,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.DataRequestsPerSec,
				prometheus.CounterValue,
				share.DataRequestsPerSec,
				share.Name,
			)

			ch <- prometheus.MustNewConstMetric(
				c.CurrentDataQueueLength,
				prometheus.GaugeValue,
				share.CurrentDataQueueLength,
				share.Name,
			) */
	}

	return nil, nil
}
