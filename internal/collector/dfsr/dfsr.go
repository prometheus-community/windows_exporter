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

package dfsr

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus-community/windows_exporter/internal/mi"
	"github.com/prometheus-community/windows_exporter/internal/pdh"
	"github.com/prometheus-community/windows_exporter/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "dfsr"

type Config struct {
	CollectorsEnabled []string `yaml:"collectors_enabled"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	CollectorsEnabled: []string{"connection", "folder", "volume"},
}

// Collector contains the metric and state data of the DFSR collectors.
type Collector struct {
	config Config

	perfDataCollectorConnection *pdh.Collector
	perfDataCollectorFolder     *pdh.Collector
	perfDataCollectorVolume     *pdh.Collector
	perfDataObjectConnection    []perfDataCounterValuesConnection
	perfDataObjectFolder        []perfDataCounterValuesFolder
	perfDataObjectVolume        []perfDataCounterValuesVolume

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

	var err error

	if slices.Contains(c.config.CollectorsEnabled, "connection") {
		c.perfDataCollectorConnection, err = pdh.NewCollector[perfDataCounterValuesConnection](pdh.CounterTypeRaw, "DFS Replication Connections", pdh.InstancesAll)
		if err != nil {
			return fmt.Errorf("failed to create DFS Replication Connections collector: %w", err)
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, "folder") {
		c.perfDataCollectorFolder, err = pdh.NewCollector[perfDataCounterValuesFolder](pdh.CounterTypeRaw, "DFS Replicated Folders", pdh.InstancesAll)
		if err != nil {
			return fmt.Errorf("failed to create DFS Replicated Folders collector: %w", err)
		}
	}

	if slices.Contains(c.config.CollectorsEnabled, "volume") {
		c.perfDataCollectorVolume, err = pdh.NewCollector[perfDataCounterValuesVolume](pdh.CounterTypeRaw, "DFS Replication Service Volumes", pdh.InstancesAll)
		if err != nil {
			return fmt.Errorf("failed to create DFS Replication Service Volumes collector: %w", err)
		}
	}

	return nil
}

// Collect implements the Collector interface.
// Sends metric values for each metric to the provided prometheus Metric channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) error {
	errs := make([]error, 0)

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
	err := c.perfDataCollectorConnection.Collect(&c.perfDataObjectConnection)
	if err != nil {
		return fmt.Errorf("failed to collect DFS Replication Connections metrics: %w", err)
	}

	for _, connection := range c.perfDataObjectConnection {
		name := connection.Name

		ch <- prometheus.MustNewConstMetric(
			c.connectionBandwidthSavingsUsingDFSReplicationTotal,
			prometheus.CounterValue,
			connection.BandwidthSavingsUsingDFSReplicationTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionBytesReceivedTotal,
			prometheus.CounterValue,
			connection.BytesReceivedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.CompressedSizeOfFilesReceivedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionFilesReceivedTotal,
			prometheus.CounterValue,
			connection.FilesReceivedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionRDCBytesReceivedTotal,
			prometheus.CounterValue,
			connection.RdcBytesReceivedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionRDCCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.RdcCompressedSizeOfFilesReceivedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionRDCSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.RdcSizeOfFilesReceivedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionRDCNumberOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.RdcNumberOfFilesReceivedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.SizeOfFilesReceivedTotal,
			name,
		)
	}

	return nil
}

func (c *Collector) collectPDHFolder(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorFolder.Collect(&c.perfDataObjectFolder)
	if err != nil {
		return fmt.Errorf("failed to collect DFS Replicated Folders metrics: %w", err)
	}

	for _, folder := range c.perfDataObjectFolder {
		name := folder.Name

		ch <- prometheus.MustNewConstMetric(
			c.folderBandwidthSavingsUsingDFSReplicationTotal,
			prometheus.CounterValue,
			folder.BandwidthSavingsUsingDFSReplicationTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.CompressedSizeOfFilesReceivedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictBytesCleanedUpTotal,
			prometheus.CounterValue,
			folder.ConflictBytesCleanedUpTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictBytesGeneratedTotal,
			prometheus.CounterValue,
			folder.ConflictBytesGeneratedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictFilesCleanedUpTotal,
			prometheus.CounterValue,
			folder.ConflictFilesCleanedUpTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictFilesGeneratedTotal,
			prometheus.CounterValue,
			folder.ConflictFilesGeneratedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictFolderCleanupsCompletedTotal,
			prometheus.CounterValue,
			folder.ConflictFolderCleanupsCompletedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictSpaceInUse,
			prometheus.GaugeValue,
			folder.ConflictSpaceInUse,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderDeletedSpaceInUse,
			prometheus.GaugeValue,
			folder.DeletedSpaceInUse,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderDeletedBytesCleanedUpTotal,
			prometheus.CounterValue,
			folder.DeletedBytesCleanedUpTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderDeletedBytesGeneratedTotal,
			prometheus.CounterValue,
			folder.DeletedBytesGeneratedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderDeletedFilesCleanedUpTotal,
			prometheus.CounterValue,
			folder.DeletedFilesCleanedUpTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderDeletedFilesGeneratedTotal,
			prometheus.CounterValue,
			folder.DeletedFilesGeneratedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderFileInstallsRetriedTotal,
			prometheus.CounterValue,
			folder.FileInstallsRetriedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderFileInstallsSucceededTotal,
			prometheus.CounterValue,
			folder.FileInstallsSucceededTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderFilesReceivedTotal,
			prometheus.CounterValue,
			folder.FilesReceivedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderRDCBytesReceivedTotal,
			prometheus.CounterValue,
			folder.RdcBytesReceivedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderRDCCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.RdcCompressedSizeOfFilesReceivedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderRDCNumberOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.RdcNumberOfFilesReceivedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderRDCSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.RdcSizeOfFilesReceivedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.SizeOfFilesReceivedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderStagingSpaceInUse,
			prometheus.GaugeValue,
			folder.StagingSpaceInUse,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderStagingBytesCleanedUpTotal,
			prometheus.CounterValue,
			folder.StagingBytesCleanedUpTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderStagingBytesGeneratedTotal,
			prometheus.CounterValue,
			folder.StagingBytesGeneratedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderStagingFilesCleanedUpTotal,
			prometheus.CounterValue,
			folder.StagingFilesCleanedUpTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderStagingFilesGeneratedTotal,
			prometheus.CounterValue,
			folder.StagingFilesGeneratedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderUpdatesDroppedTotal,
			prometheus.CounterValue,
			folder.UpdatesDroppedTotal,
			name,
		)
	}

	return nil
}

func (c *Collector) collectPDHVolume(ch chan<- prometheus.Metric) error {
	err := c.perfDataCollectorVolume.Collect(&c.perfDataObjectVolume)
	if err != nil {
		return fmt.Errorf("failed to collect DFS Replication Volumes metrics: %w", err)
	}

	for _, volume := range c.perfDataObjectVolume {
		name := volume.Name
		ch <- prometheus.MustNewConstMetric(
			c.volumeDatabaseLookupsTotal,
			prometheus.CounterValue,
			volume.DatabaseLookupsTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.volumeDatabaseCommitsTotal,
			prometheus.CounterValue,
			volume.DatabaseCommitsTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.volumeUSNJournalRecordsAcceptedTotal,
			prometheus.CounterValue,
			volume.UsnJournalRecordsAcceptedTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.volumeUSNJournalRecordsReadTotal,
			prometheus.CounterValue,
			volume.UsnJournalRecordsReadTotal,
			name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.volumeUSNJournalUnreadPercentage,
			prometheus.GaugeValue,
			volume.UsnJournalUnreadPercentage,
			name,
		)
	}

	return nil
}
