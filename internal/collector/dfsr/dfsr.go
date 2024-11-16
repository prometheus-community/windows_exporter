//go:build windows

package dfsr

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/perfdata"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "dfsr"

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

var ConfigDefaults = Config{
	CollectorsEnabled: []string{"connection", "folder", "volume"},
}

// Collector contains the metric and state data of the DFSR collectors.
type Collector struct {
	config Config

	perfDataCollectorConnection *perfdata.Collector
	perfDataCollectorFolder     *perfdata.Collector
	perfDataCollectorVolume     *perfdata.Collector

	// connection source
	connectionBandwidthSavingsUsingDFSReplicationTotal *prometheus.Desc
	connectionBytesReceivedTotal                       *prometheus.Desc
	connectionCompressedSizeOfFilesReceivedTotal       *prometheus.Desc
	connectionFilesReceivedTotal                       *prometheus.Desc
	connectionRDCBytesReceivedTotal                    *prometheus.Desc
	connectionRDCCompressedSizeOfFilesReceivedTotal    *prometheus.Desc
	connectionRDCSizeOfFilesReceivedTotal              *prometheus.Desc
	connectionRDCNumberOfFilesReceivedTotal            *prometheus.Desc
	connectionSizeOfFilesReceivedTotal                 *prometheus.Desc

	// folder source
	folderBandwidthSavingsUsingDFSReplicationTotal *prometheus.Desc
	folderCompressedSizeOfFilesReceivedTotal       *prometheus.Desc
	folderConflictBytesCleanedUpTotal              *prometheus.Desc
	folderConflictBytesGeneratedTotal              *prometheus.Desc
	folderConflictFilesCleanedUpTotal              *prometheus.Desc
	folderConflictFilesGeneratedTotal              *prometheus.Desc
	folderConflictFolderCleanupsCompletedTotal     *prometheus.Desc
	folderConflictSpaceInUse                       *prometheus.Desc
	folderDeletedSpaceInUse                        *prometheus.Desc
	folderDeletedBytesCleanedUpTotal               *prometheus.Desc
	folderDeletedBytesGeneratedTotal               *prometheus.Desc
	folderDeletedFilesCleanedUpTotal               *prometheus.Desc
	folderDeletedFilesGeneratedTotal               *prometheus.Desc
	folderFileInstallsRetriedTotal                 *prometheus.Desc
	folderFileInstallsSucceededTotal               *prometheus.Desc
	folderFilesReceivedTotal                       *prometheus.Desc
	folderRDCBytesReceivedTotal                    *prometheus.Desc
	folderRDCCompressedSizeOfFilesReceivedTotal    *prometheus.Desc
	folderRDCNumberOfFilesReceivedTotal            *prometheus.Desc
	folderRDCSizeOfFilesReceivedTotal              *prometheus.Desc
	folderSizeOfFilesReceivedTotal                 *prometheus.Desc
	folderStagingSpaceInUse                        *prometheus.Desc
	folderStagingBytesCleanedUpTotal               *prometheus.Desc
	folderStagingBytesGeneratedTotal               *prometheus.Desc
	folderStagingFilesCleanedUpTotal               *prometheus.Desc
	folderStagingFilesGeneratedTotal               *prometheus.Desc
	folderUpdatesDroppedTotal                      *prometheus.Desc

	// volume source
	volumeDatabaseLookupsTotal           *prometheus.Desc
	volumeDatabaseCommitsTotal           *prometheus.Desc
	volumeUSNJournalUnreadPercentage     *prometheus.Desc
	volumeUSNJournalRecordsAcceptedTotal *prometheus.Desc
	volumeUSNJournalRecordsReadTotal     *prometheus.Desc
}

func New(config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	if config.CollectorsEnabled == nil {
		config.CollectorsEnabled = ConfigDefaults.CollectorsEnabled
	}

	c := &Collector{
		config: *config,
	}

	return c
}

func NewWithFlags(app *kingpin.Application) *Collector {
	c := &Collector{
		config: ConfigDefaults,
	}
	c.config.CollectorsEnabled = make([]string, 0)

	var collectorsEnabled string

	app.Flag("collector.dfsr.sources-enabled", "Comma-separated list of DFSR Perflib sources to use.").
		Default(strings.Join(ConfigDefaults.CollectorsEnabled, ",")).StringVar(&collectorsEnabled)

	app.Action(func(*kingpin.ParseContext) error {
		c.config.CollectorsEnabled = strings.Split(collectorsEnabled, ",")

		return nil
	})

	return c
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) Close() error {
	if slices.Contains(c.config.CollectorsEnabled, "connection") {
		c.perfDataCollectorConnection.Close()
	}

	if slices.Contains(c.config.CollectorsEnabled, "folder") {
		c.perfDataCollectorFolder.Close()
	}

	if slices.Contains(c.config.CollectorsEnabled, "volume") {
		c.perfDataCollectorVolume.Close()
	}

	return nil
}

func (c *Collector) Build(logger *slog.Logger, _ *mi.Session) error {
	logger = logger.With(slog.String("collector", Name))

	logger.Info("dfsr collector is in an experimental state! Metrics for this collector have not been tested.")

	var err error

	if slices.Contains(c.config.CollectorsEnabled, "connection") {
		c.perfDataCollectorConnection, err = perfdata.NewCollector("DFS Replication Connections", perfdata.InstanceAll, []string{
			bandwidthSavingsUsingDFSReplicationTotal,
			bytesReceivedTotal,
			compressedSizeOfFilesReceivedTotal,
			filesReceivedTotal,
			rdcBytesReceivedTotal,
			rdcCompressedSizeOfFilesReceivedTotal,
			rdcNumberOfFilesReceivedTotal,
			rdcSizeOfFilesReceivedTotal,
			sizeOfFilesReceivedTotal,
		})
		if err != nil {
			return fmt.Errorf("failed to create DFS Replication Connections collector: %w", err)
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, "folder") {
		c.perfDataCollectorFolder, err = perfdata.NewCollector("DFS Replicated Folders", perfdata.InstanceAll, []string{
			bandwidthSavingsUsingDFSReplicationTotal,
			compressedSizeOfFilesReceivedTotal,
			conflictBytesCleanedUpTotal,
			conflictBytesGeneratedTotal,
			conflictFilesCleanedUpTotal,
			conflictFilesGeneratedTotal,
			conflictFolderCleanupsCompletedTotal,
			conflictSpaceInUse,
			deletedSpaceInUse,
			deletedBytesCleanedUpTotal,
			deletedBytesGeneratedTotal,
			deletedFilesCleanedUpTotal,
			deletedFilesGeneratedTotal,
			fileInstallsRetriedTotal,
			fileInstallsSucceededTotal,
			filesReceivedTotal,
			rdcBytesReceivedTotal,
			rdcCompressedSizeOfFilesReceivedTotal,
			rdcNumberOfFilesReceivedTotal,
			rdcSizeOfFilesReceivedTotal,
			sizeOfFilesReceivedTotal,
			stagingSpaceInUse,
			stagingBytesCleanedUpTotal,
			stagingBytesGeneratedTotal,
			stagingFilesCleanedUpTotal,
			stagingFilesGeneratedTotal,
			updatesDroppedTotal,
		})
		if err != nil {
			return fmt.Errorf("failed to create DFS Replicated Folders collector: %w", err)
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, "volume") {
		c.perfDataCollectorVolume, err = perfdata.NewCollector("DFS Replication Service Volumes", perfdata.InstanceAll, []string{
			databaseCommitsTotal,
			databaseLookupsTotal,
			usnJournalRecordsReadTotal,
			usnJournalRecordsAcceptedTotal,
			usnJournalUnreadPercentage,
		})
		if err != nil {
			return fmt.Errorf("failed to create DFS Replication Service Volumes collector: %w", err)
		}
	}
	// connection
	c.connectionBandwidthSavingsUsingDFSReplicationTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_bandwidth_savings_using_dfs_replication_bytes_total"),
		"Total bytes of bandwidth saved using DFS Replication for this connection",
		[]string{"name"},
		nil,
	)

	c.connectionBytesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_bytes_received_total"),
		"Total bytes received for connection",
		[]string{"name"},
		nil,
	)

	c.connectionCompressedSizeOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_compressed_size_of_files_received_bytes_total"),
		"Total compressed size of files received on the connection, in bytes",
		[]string{"name"},
		nil,
	)

	c.connectionFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_received_files_total"),
		"Total number of files received for connection",
		[]string{"name"},
		nil,
	)

	c.connectionRDCBytesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_rdc_received_bytes_total"),
		"Total bytes received on the connection while replicating files using Remote Differential Compression",
		[]string{"name"},
		nil,
	)

	c.connectionRDCCompressedSizeOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_rdc_compressed_size_of_received_files_bytes_total"),
		"Total uncompressed size of files received with Remote Differential Compression for connection",
		[]string{"name"},
		nil,
	)

	c.connectionRDCNumberOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_rdc_received_files_total"),
		"Total number of files received using remote differential compression",
		[]string{"name"},
		nil,
	)

	c.connectionRDCSizeOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_rdc_size_of_received_files_bytes_total"),
		"Total size of received Remote Differential Compression files, in bytes.",
		[]string{"name"},
		nil,
	)

	c.connectionSizeOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "connection_files_received_bytes_total"),
		"Total size of files received, in bytes",
		[]string{"name"},
		nil,
	)

	// folder
	c.folderBandwidthSavingsUsingDFSReplicationTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_bandwidth_savings_using_dfs_replication_bytes_total"),
		"Total bytes of bandwidth saved using DFS Replication for this folder",
		[]string{"name"},
		nil,
	)

	c.folderCompressedSizeOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_compressed_size_of_received_files_bytes_total"),
		"Total compressed size of files received on the folder, in bytes",
		[]string{"name"},
		nil,
	)

	c.folderConflictBytesCleanedUpTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_conflict_cleaned_up_bytes_total"),
		"Total size of conflict loser files and folders deleted from the Conflict and Deleted folder, in bytes",
		[]string{"name"},
		nil,
	)

	c.folderConflictBytesGeneratedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_conflict_generated_bytes_total"),
		"Total size of conflict loser files and folders moved to the Conflict and Deleted folder, in bytes",
		[]string{"name"},
		nil,
	)

	c.folderConflictFilesCleanedUpTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_conflict_cleaned_up_files_total"),
		"Number of conflict loser files deleted from the Conflict and Deleted folder",
		[]string{"name"},
		nil,
	)

	c.folderConflictFilesGeneratedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_conflict_generated_files_total"),
		"Number of files and folders moved to the Conflict and Deleted folder",
		[]string{"name"},
		nil,
	)

	c.folderConflictFolderCleanupsCompletedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_conflict_folder_cleanups_total"),
		"Number of deletions of conflict loser files and folders in the Conflict and Deleted",
		[]string{"name"},
		nil,
	)

	c.folderConflictSpaceInUse = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_conflict_space_in_use_bytes"),
		"Total size of the conflict loser files and folders currently in the Conflict and Deleted folder",
		[]string{"name"},
		nil,
	)

	c.folderDeletedSpaceInUse = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_deleted_space_in_use_bytes"),
		"Total size (in bytes) of the deleted files and folders currently in the Conflict and Deleted folder",
		[]string{"name"},
		nil,
	)

	c.folderDeletedBytesCleanedUpTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_deleted_cleaned_up_bytes_total"),
		"Total size (in bytes) of replicating deleted files and folders that were cleaned up from the Conflict and Deleted folder",
		[]string{"name"},
		nil,
	)

	c.folderDeletedBytesGeneratedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_deleted_generated_bytes_total"),
		"Total size (in bytes) of replicated deleted files and folders that were moved to the Conflict and Deleted folder after they were deleted from a replicated folder on a sending member",
		[]string{"name"},
		nil,
	)

	c.folderDeletedFilesCleanedUpTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_deleted_cleaned_up_files_total"),
		"Number of files and folders that were cleaned up from the Conflict and Deleted folder",
		[]string{"name"},
		nil,
	)

	c.folderDeletedFilesGeneratedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_deleted_generated_files_total"),
		"Number of deleted files and folders that were moved to the Conflict and Deleted folder",
		[]string{"name"},
		nil,
	)

	c.folderFileInstallsRetriedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_file_installs_retried_total"),
		"Total number of file installs that are being retried due to sharing violations or other errors encountered when installing the files",
		[]string{"name"},
		nil,
	)

	c.folderFileInstallsSucceededTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_file_installs_succeeded_total"),
		"Total number of files that were successfully received from sending members and installed locally on this server",
		[]string{"name"},
		nil,
	)

	c.folderFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_received_files_total"),
		"Total number of files received",
		[]string{"name"},
		nil,
	)

	c.folderRDCBytesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_rdc_received_bytes_total"),
		"Total number of bytes received in replicating files using Remote Differential Compression",
		[]string{"name"},
		nil,
	)

	c.folderRDCCompressedSizeOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_rdc_compressed_size_of_received_files_bytes_total"),
		"Total compressed size (in bytes) of the files received with Remote Differential Compression",
		[]string{"name"},
		nil,
	)

	c.folderRDCNumberOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_rdc_received_files_total"),
		"Total number of files received with Remote Differential Compression",
		[]string{"name"},
		nil,
	)

	c.folderRDCSizeOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_rdc_files_received_bytes_total"),
		"Total uncompressed size (in bytes) of the files received with Remote Differential Compression",
		[]string{"name"},
		nil,
	)

	c.folderSizeOfFilesReceivedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_files_received_bytes_total"),
		"Total uncompressed size (in bytes) of the files received",
		[]string{"name"},
		nil,
	)

	c.folderStagingSpaceInUse = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_staging_space_in_use_bytes"),
		"Total size of files and folders currently in the staging folder.",
		[]string{"name"},
		nil,
	)

	c.folderStagingBytesCleanedUpTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_staging_cleaned_up_bytes_total"),
		"Total size (in bytes) of the files and folders that have been cleaned up from the staging folder",
		[]string{"name"},
		nil,
	)

	c.folderStagingBytesGeneratedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_staging_generated_bytes_total"),
		"Total size (in bytes) of replicated files and folders in the staging folder created by the DFS Replication service since last restart",
		[]string{"name"},
		nil,
	)

	c.folderStagingFilesCleanedUpTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_staging_cleaned_up_files_total"),
		"Total number of files and folders that have been cleaned up from the staging folder",
		[]string{"name"},
		nil,
	)

	c.folderStagingFilesGeneratedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_staging_generated_files_total"),
		"Total number of times replicated files and folders have been staged by the DFS Replication service",
		[]string{"name"},
		nil,
	)

	c.folderUpdatesDroppedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "folder_dropped_updates_total"),
		"Total number of redundant file replication update records that have been ignored by the DFS Replication service because they did not change the replicated file or folder",
		[]string{"name"},
		nil,
	)

	// volume
	c.volumeDatabaseCommitsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "volume_database_commits_total"),
		"Total number of DFSR volume database commits",
		[]string{"name"},
		nil,
	)

	c.volumeDatabaseLookupsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "volume_database_lookups_total"),
		"Total number of DFSR volume database lookups",
		[]string{"name"},
		nil,
	)

	c.volumeUSNJournalUnreadPercentage = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "volume_usn_journal_unread_percentage"),
		"Percentage of DFSR volume USN journal records that are unread",
		[]string{"name"},
		nil,
	)

	c.volumeUSNJournalRecordsAcceptedTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "volume_usn_journal_accepted_records_total"),
		"Total number of USN journal records accepted",
		[]string{"name"},
		nil,
	)

	c.volumeUSNJournalRecordsReadTotal = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "volume_usn_journal_read_records_total"),
		"Total number of DFSR volume USN journal records read",
		[]string{"name"},
		nil,
	)

	return nil
}

// Collect implements the Collector interface.
// Sends metric values for each metric to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0, 3)

	if slices.Contains(c.config.CollectorsEnabled, "connection") {
		errs = append(errs, c.collectPDHConnection(ch))
	}

	if slices.Contains(c.config.CollectorsEnabled, "folder") {
		errs = append(errs, c.collectPDHFolder(ch))
	}

	if slices.Contains(c.config.CollectorsEnabled, "volume") {
		errs = append(errs, c.collectPDHVolume(ch))
	}

	return errors.Join(errs...)
}

func (c *Collector) collectPDHConnection(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorConnection.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect DFS Replication Connections metrics: %w", err)
	}

	if len(perfData) == 0 {
		return errors.New("perflib query for DFS Replication Connections returned empty result set")
	}

	for name, connection := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.connectionBandwidthSavingsUsingDFSReplicationTotal,
			prometheus.CounterValue,
			connection[bandwidthSavingsUsingDFSReplicationTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionBytesReceivedTotal,
			prometheus.CounterValue,
			connection[bytesReceivedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection[compressedSizeOfFilesReceivedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionFilesReceivedTotal,
			prometheus.CounterValue,
			connection[filesReceivedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionRDCBytesReceivedTotal,
			prometheus.CounterValue,
			connection[rdcBytesReceivedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionRDCCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection[rdcCompressedSizeOfFilesReceivedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionRDCSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection[rdcSizeOfFilesReceivedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionRDCNumberOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection[rdcNumberOfFilesReceivedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection[sizeOfFilesReceivedTotal].FirstValue,
			name,
		)
	}

	return nil
}

func (c *Collector) collectPDHFolder(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorFolder.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect DFS Replicated Folders metrics: %w", err)
	}

	if len(perfData) == 0 {
		return errors.New("perflib query for DFS Replicated Folders returned empty result set")
	}

	for name, folder := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.folderBandwidthSavingsUsingDFSReplicationTotal,
			prometheus.CounterValue,
			folder[bandwidthSavingsUsingDFSReplicationTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder[compressedSizeOfFilesReceivedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictBytesCleanedUpTotal,
			prometheus.CounterValue,
			folder[conflictBytesCleanedUpTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictBytesGeneratedTotal,
			prometheus.CounterValue,
			folder[conflictBytesGeneratedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictFilesCleanedUpTotal,
			prometheus.CounterValue,
			folder[conflictFilesCleanedUpTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictFilesGeneratedTotal,
			prometheus.CounterValue,
			folder[conflictFilesGeneratedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictFolderCleanupsCompletedTotal,
			prometheus.CounterValue,
			folder[conflictFolderCleanupsCompletedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictSpaceInUse,
			prometheus.GaugeValue,
			folder[conflictSpaceInUse].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderDeletedSpaceInUse,
			prometheus.GaugeValue,
			folder[deletedSpaceInUse].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderDeletedBytesCleanedUpTotal,
			prometheus.CounterValue,
			folder[deletedBytesCleanedUpTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderDeletedBytesGeneratedTotal,
			prometheus.CounterValue,
			folder[deletedBytesGeneratedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderDeletedFilesCleanedUpTotal,
			prometheus.CounterValue,
			folder[deletedFilesCleanedUpTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderDeletedFilesGeneratedTotal,
			prometheus.CounterValue,
			folder[deletedFilesGeneratedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderFileInstallsRetriedTotal,
			prometheus.CounterValue,
			folder[fileInstallsRetriedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderFileInstallsSucceededTotal,
			prometheus.CounterValue,
			folder[fileInstallsSucceededTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderFilesReceivedTotal,
			prometheus.CounterValue,
			folder[filesReceivedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderRDCBytesReceivedTotal,
			prometheus.CounterValue,
			folder[rdcBytesReceivedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderRDCCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder[rdcCompressedSizeOfFilesReceivedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderRDCNumberOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder[rdcNumberOfFilesReceivedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderRDCSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder[rdcSizeOfFilesReceivedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder[sizeOfFilesReceivedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderStagingSpaceInUse,
			prometheus.GaugeValue,
			folder[stagingSpaceInUse].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderStagingBytesCleanedUpTotal,
			prometheus.CounterValue,
			folder[stagingBytesCleanedUpTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderStagingBytesGeneratedTotal,
			prometheus.CounterValue,
			folder[stagingBytesGeneratedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderStagingFilesCleanedUpTotal,
			prometheus.CounterValue,
			folder[stagingFilesCleanedUpTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderStagingFilesGeneratedTotal,
			prometheus.CounterValue,
			folder[stagingFilesGeneratedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderUpdatesDroppedTotal,
			prometheus.CounterValue,
			folder[updatesDroppedTotal].FirstValue,
			name,
		)
	}

	return nil
}

func (c *Collector) collectPDHVolume(ch chan<- prometheus.Metric) error {
	perfData, err := c.perfDataCollectorVolume.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect DFS Replication Volumes metrics: %w", err)
	}

	if len(perfData) == 0 {
		return errors.New("perflib query for DFS Replication Volumes returned empty result set")
	}

	for name, volume := range perfData {
		ch <- prometheus.MustNewConstMetric(
			c.volumeDatabaseLookupsTotal,
			prometheus.CounterValue,
			volume[databaseLookupsTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.volumeDatabaseCommitsTotal,
			prometheus.CounterValue,
			volume[databaseCommitsTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.volumeUSNJournalRecordsAcceptedTotal,
			prometheus.CounterValue,
			volume[usnJournalRecordsAcceptedTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.volumeUSNJournalRecordsReadTotal,
			prometheus.CounterValue,
			volume[usnJournalRecordsReadTotal].FirstValue,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.volumeUSNJournalUnreadPercentage,
			prometheus.GaugeValue,
			volume[usnJournalUnreadPercentage].FirstValue,
			name,
		)
	}

	return nil
}
