//go:build windows

package dfsr

import (
	"slices"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/perflib"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yusufpapurcu/wmi"
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

	// connection source
	connectionBandwidthSavingsUsingDFSReplicationTotal *prometheus.Desc
	connectionBytesReceivedTotal                       *prometheus.Desc
	connectionCompressedSizeOfFilesReceivedTotal       *prometheus.Desc
	connectionFilesReceivedTotal                       *prometheus.Desc
	connectionRDCBytesReceivedTotal                    *prometheus.Desc
	connectionRDCCompressedSizeOfFilesReceivedTotal    *prometheus.Desc
	connectionRDCSizeOfFilesReceivedTotal              *prometheus.Desc
	connectionRDCNumberofFilesReceivedTotal            *prometheus.Desc
	connectionSizeOfFilesReceivedTotal                 *prometheus.Desc

	// folder source
	folderBandwidthSavingsUsingDFSReplicationTotal *prometheus.Desc
	folderCompressedSizeOfFilesReceivedTotal       *prometheus.Desc
	folderConflictBytesCleanedupTotal              *prometheus.Desc
	folderConflictBytesGeneratedTotal              *prometheus.Desc
	folderConflictFilesCleanedUpTotal              *prometheus.Desc
	folderConflictFilesGeneratedTotal              *prometheus.Desc
	folderConflictfolderCleanupsCompletedTotal     *prometheus.Desc
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
	folderRDCNumberofFilesReceivedTotal            *prometheus.Desc
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

	// Map of child Collector functions used during collection
	dfsrChildCollectors []dfsrCollectorFunc
}

type dfsrCollectorFunc func(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error

// Map Perflib sources to DFSR Collector names
// e.g, volume -> DFS Replication Service Volumes.
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

	var collectorsEnabled string

	app.Flag("collectors.dfsr.sources-enabled", "Comma-separated list of DFSR Perflib sources to use.").
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

func (c *Collector) GetPerfCounter(_ log.Logger) ([]string, error) {
	// Perflib sources are dynamic, depending on the enabled child collectors
	expandedChildCollectors := slices.Compact(c.config.CollectorsEnabled)
	perflibDependencies := make([]string, 0, len(expandedChildCollectors))

	for _, source := range expandedChildCollectors {
		perflibDependencies = append(perflibDependencies, dfsrGetPerfObjectName(source))
	}

	return perflibDependencies, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build(logger log.Logger, _ *wmi.Client) error {
	logger = log.With(logger, "collector", Name)

	_ = level.Info(logger).Log("msg", "dfsr collector is in an experimental state! Metrics for this collector have not been tested.")

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

	c.connectionRDCNumberofFilesReceivedTotal = prometheus.NewDesc(
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

	c.folderConflictBytesCleanedupTotal = prometheus.NewDesc(
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

	c.folderConflictfolderCleanupsCompletedTotal = prometheus.NewDesc(
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

	c.folderRDCNumberofFilesReceivedTotal = prometheus.NewDesc(
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

	// Perflib sources are dynamic, depending on the enabled child collectors
	expandedChildCollectors := slices.Compact(c.config.CollectorsEnabled)
	c.dfsrChildCollectors = c.getDFSRChildCollectors(expandedChildCollectors)

	return nil
}

// Maps enabled child collectors names to their relevant collection function,
// for use in Collector.Collect().
func (c *Collector) getDFSRChildCollectors(enabledCollectors []string) []dfsrCollectorFunc {
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
func (c *Collector) Collect(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	for _, fn := range c.dfsrChildCollectors {
		err := fn(ctx, logger, ch)
		if err != nil {
			return err
		}
	}
	return nil
}

// PerflibDFSRConnection Perflib: "DFS Replication Service Connections".
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

func (c *Collector) collectConnection(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var dst []PerflibDFSRConnection
	if err := perflib.UnmarshalObject(ctx.PerfObjects["DFS Replication Connections"], &dst, logger); err != nil {
		return err
	}

	for _, connection := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.connectionBandwidthSavingsUsingDFSReplicationTotal,
			prometheus.CounterValue,
			connection.BandwidthSavingsUsingDFSReplicationTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionBytesReceivedTotal,
			prometheus.CounterValue,
			connection.BytesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.CompressedSizeOfFilesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionFilesReceivedTotal,
			prometheus.CounterValue,
			connection.FilesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionRDCBytesReceivedTotal,
			prometheus.CounterValue,
			connection.RDCBytesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionRDCCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.RDCCompressedSizeOfFilesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionRDCSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.RDCSizeOfFilesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionRDCNumberofFilesReceivedTotal,
			prometheus.CounterValue,
			connection.RDCNumberofFilesReceivedTotal,
			connection.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.connectionSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			connection.SizeOfFilesReceivedTotal,
			connection.Name,
		)
	}
	return nil
}

// perflibDFSRFolder Perflib: "DFS Replicated Folder".
type perflibDFSRFolder struct {
	Name string

	BandwidthSavingsUsingDFSReplicationTotal float64 `perflib:"Bandwidth Savings Using DFS Replication"`
	CompressedSizeOfFilesReceivedTotal       float64 `perflib:"Compressed Size of Files Received"`
	ConflictBytesCleanedupTotal              float64 `perflib:"Conflict Bytes Cleaned Up"`
	ConflictBytesGeneratedTotal              float64 `perflib:"Conflict Bytes Generated"`
	ConflictFilesCleanedUpTotal              float64 `perflib:"Conflict Files Cleaned Up"`
	ConflictFilesGeneratedTotal              float64 `perflib:"Conflict Files Generated"`
	ConflictFolderCleanupsCompletedTotal     float64 `perflib:"Conflict folder Cleanups Completed"`
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

func (c *Collector) collectFolder(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var dst []perflibDFSRFolder
	if err := perflib.UnmarshalObject(ctx.PerfObjects["DFS Replicated Folders"], &dst, logger); err != nil {
		return err
	}

	for _, folder := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.folderBandwidthSavingsUsingDFSReplicationTotal,
			prometheus.CounterValue,
			folder.BandwidthSavingsUsingDFSReplicationTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.CompressedSizeOfFilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictBytesCleanedupTotal,
			prometheus.CounterValue,
			folder.ConflictBytesCleanedupTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictBytesGeneratedTotal,
			prometheus.CounterValue,
			folder.ConflictBytesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictFilesCleanedUpTotal,
			prometheus.CounterValue,
			folder.ConflictFilesCleanedUpTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictFilesGeneratedTotal,
			prometheus.CounterValue,
			folder.ConflictFilesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictfolderCleanupsCompletedTotal,
			prometheus.CounterValue,
			folder.ConflictFolderCleanupsCompletedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderConflictSpaceInUse,
			prometheus.GaugeValue,
			folder.ConflictSpaceInUse,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderDeletedSpaceInUse,
			prometheus.GaugeValue,
			folder.DeletedSpaceInUse,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderDeletedBytesCleanedUpTotal,
			prometheus.CounterValue,
			folder.DeletedBytesCleanedUpTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderDeletedBytesGeneratedTotal,
			prometheus.CounterValue,
			folder.DeletedBytesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderDeletedFilesCleanedUpTotal,
			prometheus.CounterValue,
			folder.DeletedFilesCleanedUpTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderDeletedFilesGeneratedTotal,
			prometheus.CounterValue,
			folder.DeletedFilesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderFileInstallsRetriedTotal,
			prometheus.CounterValue,
			folder.FileInstallsRetriedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderFileInstallsSucceededTotal,
			prometheus.CounterValue,
			folder.FileInstallsSucceededTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderFilesReceivedTotal,
			prometheus.CounterValue,
			folder.FilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderRDCBytesReceivedTotal,
			prometheus.CounterValue,
			folder.RDCBytesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderRDCCompressedSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.RDCCompressedSizeOfFilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderRDCNumberofFilesReceivedTotal,
			prometheus.CounterValue,
			folder.RDCNumberofFilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderRDCSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.RDCSizeOfFilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderSizeOfFilesReceivedTotal,
			prometheus.CounterValue,
			folder.SizeOfFilesReceivedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderStagingSpaceInUse,
			prometheus.GaugeValue,
			folder.StagingSpaceInUse,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderStagingBytesCleanedUpTotal,
			prometheus.CounterValue,
			folder.StagingBytesCleanedUpTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderStagingBytesGeneratedTotal,
			prometheus.CounterValue,
			folder.StagingBytesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderStagingFilesCleanedUpTotal,
			prometheus.CounterValue,
			folder.StagingFilesCleanedUpTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderStagingFilesGeneratedTotal,
			prometheus.CounterValue,
			folder.StagingFilesGeneratedTotal,
			folder.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.folderUpdatesDroppedTotal,
			prometheus.CounterValue,
			folder.UpdatesDroppedTotal,
			folder.Name,
		)
	}
	return nil
}

// perflibDFSRVolume Perflib: "DFS Replication Service Volumes".
type perflibDFSRVolume struct {
	Name string

	DatabaseCommitsTotal           float64 `perflib:"Database Commits"`
	DatabaseLookupsTotal           float64 `perflib:"Database Lookups"`
	USNJournalRecordsReadTotal     float64 `perflib:"USN Journal Records Read"`
	USNJournalRecordsAcceptedTotal float64 `perflib:"USN Journal Records Accepted"`
	USNJournalUnreadPercentage     float64 `perflib:"USN Journal Records Unread Percentage"`
}

func (c *Collector) collectVolume(ctx *types.ScrapeContext, logger log.Logger, ch chan<- prometheus.Metric) error {
	logger = log.With(logger, "collector", Name)
	var dst []perflibDFSRVolume
	if err := perflib.UnmarshalObject(ctx.PerfObjects["DFS Replication Service Volumes"], &dst, logger); err != nil {
		return err
	}

	for _, volume := range dst {
		ch <- prometheus.MustNewConstMetric(
			c.volumeDatabaseLookupsTotal,
			prometheus.CounterValue,
			volume.DatabaseLookupsTotal,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.volumeDatabaseCommitsTotal,
			prometheus.CounterValue,
			volume.DatabaseCommitsTotal,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.volumeUSNJournalRecordsAcceptedTotal,
			prometheus.CounterValue,
			volume.USNJournalRecordsAcceptedTotal,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.volumeUSNJournalRecordsReadTotal,
			prometheus.CounterValue,
			volume.USNJournalRecordsReadTotal,
			volume.Name,
		)

		ch <- prometheus.MustNewConstMetric(
			c.volumeUSNJournalUnreadPercentage,
			prometheus.GaugeValue,
			volume.USNJournalUnreadPercentage,
			volume.Name,
		)
	}
	return nil
}
