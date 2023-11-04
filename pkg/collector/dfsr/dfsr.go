//go:build windows

package dfsr

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Name                      = "dfsr"
	FlagDfsrEnabledCollectors = "collectors.dfsr.sources-enabled"
)

type Config struct {
	DfsrEnabledCollectors string `yaml:"enabled_collectors"`
}

var ConfigDefaults = Config{
	DfsrEnabledCollectors: "",
}

// collector contains the metric and state data of the DFSR collectors.
type collector struct {
	logger log.Logger

	dfsrEnabledCollectors *string

	// Connection source
	ConnectionBandwidthSavingsUsingDFSReplicationTotal *prometheus.Desc
	ConnectionBytesReceivedTotal                       *prometheus.Desc
	ConnectionCompressedSizeOfFilesReceivedTotal       *prometheus.Desc
	ConnectionFilesReceivedTotal                       *prometheus.Desc
	ConnectionRDCBytesReceivedTotal                    *prometheus.Desc
	ConnectionRDCCompressedSizeOfFilesReceivedTotal    *prometheus.Desc
	ConnectionRDCSizeOfFilesReceivedTotal              *prometheus.Desc
	ConnectionRDCNumberofFilesReceivedTotal            *prometheus.Desc
	ConnectionSizeOfFilesReceivedTotal                 *prometheus.Desc

	// Folder source
	FolderBandwidthSavingsUsingDFSReplicationTotal *prometheus.Desc
	FolderCompressedSizeOfFilesReceivedTotal       *prometheus.Desc
	FolderConflictBytesCleanedupTotal              *prometheus.Desc
	FolderConflictBytesGeneratedTotal              *prometheus.Desc
	FolderConflictFilesCleanedUpTotal              *prometheus.Desc
	FolderConflictFilesGeneratedTotal              *prometheus.Desc
	FolderConflictFolderCleanupsCompletedTotal     *prometheus.Desc
	FolderConflictSpaceInUse                       *prometheus.Desc
	FolderDeletedSpaceInUse                        *prometheus.Desc
	FolderDeletedBytesCleanedUpTotal               *prometheus.Desc
	FolderDeletedBytesGeneratedTotal               *prometheus.Desc
	FolderDeletedFilesCleanedUpTotal               *prometheus.Desc
	FolderDeletedFilesGeneratedTotal               *prometheus.Desc
	FolderFileInstallsRetriedTotal                 *prometheus.Desc
	FolderFileInstallsSucceededTotal               *prometheus.Desc
	FolderFilesReceivedTotal                       *prometheus.Desc
	FolderRDCBytesReceivedTotal                    *prometheus.Desc
	FolderRDCCompressedSizeOfFilesReceivedTotal    *prometheus.Desc
	FolderRDCNumberofFilesReceivedTotal            *prometheus.Desc
	FolderRDCSizeOfFilesReceivedTotal              *prometheus.Desc
	FolderSizeOfFilesReceivedTotal                 *prometheus.Desc
	FolderStagingSpaceInUse                        *prometheus.Desc
	FolderStagingBytesCleanedUpTotal               *prometheus.Desc
	FolderStagingBytesGeneratedTotal               *prometheus.Desc
	FolderStagingFilesCleanedUpTotal               *prometheus.Desc
	FolderStagingFilesGeneratedTotal               *prometheus.Desc
	FolderUpdatesDroppedTotal                      *prometheus.Desc

	// Volume source
	VolumeDatabaseLookupsTotal           *prometheus.Desc
	VolumeDatabaseCommitsTotal           *prometheus.Desc
	VolumeUSNJournalUnreadPercentage     *prometheus.Desc
	VolumeUSNJournalRecordsAcceptedTotal *prometheus.Desc
	VolumeUSNJournalRecordsReadTotal     *prometheus.Desc

	// Map of child collector functions used during collection
	dfsrChildCollectors []dfsrCollectorFunc
}

type dfsrCollectorFunc func(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error

// Map Perflib sources to DFSR collector names
// e.g, volume -> DFS Replication Service Volumes
func dfsrGetPerfObjectName(collector string) string {
	prefix := "DFS "
	suffix := ""
	switch collector {
	case "connection":
		suffix = "Replication Connections"
	case "folder":
		suffix = "Replicated Folders"
	case "volume":
		suffix = "Replication Service Volumes"
	}
	return prefix + suffix
}

func New(logger log.Logger, config *Config) types.Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &collector{
		dfsrEnabledCollectors: &config.DfsrEnabledCollectors,
	}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(app *kingpin.Application) types.Collector {
	return &collector{
		dfsrEnabledCollectors: app.
			Flag(FlagDfsrEnabledCollectors, "Comma-seperated list of DFSR Perflib sources to use.").Default("connection,folder,volume").
			Default(ConfigDefaults.DfsrEnabledCollectors).String(),
	}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	// Perflib sources are dynamic, depending on the enabled child collectors
	var perflibDependencies []string
	for _, source := range utils.ExpandEnabledChildCollectors(*c.dfsrEnabledCollectors) {
		perflibDependencies = append(perflibDependencies, dfsrGetPerfObjectName(source))
	}

	return perflibDependencies, nil
}

func (c *collector) Build() error {
	_ = level.Info(c.logger).Log("msg", "dfsr collector is in an experimental state! Metrics for this collector have not been tested.")

	// Connection
	c.ConnectionBandwidthSavingsUsingDFSReplicationTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_bandwidth_savings_using_dfs_replication_bytes_total"),
		"Total bytes of bandwidth saved using DFS Replication for this connection",
		[]string{"name"},
		nil,
	)

	c.ConnectionBytesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_bytes_received_total"),
		"Total bytes received for connection",
		[]string{"name"},
		nil,
	)

	c.ConnectionCompressedSizeOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_compressed_size_of_files_received_bytes_total"),
		"Total compressed size of files received on the connection, in bytes",
		[]string{"name"},
		nil,
	)

	c.ConnectionFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_received_files_total"),
		"Total number of files received for connection",
		[]string{"name"},
		nil,
	)

	c.ConnectionRDCBytesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_rdc_received_bytes_total"),
		"Total bytes received on the connection while replicating files using Remote Differential Compression",
		[]string{"name"},
		nil,
	)

	c.ConnectionRDCCompressedSizeOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_rdc_compressed_size_of_received_files_bytes_total"),
		"Total uncompressed size of files received with Remote Differential Compression for connection",
		[]string{"name"},
		nil,
	)

	c.ConnectionRDCNumberofFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_rdc_received_files_total"),
		"Total number of files received using remote differential compression",
		[]string{"name"},
		nil,
	)

	c.ConnectionRDCSizeOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_rdc_size_of_received_files_bytes_total"),
		"Total size of received Remote Differential Compression files, in bytes.",
		[]string{"name"},
		nil,
	)

	c.ConnectionSizeOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_files_received_bytes_total"),
		"Total size of files received, in bytes",
		[]string{"name"},
		nil,
	)

	c. // Folder
		FolderBandwidthSavingsUsingDFSReplicationTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_bandwidth_savings_using_dfs_replication_bytes_total"),
		"Total bytes of bandwidth saved using DFS Replication for this folder",
		[]string{"name"},
		nil,
	)

	c.FolderCompressedSizeOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_compressed_size_of_received_files_bytes_total"),
		"Total compressed size of files received on the folder, in bytes",
		[]string{"name"},
		nil,
	)

	c.FolderConflictBytesCleanedupTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_conflict_cleaned_up_bytes_total"),
		"Total size of conflict loser files and folders deleted from the Conflict and Deleted folder, in bytes",
		[]string{"name"},
		nil,
	)

	c.FolderConflictBytesGeneratedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_conflict_generated_bytes_total"),
		"Total size of conflict loser files and folders moved to the Conflict and Deleted folder, in bytes",
		[]string{"name"},
		nil,
	)

	c.FolderConflictFilesCleanedUpTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_conflict_cleaned_up_files_total"),
		"Number of conflict loser files deleted from the Conflict and Deleted folder",
		[]string{"name"},
		nil,
	)

	c.FolderConflictFilesGeneratedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_conflict_generated_files_total"),
		"Number of files and folders moved to the Conflict and Deleted folder",
		[]string{"name"},
		nil,
	)

	c.FolderConflictFolderCleanupsCompletedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_conflict_folder_cleanups_total"),
		"Number of deletions of conflict loser files and folders in the Conflict and Deleted",
		[]string{"name"},
		nil,
	)

	c.FolderConflictSpaceInUse = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_conflict_space_in_use_bytes"),
		"Total size of the conflict loser files and folders currently in the Conflict and Deleted folder",
		[]string{"name"},
		nil,
	)

	c.FolderDeletedSpaceInUse = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_deleted_space_in_use_bytes"),
		"Total size (in bytes) of the deleted files and folders currently in the Conflict and Deleted folder",
		[]string{"name"},
		nil,
	)

	c.FolderDeletedBytesCleanedUpTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_deleted_cleaned_up_bytes_total"),
		"Total size (in bytes) of replicating deleted files and folders that were cleaned up from the Conflict and Deleted folder",
		[]string{"name"},
		nil,
	)

	c.FolderDeletedBytesGeneratedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_deleted_generated_bytes_total"),
		"Total size (in bytes) of replicated deleted files and folders that were moved to the Conflict and Deleted folder after they were deleted from a replicated folder on a sending member",
		[]string{"name"},
		nil,
	)

	c.FolderDeletedFilesCleanedUpTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_deleted_cleaned_up_files_total"),
		"Number of files and folders that were cleaned up from the Conflict and Deleted folder",
		[]string{"name"},
		nil,
	)

	c.FolderDeletedFilesGeneratedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_deleted_generated_files_total"),
		"Number of deleted files and folders that were moved to the Conflict and Deleted folder",
		[]string{"name"},
		nil,
	)

	c.FolderFileInstallsRetriedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_file_installs_retried_total"),
		"Total number of file installs that are being retried due to sharing violations or other errors encountered when installing the files",
		[]string{"name"},
		nil,
	)

	c.FolderFileInstallsSucceededTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_file_installs_succeeded_total"),
		"Total number of files that were successfully received from sending members and installed locally on this server",
		[]string{"name"},
		nil,
	)

	c.FolderFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_received_files_total"),
		"Total number of files received",
		[]string{"name"},
		nil,
	)

	c.FolderRDCBytesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_rdc_received_bytes_total"),
		"Total number of bytes received in replicating files using Remote Differential Compression",
		[]string{"name"},
		nil,
	)

	c.FolderRDCCompressedSizeOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_rdc_compressed_size_of_received_files_bytes_total"),
		"Total compressed size (in bytes) of the files received with Remote Differential Compression",
		[]string{"name"},
		nil,
	)

	c.FolderRDCNumberofFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_rdc_received_files_total"),
		"Total number of files received with Remote Differential Compression",
		[]string{"name"},
		nil,
	)

	c.FolderRDCSizeOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_rdc_files_received_bytes_total"),
		"Total uncompressed size (in bytes) of the files received with Remote Differential Compression",
		[]string{"name"},
		nil,
	)

	c.FolderSizeOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_files_received_bytes_total"),
		"Total uncompressed size (in bytes) of the files received",
		[]string{"name"},
		nil,
	)

	c.FolderStagingSpaceInUse = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_staging_space_in_use_bytes"),
		"Total size of files and folders currently in the staging folder.",
		[]string{"name"},
		nil,
	)

	c.FolderStagingBytesCleanedUpTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_staging_cleaned_up_bytes_total"),
		"Total size (in bytes) of the files and folders that have been cleaned up from the staging folder",
		[]string{"name"},
		nil,
	)

	c.FolderStagingBytesGeneratedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_staging_generated_bytes_total"),
		"Total size (in bytes) of replicated files and folders in the staging folder created by the DFS Replication service since last restart",
		[]string{"name"},
		nil,
	)

	c.FolderStagingFilesCleanedUpTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_staging_cleaned_up_files_total"),
		"Total number of files and folders that have been cleaned up from the staging folder",
		[]string{"name"},
		nil,
	)

	c.FolderStagingFilesGeneratedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_staging_generated_files_total"),
		"Total number of times replicated files and folders have been staged by the DFS Replication service",
		[]string{"name"},
		nil,
	)

	c.FolderUpdatesDroppedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_dropped_updates_total"),
		"Total number of redundant file replication update records that have been ignored by the DFS Replication service because they did not change the replicated file or folder",
		[]string{"name"},
		nil,
	)

	c. // Volume
		VolumeDatabaseCommitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "volume_database_commits_total"),
		"Total number of DFSR Volume database commits",
		[]string{"name"},
		nil,
	)

	c.VolumeDatabaseLookupsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "volume_database_lookups_total"),
		"Total number of DFSR Volume database lookups",
		[]string{"name"},
		nil,
	)

	c.VolumeUSNJournalUnreadPercentage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "volume_usn_journal_unread_percentage"),
		"Percentage of DFSR Volume USN journal records that are unread",
		[]string{"name"},
		nil,
	)

	c.VolumeUSNJournalRecordsAcceptedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "volume_usn_journal_accepted_records_total"),
		"Total number of USN journal records accepted",
		[]string{"name"},
		nil,
	)

	c.VolumeUSNJournalRecordsReadTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "volume_usn_journal_read_records_total"),
		"Total number of DFSR Volume USN journal records read",
		[]string{"name"},
		nil,
	)

	c.dfsrChildCollectors = c.getDFSRChildCollectors(utils.ExpandEnabledChildCollectors(*c.dfsrEnabledCollectors))

	return nil
}

// Maps enabled child collectors names to their relevant collection function,
// for use in collector.Collect()
func (c *collector) getDFSRChildCollectors(enabledCollectors []string) []dfsrCollectorFunc {
	var dfsrCollectors []dfsrCollectorFunc
	for _, collector := range enabledCollectors {
		switch collector {
		case "connection":
			dfsrCollectors = append(dfsrCollectors, c.collectConnection)
		case "folder":
			dfsrCollectors = append(dfsrCollectors, c.collectFolder)
		case "volume":
			dfsrCollectors = append(dfsrCollectors, c.collectVolume)
		}
	}

	return dfsrCollectors
}

// Collect implements the Collector interface.
// Sends metric values for each metric to the provided prometheus Metric channel.
func (c *collector) Collect(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	for _, fn := range c.dfsrChildCollectors {
		err := fn(ctx, ch)
		if err != nil {
			return err
		}
	}
	return nil
}

// PerflibDFSRConnection Perflib: "DFS Replication Service Connections"
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

func (c *collector) collectConnection(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []PerflibDFSRConnection
	if err := perflib.UnmarshalObject(ctx.PerfObjects["DFS Replication Connections"], &dst, c.logger); err != nil {
		return err
	}

	for _, connection := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.ConnectionBandwidthSavingsUsingDFSReplicationTotal,
			prometheus.CounterValue,
			connection.BandwidthSavingsUsingDFSReplicationTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConnectionBytesReceivedTotal,
			prometheus.CounterValue,
			connection.BytesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConnectionCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.CompressedSizeOfFilesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConnectionFilesReceivedTotal,
			prometheus.CounterValue,
			connection.FilesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConnectionRDCBytesReceivedTotal,
			prometheus.CounterValue,
			connection.RDCBytesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConnectionRDCCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.RDCCompressedSizeOfFilesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConnectionRDCSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.RDCSizeOfFilesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConnectionRDCNumberofFilesReceivedTotal,
			prometheus.CounterValue,
			connection.RDCNumberofFilesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.ConnectionSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.SizeOfFilesReceivedTotal,
			connection.Name,
		)

	}
	return nil
}

// PerflibDFSRFolder Perflib: "DFS Replicated Folder"
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

func (c *collector) collectFolder(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []PerflibDFSRFolder
	if err := perflib.UnmarshalObject(ctx.PerfObjects["DFS Replicated Folders"], &dst, c.logger); err != nil {
		return err
	}

	for _, folder := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.FolderBandwidthSavingsUsingDFSReplicationTotal,
			prometheus.CounterValue,
			folder.BandwidthSavingsUsingDFSReplicationTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.CompressedSizeOfFilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderConflictBytesCleanedupTotal,
			prometheus.CounterValue,
			folder.ConflictBytesCleanedupTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderConflictBytesGeneratedTotal,
			prometheus.CounterValue,
			folder.ConflictBytesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderConflictFilesCleanedUpTotal,
			prometheus.CounterValue,
			folder.ConflictFilesCleanedUpTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderConflictFilesGeneratedTotal,
			prometheus.CounterValue,
			folder.ConflictFilesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderConflictFolderCleanupsCompletedTotal,
			prometheus.CounterValue,
			folder.ConflictFolderCleanupsCompletedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderConflictSpaceInUse,
			prometheus.GaugeValue,
			folder.ConflictSpaceInUse,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderDeletedSpaceInUse,
			prometheus.GaugeValue,
			folder.DeletedSpaceInUse,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderDeletedBytesCleanedUpTotal,
			prometheus.CounterValue,
			folder.DeletedBytesCleanedUpTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderDeletedBytesGeneratedTotal,
			prometheus.CounterValue,
			folder.DeletedBytesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderDeletedFilesCleanedUpTotal,
			prometheus.CounterValue,
			folder.DeletedFilesCleanedUpTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderDeletedFilesGeneratedTotal,
			prometheus.CounterValue,
			folder.DeletedFilesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderFileInstallsRetriedTotal,
			prometheus.CounterValue,
			folder.FileInstallsRetriedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderFileInstallsSucceededTotal,
			prometheus.CounterValue,
			folder.FileInstallsSucceededTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderFilesReceivedTotal,
			prometheus.CounterValue,
			folder.FilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderRDCBytesReceivedTotal,
			prometheus.CounterValue,
			folder.RDCBytesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderRDCCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.RDCCompressedSizeOfFilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderRDCNumberofFilesReceivedTotal,
			prometheus.CounterValue,
			folder.RDCNumberofFilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderRDCSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.RDCSizeOfFilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.SizeOfFilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderStagingSpaceInUse,
			prometheus.GaugeValue,
			folder.StagingSpaceInUse,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderStagingBytesCleanedUpTotal,
			prometheus.CounterValue,
			folder.StagingBytesCleanedUpTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderStagingBytesGeneratedTotal,
			prometheus.CounterValue,
			folder.StagingBytesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderStagingFilesCleanedUpTotal,
			prometheus.CounterValue,
			folder.StagingFilesCleanedUpTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderStagingFilesGeneratedTotal,
			prometheus.CounterValue,
			folder.StagingFilesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.FolderUpdatesDroppedTotal,
			prometheus.CounterValue,
			folder.UpdatesDroppedTotal,
			folder.Name,
		)
	}
	return nil
}

// PerflibDFSRVolume Perflib: "DFS Replication Service Volumes"
type PerflibDFSRVolume struct {
	Name string

	DatabaseCommitsTotal           float64 `perflib:"Database Commits"`
	DatabaseLookupsTotal           float64 `perflib:"Database Lookups"`
	USNJournalRecordsReadTotal     float64 `perflib:"USN Journal Records Read"`
	USNJournalRecordsAcceptedTotal float64 `perflib:"USN Journal Records Accepted"`
	USNJournalUnreadPercentage     float64 `perflib:"USN Journal Records Unread Percentage"`
}

func (c *collector) collectVolume(ctx *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	var dst []PerflibDFSRVolume
	if err := perflib.UnmarshalObject(ctx.PerfObjects["DFS Replication Service Volumes"], &dst, c.logger); err != nil {
		return err
	}

	for _, volume := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.VolumeDatabaseLookupsTotal,
			prometheus.CounterValue,
			volume.DatabaseLookupsTotal,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VolumeDatabaseCommitsTotal,
			prometheus.CounterValue,
			volume.DatabaseCommitsTotal,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VolumeUSNJournalRecordsAcceptedTotal,
			prometheus.CounterValue,
			volume.USNJournalRecordsAcceptedTotal,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VolumeUSNJournalRecordsReadTotal,
			prometheus.CounterValue,
			volume.USNJournalRecordsReadTotal,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.VolumeUSNJournalUnreadPercentage,
			prometheus.GaugeValue,
			volume.USNJournalUnreadPercentage,
			volume.Name,
		)

	}
	return nil
}
