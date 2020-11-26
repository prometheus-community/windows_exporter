// +build windows

package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	dfsrVolumeSubsystem     = "dfsr_volume"
	dfsrConnectionSubsystem = "dfsr_connection"
	dfsrFoldersSubsystem    = "dfsr_folder"
)

func init() {
	registerCollector(dfsrConnectionSubsystem, NewDFSRConnectionCollector, "DFS Replication Service Connections")
	registerCollector(dfsrFoldersSubsystem, NewDFSRConnectionCollector, "DFS Replication Service Folders")
	registerCollector(dfsrVolumeSubsystem, NewDFSRConnectionCollector, "DFS Replication Service Volumes")
}

type DFSRConnectionCollector struct {
	BandwidthSavingsUsingDFSReplicationTotal *prometheus.Desc
	BytesReceivedTotal                       *prometheus.Desc
	CompressedSizeOfFilesReceivedTotal       *prometheus.Desc
	FilesReceivedTotal                       *prometheus.Desc
	RDCBytesReceivedTotal                    *prometheus.Desc
	RDCCompressedSizeOfFilesReceivedTotal    *prometheus.Desc
	RDCSizeOfFilesReceivedTotal              *prometheus.Desc
	RDCNumberofFilesReceivedTotal            *prometheus.Desc
	SizeOfFilesReceivedTotal                 *prometheus.Desc
}

func NewDFSRConnectionCollector() (Collector, error) {
	return &DFSRConnectionCollector{
		BandwidthSavingsUsingDFSReplicationTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "bandwidth_savings_using_dfs_replication_total"),
			"Total amount of bandwidth savings using DFS Replication for this connection, in bytes",
			[]string{"name"},
			nil,
		),

		BytesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "bytes_received_total"),
			"Total bytes received for connection",
			[]string{"name"},
			nil,
		),

		CompressedSizeOfFilesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "compressed_size_of_files_received_total"),
			"",
			[]string{"name"},
			nil,
		),

		FilesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "files_received_total"),
			"Total number of files receieved for connection",
			[]string{"name"},
			nil,
		),

		RDCBytesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "rdc_bytes_received_total"),
			"",
			[]string{"name"},
			nil,
		),

		RDCCompressedSizeOfFilesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "rdc_compressed_size_of_files_received_total"),
			"",
			[]string{"name"},
			nil,
		),

		RDCNumberofFilesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "rdc_number_of_files_received_total"),
			"",
			[]string{"name"},
			nil,
		),

		RDCSizeOfFilesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "rdc_size_of_files_received_total"),
			"",
			[]string{"name"},
			nil,
		),

		SizeOfFilesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "size_of_files_received_total"),
			"",
			[]string{"name"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *DFSRConnectionCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting dfsr_connection metrics:", desc, err)
		return err
	}
	return nil
}

// Perflib: "DFS Replication Service Connections"
type PerflibDFSRConnection struct {
	Name string

	BandwidthSavingsUsingDFSReplicationTotal float64 `perflib:"Bandwidth Savings Using DFS Replication"`
	BytesReceivedTotal                       float64 `perflib:"Total Bytes Received"`
	CompressedSizeOfFilesReceivedTotal       float64 `perflib:"Compressed Size of Files Received"`
	FilesReceivedTotal                       float64 `perflib:"Total Files Received"`
	RDCBytesReceivedTotal                    float64 `perflib:"RDC Bytes Received"`
	RDCCompressedSizeOfFilesReceivedTotal    float64 `perflib:"RDC Compressed Size of Files Received"`
	RDCNumberofFilesReceivedTotal            float64 `perflib:"RDC Number of Files Received"`
	RDCSizeOfFilesReceivedTotal              float64 `perflib:"RDC Size of Files Received"`
	SizeOfFilesReceivedTotal                 float64 `perflib:"Size of Files Received"`
}

func (c *DFSRConnectionCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []PerflibDFSRConnection
	if err := unmarshalObject(ctx.perfObjects["DFS Replication Connections"], &dst); err != nil {
		return nil, err
	}

	for _, connection := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.BandwidthSavingsUsingDFSReplicationTotal,
			prometheus.CounterValue,
			connection.BandwidthSavingsUsingDFSReplicationTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.BytesReceivedTotal,
			prometheus.CounterValue,
			connection.BytesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.CompressedSizeOfFilesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FilesReceivedTotal,
			prometheus.CounterValue,
			connection.FilesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RDCBytesReceivedTotal,
			prometheus.CounterValue,
			connection.RDCBytesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RDCCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.RDCCompressedSizeOfFilesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RDCSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.RDCSizeOfFilesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RDCNumberofFilesReceivedTotal,
			prometheus.CounterValue,
			connection.RDCNumberofFilesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.SizeOfFilesReceivedTotal,
			connection.Name,
		)

	}
	return nil, nil
}

type DFSRVolumeCollector struct {
	DatabaseLookupsTotal           *prometheus.Desc
	DatabaseCommitsTotal           *prometheus.Desc
	USNJournalUnreadPercentage     *prometheus.Desc
	USNJournalRecordsAcceptedTotal *prometheus.Desc
	USNJournalRecordsReadTotal     *prometheus.Desc
}

func NewDFSRVolumeCollector() (Collector, error) {
	return &DFSRVolumeCollector{
		DatabaseCommitsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "database_commits_total"),
			"Total number of DFSR Volume database commits",
			[]string{"name"},
			nil,
		),

		DatabaseLookupsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "database_lookups_total"),
			"Total number of DFSR Volume database lookups",
			[]string{"name"},
			nil,
		),

		USNJournalUnreadPercentage: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "usn_journal_unread_percentage"),
			"Percentage of DFSR Volume USN journal records that are unread",
			[]string{"name"},
			nil,
		),

		USNJournalRecordsAcceptedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "usn_journal_records_accepted_total"),
			"Total number of USN journal records accepted",
			[]string{"name"},
			nil,
		),

		USNJournalRecordsReadTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "usn_journal_records_read_total"),
			"Total number of DFSR Volume USN journal records read",
			[]string{"name"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *DFSRVolumeCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting dfsr_volume metrics:", desc, err)
		return err
	}
	return nil
}

// Perflib: "DFS Replication Service Volumes"
type PerflibDFSRVolume struct {
	Name string

	DatabaseCommitsTotal           float64 `perflib:"Database Commits"`
	DatabaseLookupsTotal           float64 `perflib:"Database Lookups"`
	USNJournalRecordsReadTotal     float64 `perflib:"USN Journal Records Read"`
	USNJournalRecordsAcceptedTotal float64 `perflib:"USN Journal Records Accepted"`
	USNJournalUnreadPercentage     float64 `perflib:"USN Journal Records Unread Percentage"`
}

func (c *DFSRVolumeCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []PerflibDFSRVolume
	if err := unmarshalObject(ctx.perfObjects["DFS Replication Service Volumes"], &dst); err != nil {
		return nil, err
	}

	for _, volume := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.DatabaseLookupsTotal,
			prometheus.CounterValue,
			volume.DatabaseLookupsTotal,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DatabaseCommitsTotal,
			prometheus.CounterValue,
			volume.DatabaseCommitsTotal,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.USNJournalRecordsAcceptedTotal,
			prometheus.CounterValue,
			volume.USNJournalRecordsAcceptedTotal,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.USNJournalRecordsReadTotal,
			prometheus.CounterValue,
			volume.USNJournalRecordsReadTotal,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.USNJournalUnreadPercentage,
			prometheus.GaugeValue,
			volume.USNJournalUnreadPercentage,
			volume.Name,
		)

	}
	return nil, nil
}

type DFSRReplicatedFoldersCollector struct {
	BandwidthSavingsUsingDFSReplicationTotal *prometheus.Desc
	CompressedSizeOfFilesReceivedTotal       *prometheus.Desc
	ConflictBytesCleanedupTotal              *prometheus.Desc
	ConflictBytesGeneratedTotal              *prometheus.Desc
	ConflictFilesCleanedUpTotal              *prometheus.Desc
	ConflictFilesGeneratedTotal              *prometheus.Desc
	ConflictFolderCleanupsCompletedTotal     *prometheus.Desc
	ConflictSpaceInUse                       *prometheus.Desc
	DeletedSpaceInUse                        *prometheus.Desc
	DeletedBytesCleanedUpTotal               *prometheus.Desc
	DeletedBytesGeneratedTotal               *prometheus.Desc
	DeletedFilesCleanedUpTotal               *prometheus.Desc
	DeletedFilesGeneratedTotal               *prometheus.Desc
	FileInstallsRetriedTotal                 *prometheus.Desc
	FileInstallsSucceededTotal               *prometheus.Desc
	FilesReceivedTotal                       *prometheus.Desc
	RDCBytesReceivedTotal                    *prometheus.Desc
	RDCCompressedSizeOfFilesReceivedTotal    *prometheus.Desc
	RDCNumberofFilesReceivedTotal            *prometheus.Desc
	RDCSizeOfFilesReceivedTotal              *prometheus.Desc
	SizeOfFilesReceivedTotal                 *prometheus.Desc
	StagingSpaceInUse                        *prometheus.Desc
	StagingBytesCleanedUpTotal               *prometheus.Desc
	StagingBytesGeneratedTotal               *prometheus.Desc
	StagingFilesCleanedUpTotal               *prometheus.Desc
	StagingFilesGeneratedTotal               *prometheus.Desc
	UpdatesDroppedTotal                      *prometheus.Desc
}

func NewDFSRReplicatedFoldersCollector() (Collector, error) {
	return &DFSRReplicatedFoldersCollector{
		BandwidthSavingsUsingDFSReplicationTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "bandwidth_savings_using_dfs_replication_total"),
			"",
			[]string{"name"},
			nil,
		),

		CompressedSizeOfFilesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "compressed_size_of_files_received_total"),
			"",
			[]string{"name"},
			nil,
		),

		ConflictBytesCleanedupTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "conflict_bytes_cleaned_up_total"),
			"",
			[]string{"name"},
			nil,
		),

		ConflictBytesGeneratedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "conflict_bytes_generated_total"),
			"",
			[]string{"name"},
			nil,
		),

		ConflictFilesCleanedUpTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "conflict_files_cleaned_up_total"),
			"",
			[]string{"name"},
			nil,
		),

		ConflictFilesGeneratedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "conflict_files_generated_total"),
			"",
			[]string{"name"},
			nil,
		),

		ConflictFolderCleanupsCompletedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "conflict_folder_cleanups_total"),
			"",
			[]string{"name"},
			nil,
		),

		ConflictSpaceInUse: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "conflict_space_in_use"),
			"",
			[]string{"name"},
			nil,
		),

		DeletedSpaceInUse: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "deleted_space_in_use"),
			"",
			[]string{"name"},
			nil,
		),

		DeletedBytesCleanedUpTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "deleted_bytes_cleaned_up_total"),
			"",
			[]string{"name"},
			nil,
		),

		DeletedBytesGeneratedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "deleted_bytes_generated_total"),
			"",
			[]string{"name"},
			nil,
		),

		DeletedFilesCleanedUpTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "deleted_files_cleaned_up_total"),
			"",
			[]string{"name"},
			nil,
		),

		DeletedFilesGeneratedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "deleted_files_generated_total"),
			"",
			[]string{"name"},
			nil,
		),

		FileInstallsRetriedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "file_installs_retried_total"),
			"",
			[]string{"name"},
			nil,
		),

		FileInstallsSucceededTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "file_installs_succeeded_total"),
			"",
			[]string{"name"},
			nil,
		),

		FilesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "files_received_total"),
			"",
			[]string{"name"},
			nil,
		),

		RDCBytesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "rdc_bytes_received_total"),
			"",
			[]string{"name"},
			nil,
		),

		RDCCompressedSizeOfFilesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "rdc_compressed_size_of_files_received_total"),
			"",
			[]string{"name"},
			nil,
		),

		RDCNumberofFilesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "rdc_number_of_files_received_total"),
			"",
			[]string{"name"},
			nil,
		),

		RDCSizeOfFilesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "rdc_size_of_files_received_total"),
			"",
			[]string{"name"},
			nil,
		),

		SizeOfFilesReceivedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "size_of_files_received_total"),
			"",
			[]string{"name"},
			nil,
		),

		StagingSpaceInUse: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "staging_space_in_use"),
			"",
			[]string{"name"},
			nil,
		),

		StagingBytesCleanedUpTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "staging_bytes_cleaned_up_total"),
			"",
			[]string{"name"},
			nil,
		),

		StagingBytesGeneratedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "staging_bytes_generated_total"),
			"",
			[]string{"name"},
			nil,
		),

		StagingFilesCleanedUpTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "staging_files_cleaned_up_total"),
			"",
			[]string{"name"},
			nil,
		),

		StagingFilesGeneratedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "staging_files_generated_total"),
			"",
			[]string{"name"},
			nil,
		),

		UpdatesDroppedTotal: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, dfsrVolumeSubsystem, "updates_dropped_total"),
			"",
			[]string{"name"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *DFSRReplicatedFoldersCollector) Collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ctx, ch); err != nil {
		log.Error("failed collecting dfsr_folder metrics:", desc, err)
		return err
	}
	return nil
}

// Perflib: "DFS Replicated Folder"
type PerflibDFSRFolder struct {
	Name string

	BandwidthSavingsUsingDFSReplicationTotal float64 `perflib:"Bandwidth Savings Using DFS Replication"`
	CompressedSizeOfFilesReceivedTotal       float64 `perflib:"Compressed Size of Files Received"`
	ConflictBytesCleanedupTotal              float64 `perflib:"Conflict Bytes Cleaned Up"`
	ConflictBytesGeneratedTotal              float64 `perflib:"Conflict Bytes Generated"`
	ConflictFilesCleanedUpTotal              float64 `perflib:"Conflict Files Cleaned Up"`
	ConflictFilesGeneratedTotal              float64 `perflib:"Conflict Files Generated"`
	ConflictFolderCleanupsCompletedTotal     float64 `perflib:"Conflict Folder Cleanups Completed"`
	ConflictSpaceInUse                       float64 `perflib:"Conflict Space In Use"`
	DeletedSpaceInUse                        float64 `perflib:"Deleted Space In Use"`
	DeletedBytesCleanedUpTotal               float64 `perflib:"Deleted Bytes Cleaned Up"`
	DeletedBytesGeneratedTotal               float64 `perflib:"Deleted Bytes Generated"`
	DeletedFilesCleanedUpTotal               float64 `perflib:"Deleted Files Cleaned Up"`
	DeletedFilesGeneratedTotal               float64 `perflib:"Deleted Files Generated"`
	FileInstallsRetriedTotal                 float64 `perflib:"File Installs Retried"`
	FileInstallsSucceededTotal               float64 `perflib:"File Installs Succeeded"`
	FilesReceivedTotal                       float64 `perflib:"Total Files Received"`
	RDCBytesReceivedTotal                    float64 `perflib:"RDC Bytes Received"`
	RDCCompressedSizeOfFilesReceivedTotal    float64 `perflib:"RDC Compressed Size of Files Received"`
	RDCNumberofFilesReceivedTotal            float64 `perflib:"RDC Number of Files Received"`
	RDCSizeOfFilesReceivedTotal              float64 `perflib:"RDC Size of Files Received"`
	SizeOfFilesReceivedTotal                 float64 `perflib:"Size of Files Received"`
	StagingSpaceInUse                        float64 `perflib:"Staging Space In Use"`
	StagingBytesCleanedUpTotal               float64 `perflib:"Staging Bytes Cleaned Up"`
	StagingBytesGeneratedTotal               float64 `perflib:"Staging Bytes Generated"`
	StagingFilesCleanedUpTotal               float64 `perflib:"Staging Files Cleaned Up"`
	StagingFilesGeneratedTotal               float64 `perflib:"Staging Files Generated"`
	UpdatesDroppedTotal                      float64 `perflib:"Updates Dropped"`
}

func (c *DFSRReplicatedFoldersCollector) collect(ctx *ScrapeContext, ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []PerflibDFSRFolder
	if err := unmarshalObject(ctx.perfObjects["DFS Replicated Folders"], &dst); err != nil {
		return nil, err
	}

	for _, folder := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.BandwidthSavingsUsingDFSReplicationTotal,
			prometheus.CounterValue,
			folder.BandwidthSavingsUsingDFSReplicationTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.CompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.CompressedSizeOfFilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConflictBytesCleanedupTotal,
			prometheus.CounterValue,
			folder.ConflictBytesCleanedupTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConflictBytesGeneratedTotal,
			prometheus.CounterValue,
			folder.ConflictBytesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConflictFilesCleanedUpTotal,
			prometheus.CounterValue,
			folder.ConflictFilesCleanedUpTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConflictFilesGeneratedTotal,
			prometheus.CounterValue,
			folder.ConflictFilesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConflictFolderCleanupsCompletedTotal,
			prometheus.CounterValue,
			folder.ConflictFolderCleanupsCompletedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConflictSpaceInUse,
			prometheus.GaugeValue,
			folder.ConflictSpaceInUse,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DeletedSpaceInUse,
			prometheus.GaugeValue,
			folder.DeletedSpaceInUse,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DeletedBytesCleanedUpTotal,
			prometheus.CounterValue,
			folder.DeletedBytesCleanedUpTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DeletedBytesGeneratedTotal,
			prometheus.CounterValue,
			folder.DeletedBytesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DeletedFilesCleanedUpTotal,
			prometheus.CounterValue,
			folder.DeletedFilesCleanedUpTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.DeletedFilesGeneratedTotal,
			prometheus.CounterValue,
			folder.DeletedFilesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FileInstallsRetriedTotal,
			prometheus.CounterValue,
			folder.FileInstallsRetriedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FileInstallsSucceededTotal,
			prometheus.CounterValue,
			folder.FileInstallsSucceededTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FilesReceivedTotal,
			prometheus.CounterValue,
			folder.FilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RDCBytesReceivedTotal,
			prometheus.CounterValue,
			folder.RDCBytesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RDCCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.RDCCompressedSizeOfFilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RDCNumberofFilesReceivedTotal,
			prometheus.CounterValue,
			folder.RDCNumberofFilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.RDCSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.RDCSizeOfFilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.SizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.SizeOfFilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.StagingSpaceInUse,
			prometheus.GaugeValue,
			folder.StagingSpaceInUse,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.StagingBytesCleanedUpTotal,
			prometheus.CounterValue,
			folder.StagingBytesCleanedUpTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.StagingBytesGeneratedTotal,
			prometheus.CounterValue,
			folder.StagingBytesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.StagingFilesCleanedUpTotal,
			prometheus.CounterValue,
			folder.StagingFilesCleanedUpTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.StagingFilesGeneratedTotal,
			prometheus.CounterValue,
			folder.StagingFilesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.UpdatesDroppedTotal,
			prometheus.CounterValue,
			folder.UpdatesDroppedTotal,
			folder.Name,
		)
	}
	return nil, nil
}
